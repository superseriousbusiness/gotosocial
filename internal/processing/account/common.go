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

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

const deleteSelectLimit = 50

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

// deleteRelationshipsForAccount deletes:
//   - Blocks created by or targeting account.
//   - Follow requests created by or targeting account.
//   - Follows created by or targeting account.
func (p *Processor) deleteRelationshipsForAccount(ctx context.Context, account *gtsmodel.Account) error {
	// Delete blocks created by this account.
	if err := p.state.DB.DeleteBlocksByOriginAccountID(ctx, account.ID); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting blocks created by account %s: %w", account.ID, err)
	}

	// Delete blocks targeting this account.
	if err := p.state.DB.DeleteBlocksByTargetAccountID(ctx, account.ID); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting blocks targeting account %s: %w", account.ID, err)
	}

	// Delete follow requests created by this account.
	if err := p.state.DB.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests created by account %s: %w", account.ID, err)
	}

	// Delete follow requests targeting this account.
	if err := p.state.DB.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests targeting account %s: %w", account.ID, err)
	}

	// Delete follows created by this account.
	if err := p.state.DB.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests created by account %s: %w", account.ID, err)
	}

	// Delete follows targeting this account.
	if err := p.state.DB.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests targeting account %s: %w", account.ID, err)
	}

	return nil
}

// deleteAccountStatuses iterates through all statuses owned by
// the given account, passing each discovered status
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
	)

	for statuses, err = p.state.DB.GetAccountStatuses(ctx, account.ID, deleteSelectLimit, false, false, maxID, "", false, false); err == nil && len(statuses) != 0; statuses, err = p.state.DB.GetAccountStatuses(ctx, account.ID, deleteSelectLimit, false, false, maxID, "", false, false) {
		// Update next maxID from last status.
		maxID = statuses[len(statuses)-1].ID

		for _, status := range statuses {
			status.Account = account // ensure account is set

			// Pass the status delete through the client api worker for processing.
			p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
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
					// Fetch the relevant account for this status boost
					boostAcc, err := p.state.DB.GetAccountByID(ctx, boost.AccountID)
					if err != nil {
						return fmt.Errorf("deleteAccountStatuses: error fetching boosted status account for %s: %w", boost.AccountID, err)
					}

					// Set account model
					boost.Account = boostAcc
				}

				// Pass the boost delete through the client api worker for processing.
				p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
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

	return nil
}

func (p *Processor) deleteAccountNotifications(ctx context.Context, account *gtsmodel.Account) error {
	// Delete all notifications targeting given account.
	if err := p.state.DB.DeleteNotifications(ctx, account.ID, ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Delete all notifications originating from given account.
	if err := p.state.DB.DeleteNotifications(ctx, "", account.ID); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

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
