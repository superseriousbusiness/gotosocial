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

// Delete deletes an account, and all of that account's statuses, media, follows, notifications, etc etc etc.
// The origin passed here should be either the ID of the account doing the delete (can be itself), or the ID of a domain block.
func (p *Processor) Delete(ctx context.Context, account *gtsmodel.Account, origin string) gtserror.WithCode {
	l := log.WithContext(ctx).WithFields(kv.Fields{{"username", account.Username}}...)
	l.Trace("beginning account delete process")

	if account.IsLocal() {
		if err := p.deleteUserAndTokensForAccount(ctx, account); err != nil {
			return gtserror.NewErrorInternalError(err)
		}
	}

	if err := p.deleteRelationshipsForAccount(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	if err := p.deleteAccountStatuses(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
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

// DeleteLocal is like Delete, but specifically for deletion of local accounts rather than federated ones.
// Unlike Delete, it will propagate the deletion out across the federating API to other instances.
func (p *Processor) DeleteLocal(ctx context.Context, account *gtsmodel.Account, form *apimodel.AccountDeleteRequest) gtserror.WithCode {
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
