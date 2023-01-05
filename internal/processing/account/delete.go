/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package account

import (
	"context"
	"errors"
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"golang.org/x/crypto/bcrypt"
)

// Delete handles the complete deletion of an account.
//
// To be done in this function:
// 1. Delete account's application(s), clients, and oauth tokens
// 2. Delete account's blocks
// 3. Delete account's emoji
// 4. Delete account's follow requests
// 5. Delete account's follows
// 6. Delete account's statuses
// 7. Delete account's media attachments
// 8. Delete account's mentions
// 9. Delete account's polls
// 10. Delete account's notifications
// 11. Delete account's bookmarks
// 12. Delete account's faves
// 13. Delete account's mutes
// 14. Delete account's streams
// 15. Delete account's tags
// 16. Delete account's user
// 17. Delete account's timeline
// 18. Delete account itself
func (p *processor) Delete(ctx context.Context, account *gtsmodel.Account, origin string) gtserror.WithCode {
	fields := kv.Fields{{"username", account.Username}}

	if account.Domain != "" {
		fields = append(fields, kv.Field{
			"domain", account.Domain,
		})
	}

	l := log.WithFields(fields...)
	l.Trace("beginning account delete process")

	// 1. Delete account's application(s), clients, and oauth tokens
	// we only need to do this step for local account since remote ones won't have any tokens or applications on our server
	var user *gtsmodel.User
	if account.Domain == "" {
		// see if we can get a user for this account
		var err error
		if user, err = p.db.GetUserByAccountID(ctx, account.ID); err == nil {
			// we got one! select all tokens with the user's ID
			tokens := []*gtsmodel.Token{}
			if err := p.db.GetWhere(ctx, []db.Where{{Key: "user_id", Value: user.ID}}, &tokens); err == nil {
				// we have some tokens to delete
				for _, t := range tokens {
					// delete client(s) associated with this token
					if err := p.db.DeleteByID(ctx, t.ClientID, &gtsmodel.Client{}); err != nil {
						l.Errorf("error deleting oauth client: %s", err)
					}
					// delete application(s) associated with this token
					if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "client_id", Value: t.ClientID}}, &gtsmodel.Application{}); err != nil {
						l.Errorf("error deleting application: %s", err)
					}
					// delete the token itself
					if err := p.db.DeleteByID(ctx, t.ID, t); err != nil {
						l.Errorf("error deleting oauth token: %s", err)
					}
				}
			}
		}
	}

	// 2. Delete account's blocks
	l.Trace("deleting account blocks")
	// first delete any blocks that this account created
	if err := p.db.DeleteBlocksByOriginAccountID(ctx, account.ID); err != nil {
		l.Errorf("error deleting blocks created by account: %s", err)
	}

	// now delete any blocks that target this account
	if err := p.db.DeleteBlocksByTargetAccountID(ctx, account.ID); err != nil {
		l.Errorf("error deleting blocks targeting account: %s", err)
	}

	// 3. Delete account's emoji
	// nothing to do here

	// 4. Delete account's follow requests
	// TODO: federate these if necessary
	l.Trace("deleting account follow requests")
	// first delete any follow requests that this account created
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		l.Errorf("error deleting follow requests created by account: %s", err)
	}

	// now delete any follow requests that target this account
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		l.Errorf("error deleting follow requests targeting account: %s", err)
	}

	// 5. Delete account's follows
	// TODO: federate these if necessary
	l.Trace("deleting account follows")
	// first delete any follows that this account created
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		l.Errorf("error deleting follows created by account: %s", err)
	}

	// now delete any follows that target this account
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		l.Errorf("error deleting follows targeting account: %s", err)
	}

	var maxID string

	// 6. Delete account's statuses
	l.Trace("deleting account statuses")

	// we'll select statuses 20 at a time so we don't wreck the db, and pass them through to the client api channel
	// Deleting the statuses in this way also handles 7. Delete account's media attachments, 8. Delete account's mentions, and 9. Delete account's polls,
	// since these are all attached to statuses.

	for {
		// Fetch next block of account statuses from database
		statuses, err := p.db.GetAccountStatuses(ctx, account.ID, 20, false, false, maxID, "", false, false, false)
		if err != nil {
			if !errors.Is(err, db.ErrNoEntries) {
				// an actual error has occurred
				l.Errorf("Delete: db error selecting statuses for account %s: %s", account.Username, err)
			}
			break
		}

		for _, status := range statuses {
			// Ensure account is set
			status.Account = account

			l.Tracef("queue client API status delete: %s", status.ID)

			// pass the status delete through the client api channel for processing
			p.clientWorker.Queue(messages.FromClientAPI{
				APObjectType:   ap.ObjectNote,
				APActivityType: ap.ActivityDelete,
				GTSModel:       status,
				OriginAccount:  account,
				TargetAccount:  account,
			})

			// Look for any boosts of this status in DB
			boosts, err := p.db.GetStatusReblogs(ctx, status)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				l.Errorf("error fetching status reblogs for %q: %v", status.ID, err)
				continue
			}

			for _, boost := range boosts {
				if boost.Account == nil {
					// Fetch the relevant account for this status boost
					boostAcc, err := p.db.GetAccountByID(ctx, boost.AccountID)
					if err != nil {
						l.Errorf("error fetching boosted status account for %q: %v", boost.AccountID, err)
						continue
					}

					// Set account model
					boost.Account = boostAcc
				}

				l.Tracef("queue client API boost delete: %s", status.ID)

				// pass the boost delete through the client api channel for processing
				p.clientWorker.Queue(messages.FromClientAPI{
					APObjectType:   ap.ActivityAnnounce,
					APActivityType: ap.ActivityUndo,
					GTSModel:       status,
					OriginAccount:  boost.Account,
					TargetAccount:  account,
				})
			}
		}

		// Update next maxID from last status
		maxID = statuses[len(statuses)-1].ID
	}

	// 10. Delete account's notifications
	l.Trace("deleting account notifications")
	// first notifications created by account
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "origin_account_id", Value: account.ID}}, &[]*gtsmodel.Notification{}); err != nil {
		l.Errorf("error deleting notifications created by account: %s", err)
	}

	// now notifications targeting account
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.Notification{}); err != nil {
		l.Errorf("error deleting notifications targeting account: %s", err)
	}

	// 11. Delete account's bookmarks
	l.Trace("deleting account bookmarks")
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.StatusBookmark{}); err != nil {
		l.Errorf("error deleting bookmarks created by account: %s", err)
	}

	// 12. Delete account's faves
	// TODO: federate these if necessary
	l.Trace("deleting account faves")
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.StatusFave{}); err != nil {
		l.Errorf("error deleting faves created by account: %s", err)
	}

	// 13. Delete account's mutes
	l.Trace("deleting account mutes")
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.StatusMute{}); err != nil {
		l.Errorf("error deleting status mutes created by account: %s", err)
	}

	// 14. Delete account's streams
	// TODO

	// 15. Delete account's tags
	// TODO

	// 16. Delete account's user
	if user != nil {
		l.Trace("deleting account user")
		if err := p.db.DeleteUserByID(ctx, user.ID); err != nil {
			return gtserror.NewErrorInternalError(err)
		}
	}

	// 17. Delete account's timeline
	// TODO

	// 18. Delete account itself
	// to prevent the account being created again, set all these fields and update it in the db
	// the account won't actually be *removed* from the database but it will be set to just a stub
	account.Note = ""
	account.DisplayName = ""
	account.AvatarMediaAttachmentID = ""
	account.AvatarRemoteURL = ""
	account.HeaderMediaAttachmentID = ""
	account.HeaderRemoteURL = ""
	account.Reason = ""
	account.Emojis = []*gtsmodel.Emoji{}
	account.EmojiIDs = []string{}
	account.Fields = []gtsmodel.Field{}
	hideCollections := true
	account.HideCollections = &hideCollections
	discoverable := false
	account.Discoverable = &discoverable
	account.SuspendedAt = time.Now()
	account.SuspensionOrigin = origin
	err := p.db.UpdateAccount(ctx, account)
	if err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	l.Infof("deleted account with username %s from domain %s", account.Username, account.Domain)
	return nil
}

func (p *processor) DeleteLocal(ctx context.Context, account *gtsmodel.Account, form *apimodel.AccountDeleteRequest) gtserror.WithCode {
	fromClientAPIMessage := messages.FromClientAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityDelete,
		TargetAccount:  account,
	}

	if form.DeleteOriginID == account.ID {
		// the account owner themself has requested deletion via the API, get their user from the db
		user, err := p.db.GetUserByAccountID(ctx, account.ID)
		if err != nil {
			return gtserror.NewErrorInternalError(err)
		}

		// now check that the password they supplied is correct
		// make sure a password is actually set and bail if not
		if user.EncryptedPassword == "" {
			return gtserror.NewErrorForbidden(errors.New("user password was not set"))
		}

		// compare the provided password with the encrypted one from the db, bail if they don't match
		if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(form.Password)); err != nil {
			return gtserror.NewErrorForbidden(errors.New("invalid password"))
		}

		fromClientAPIMessage.OriginAccount = account
	} else {
		// the delete has been requested by some other account, grab it;
		// if we've reached this point we know it has permission already
		requestingAccount, err := p.db.GetAccountByID(ctx, form.DeleteOriginID)
		if err != nil {
			return gtserror.NewErrorInternalError(err)
		}

		fromClientAPIMessage.OriginAccount = requestingAccount
	}

	// put the delete in the processor queue to handle the rest of it asynchronously
	p.clientWorker.Queue(fromClientAPIMessage)

	return nil
}
