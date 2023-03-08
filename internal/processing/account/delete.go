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
	"fmt"
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

const deleteSelectLimit = 50

// Delete deletes an account, and all of that account's statuses, media, follows, notifications, etc etc etc.
// The origin passed here should be either the ID of the account doing the delete (can be itself), or the ID of a domain block.
func (p *Processor) Delete(ctx context.Context, account *gtsmodel.Account, origin string) gtserror.WithCode {
	l := log.WithContext(ctx).WithFields(kv.Fields{
		{"username", account.Username},
		{"domain", account.Domain},
	}...)
	l.Trace("beginning account delete process")

	if account.IsLocal() {
		if err := p.deleteUserAndTokensForAccount(ctx, account); err != nil {
			return gtserror.NewErrorInternalError(err)
		}
	}

	if err := p.deleteAccountFollows(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	if err := p.deleteAccountBlocks(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	if err := p.deleteAccountStatuses(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	if err := p.deleteAccountNotifications(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	if err := p.deleteAccountPeripheral(ctx, account); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	// To prevent the account being created again,
	// stubbify it and update it in the db.
	// The account will not be deleted, but it
	// will become completely unusable.
	columns := stubbifyAccount(account, origin)
	if err := p.state.DB.UpdateAccount(ctx, account, columns...); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	l.Info("account deleted")
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
		user, err := p.state.DB.GetUserByAccountID(ctx, account.ID)
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
		requestingAccount, err := p.state.DB.GetAccountByID(ctx, form.DeleteOriginID)
		if err != nil {
			return gtserror.NewErrorInternalError(err)
		}

		fromClientAPIMessage.OriginAccount = requestingAccount
	}

	// put the delete in the processor queue to handle the rest of it asynchronously
	p.state.Workers.EnqueueClientAPI(ctx, fromClientAPIMessage)

	return nil
}

// deleteUserAndTokensForAccount deletes the gtsmodel.User and
// any OAuth tokens and applications for the given account.
//
// Callers to this function should already have checked that
// this is a local account, or else it won't have a user associated
// with it, and this will fail.
func (p *Processor) deleteUserAndTokensForAccount(ctx context.Context, account *gtsmodel.Account) error {
	user, err := p.state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("deleteUserAndTokensForAccount: db error getting user: %w", err)
	}

	tokens := []*gtsmodel.Token{}
	if err := p.state.DB.GetWhere(ctx, []db.Where{{Key: "user_id", Value: user.ID}}, &tokens); err != nil {
		return fmt.Errorf("deleteUserAndTokensForAccount: db error getting tokens: %w", err)
	}

	for _, t := range tokens {
		// Delete any OAuth clients associated with this token.
		if err := p.state.DB.DeleteByID(ctx, t.ClientID, &[]*gtsmodel.Client{}); err != nil {
			return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting client: %w", err)
		}

		// Delete any OAuth applications associated with this token.
		if err := p.state.DB.DeleteWhere(ctx, []db.Where{{Key: "client_id", Value: t.ClientID}}, &[]*gtsmodel.Application{}); err != nil {
			return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting application: %w", err)
		}

		// Delete the token itself.
		if err := p.state.DB.DeleteByID(ctx, t.ID, t); err != nil {
			return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting token: %w", err)
		}
	}

	if err := p.state.DB.DeleteUserByID(ctx, user.ID); err != nil {
		return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting user: %w", err)
	}

	return nil
}

// deleteAccountFollows deletes:
//   - Follows targeting account.
//   - Follow requests targeting account.
//   - Follows created by account.
//   - Follow requests created by account.
func (p *Processor) deleteAccountFollows(ctx context.Context, account *gtsmodel.Account) error {
	// Delete follows targeting this account.
	followedBy, err := p.state.DB.GetFollows(ctx, "", account.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("deleteAccountFollows: db error getting follows targeting account %s: %w", account.ID, err)
	}

	for _, follow := range followedBy {
		if _, err := p.state.DB.Unfollow(ctx, follow.AccountID, account.ID); err != nil {
			return fmt.Errorf("deleteAccountFollows: db error unfollowing account followedBy: %w", err)
		}
	}

	// Delete follow requests targeting this account.
	followRequestedBy, err := p.state.DB.GetFollowRequests(ctx, "", account.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("deleteAccountFollows: db error getting follow requests targeting account %s: %w", account.ID, err)
	}

	for _, followRequest := range followRequestedBy {
		if _, err := p.state.DB.UnfollowRequest(ctx, followRequest.AccountID, account.ID); err != nil {
			return fmt.Errorf("deleteAccountFollows: db error unfollowing account followRequestedBy: %w", err)
		}
	}

	var (
		// Use this slice to batch unfollow messages.
		msgs = []messages.FromClientAPI{}
		// To avoid checking if account is local over + over
		// inside the subsequent loops, just generate static
		// side effects function once now.
		unfollowSideEffects = p.unfollowSideEffectsFunc(account)
	)

	// Delete follows originating from this account.
	following, err := p.state.DB.GetFollows(ctx, account.ID, "")
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("deleteAccountFollows: db error getting follows owned by account %s: %w", account.ID, err)
	}

	// For each follow owned by this account, unfollow
	// and process side effects (noop if remote account).
	for _, follow := range following {
		if uri, err := p.state.DB.Unfollow(ctx, account.ID, follow.TargetAccountID); err != nil {
			return fmt.Errorf("deleteAccountFollows: db error unfollowing account: %w", err)
		} else if uri == "" {
			// There was no follow after all.
			// Some race condition? Skip.
			continue
		}

		if msg := unfollowSideEffects(account, follow); msg != nil {
			// There was a side effect to process.
			msgs = append(msgs, *msg)
		}
	}

	// Delete follow requests originating from this account.
	followRequesting, err := p.state.DB.GetFollowRequests(ctx, account.ID, "")
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("deleteAccountFollows: db error getting follow requests owned by account %s: %w", account.ID, err)
	}

	// For each follow owned by this account, unfollow
	// and process side effects (noop if remote account).
	for _, followRequest := range followRequesting {
		uri, err := p.state.DB.UnfollowRequest(ctx, account.ID, followRequest.TargetAccountID)
		if err != nil {
			return fmt.Errorf("deleteAccountFollows: db error unfollowRequesting account: %w", err)
		}

		if uri == "" {
			// There was no follow request after all.
			// Some race condition? Skip.
			continue
		}

		// Dummy out a follow so our side effects func
		// has something to work with. This follow will
		// never enter the db, it's just for convenience.
		follow := &gtsmodel.Follow{
			URI:             uri,
			AccountID:       followRequest.AccountID,
			Account:         followRequest.Account,
			TargetAccountID: followRequest.TargetAccountID,
			TargetAccount:   followRequest.TargetAccount,
		}

		if msg := unfollowSideEffects(account, follow); msg != nil {
			// There was a side effect to process.
			msgs = append(msgs, *msg)
		}
	}

	// Process accreted messages asynchronously.
	p.state.Workers.EnqueueClientAPI(ctx, msgs...)

	return nil
}

func (p *Processor) unfollowSideEffectsFunc(deletedAccount *gtsmodel.Account) func(account *gtsmodel.Account, follow *gtsmodel.Follow) *messages.FromClientAPI {
	if !deletedAccount.IsLocal() {
		// Don't try to process side effects
		// for accounts that aren't local.
		return func(account *gtsmodel.Account, follow *gtsmodel.Follow) *messages.FromClientAPI {
			return nil // noop
		}
	}

	return func(account *gtsmodel.Account, follow *gtsmodel.Follow) *messages.FromClientAPI {
		if follow.TargetAccount == nil || follow.TargetAccount.IsLocal() {
			// TargetAccount seems to have gone, or is a local
			// account, so no side effects to process.
			return nil
		}

		// There was a follow, process side effects.
		return &messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel:       follow,
			OriginAccount:  account,
			TargetAccount:  follow.TargetAccount,
		}
	}
}

func (p *Processor) deleteAccountBlocks(ctx context.Context, account *gtsmodel.Account) error {
	// Delete blocks created by this account.
	if err := p.state.DB.DeleteBlocksByOriginAccountID(ctx, account.ID); err != nil {
		return fmt.Errorf("deleteAccountBlocks: db error deleting blocks created by account %s: %w", account.ID, err)
	}

	// Delete blocks targeting this account.
	if err := p.state.DB.DeleteBlocksByTargetAccountID(ctx, account.ID); err != nil {
		return fmt.Errorf("deleteAccountBlocks: db error deleting blocks targeting account %s: %w", account.ID, err)
	}

	return nil
}

// deleteAccountStatuses iterates through all statuses owned by
// the given account, passing each discovered status (and boosts
// thereof) to the processor workers for further async processing.
func (p *Processor) deleteAccountStatuses(ctx context.Context, account *gtsmodel.Account) error {
	// We'll select statuses 50 at a time so we don't wreck the db,
	// and pass them through to the client api worker to handle.
	//
	// Deleting the statuses in this way also handles deleting the
	// account's media attachments, mentions, and polls, since these
	// are all attached to statuses.

	var (
		statuses []*gtsmodel.Status
		err      error
		maxID    string
		msgs     = []messages.FromClientAPI{}
	)

	for statuses, err = p.state.DB.GetAccountStatuses(ctx, account.ID, deleteSelectLimit, false, false, maxID, "", false, false); err == nil && len(statuses) != 0; statuses, err = p.state.DB.GetAccountStatuses(ctx, account.ID, deleteSelectLimit, false, false, maxID, "", false, false) {
		// Update next maxID from last status.
		maxID = statuses[len(statuses)-1].ID

		for _, status := range statuses {
			status.Account = account // ensure account is set

			// Pass the status delete through the client api worker for processing.
			msgs = append(msgs, messages.FromClientAPI{
				APObjectType:   ap.ObjectNote,
				APActivityType: ap.ActivityDelete,
				GTSModel:       status,
				OriginAccount:  account,
				TargetAccount:  account,
			})

			// Look for any boosts of this status in DB.
			boosts, err := p.state.DB.GetStatusReblogs(ctx, status)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return fmt.Errorf("deleteAccountStatuses: error fetching status reblogs for %s: %w", status.ID, err)
			}

			for _, boost := range boosts {
				if boost.Account == nil {
					// Fetch the relevant account for this status boost.
					boostAcc, err := p.state.DB.GetAccountByID(ctx, boost.AccountID)
					if err != nil {
						if errors.Is(err, db.ErrNoEntries) {
							// We don't have an account for this boost
							// for some reason, so just skip processing.
							continue
						}
						return fmt.Errorf("deleteAccountStatuses: error fetching boosted status account for %s: %w", boost.AccountID, err)
					}

					// Set account model
					boost.Account = boostAcc
				}

				// Pass the boost delete through the client api worker for processing.
				msgs = append(msgs, messages.FromClientAPI{
					APObjectType:   ap.ActivityAnnounce,
					APActivityType: ap.ActivityUndo,
					GTSModel:       status,
					OriginAccount:  boost.Account,
					TargetAccount:  account,
				})
			}
		}
	}

	// Make sure we don't have a real error when we leave the loop.
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Batch process all accreted messages.
	p.state.Workers.EnqueueClientAPI(ctx, msgs...)

	return nil
}

func (p *Processor) deleteAccountNotifications(ctx context.Context, account *gtsmodel.Account) error {
	// Delete all notifications targeting given account.
	if err := p.state.DB.DeleteNotifications(ctx, account.ID, "", ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Delete all notifications originating from given account.
	if err := p.state.DB.DeleteNotifications(ctx, "", account.ID, ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	return nil
}

func (p *Processor) deleteAccountPeripheral(ctx context.Context, account *gtsmodel.Account) error {
	// Delete all bookmarks owned by given account.
	if err := p.state.DB.DeleteStatusBookmarks(ctx, account.ID, "", ""); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Delete all bookmarks targeting given account.
	if err := p.state.DB.DeleteStatusBookmarks(ctx, "", account.ID, ""); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Delete all faves owned by given account.
	if err := p.state.DB.DeleteStatusFaves(ctx, account.ID, "", ""); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Delete all faves targeting given account.
	if err := p.state.DB.DeleteStatusFaves(ctx, "", account.ID, ""); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// TODO: add status mutes here when they're implemented.

	return nil
}

// stubbifyAccount renders the given account as a stub,
// removing most information from it and marking it as
// suspended.
//
// The origin parameter refers to the origin of the
// suspension action; should be an account ID or domain
// block ID.
//
// For caller's convenience, this function returns the db
// names of all columns that are updated by it.
func stubbifyAccount(account *gtsmodel.Account, origin string) []string {
	var (
		falseBool = func() *bool { b := false; return &b }
		trueBool  = func() *bool { b := true; return &b }
		now       = time.Now()
		never     = time.Time{}
	)

	account.FetchedAt = never
	account.AvatarMediaAttachmentID = ""
	account.AvatarRemoteURL = ""
	account.HeaderMediaAttachmentID = ""
	account.HeaderRemoteURL = ""
	account.DisplayName = ""
	account.EmojiIDs = nil
	account.Emojis = nil
	account.Fields = nil
	account.Note = ""
	account.NoteRaw = ""
	account.Memorial = falseBool()
	account.AlsoKnownAs = ""
	account.MovedToAccountID = ""
	account.Reason = ""
	account.Discoverable = falseBool()
	account.StatusContentType = ""
	account.CustomCSS = ""
	account.SuspendedAt = now
	account.SuspensionOrigin = origin
	account.HideCollections = trueBool()
	account.EnableRSS = falseBool()

	return []string{
		"fetched_at",
		"avatar_media_attachment_id",
		"avatar_remote_url",
		"header_media_attachment_id",
		"header_remote_url",
		"display_name",
		"emojis",
		"fields",
		"note",
		"note_raw",
		"memorial",
		"also_known_as",
		"moved_to_account_id",
		"reason",
		"discoverable",
		"status_content_type",
		"custom_css",
		"suspended_at",
		"suspension_origin",
		"hide_collections",
		"enable_rss",
	}
}
