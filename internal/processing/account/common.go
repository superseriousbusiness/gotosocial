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
	user, err := p.db.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("deleteUserAndTokensForAccount: db error getting user: %w", err)
	}

	tokens := []*gtsmodel.Token{}
	if err := p.db.GetWhere(ctx, []db.Where{{Key: "user_id", Value: user.ID}}, &tokens); err != nil {
		return fmt.Errorf("deleteUserAndTokensForAccount: db error getting tokens: %w", err)
	}

	for _, t := range tokens {
		// Delete any OAuth clients associated with this token.
		if err := p.db.DeleteByID(ctx, t.ClientID, &[]*gtsmodel.Client{}); err != nil {
			return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting client: %w", err)
		}

		// Delete any OAuth applications associated with this token.
		if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "client_id", Value: t.ClientID}}, &[]*gtsmodel.Application{}); err != nil {
			return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting application: %w", err)
		}

		// Delete the token itself.
		if err := p.db.DeleteByID(ctx, t.ID, t); err != nil {
			return fmt.Errorf("deleteUserAndTokensForAccount: db error deleting token: %w", err)
		}
	}

	if err := p.db.DeleteUserByID(ctx, user.ID); err != nil {
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
	if err := p.db.DeleteBlocksByOriginAccountID(ctx, account.ID); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting blocks created by account %s: %w", account.ID, err)
	}

	// Delete blocks targeting this account.
	if err := p.db.DeleteBlocksByTargetAccountID(ctx, account.ID); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting blocks targeting account %s: %w", account.ID, err)
	}

	// Delete follow requests created by this account.
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests created by account %s: %w", account.ID, err)
	}

	// Delete follow requests targeting this account.
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests targeting account %s: %w", account.ID, err)
	}

	// Delete follows created by this account.
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		return fmt.Errorf("deleteRelationshipsForAccount: db error deleting follow requests created by account %s: %w", account.ID, err)
	}

	// Delete follows targeting this account.
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
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

	for statuses, err = p.db.GetAccountStatuses(ctx, account.ID, deleteSelectLimit, false, false, maxID, "", false, false); err == nil && len(statuses) != 0; statuses, err = p.db.GetAccountStatuses(ctx, account.ID, deleteSelectLimit, false, false, maxID, "", false, false) {
		// Update next maxID from last status.
		maxID = statuses[len(statuses)-1].ID

		for _, status := range statuses {
			status.Account = account // ensure account is set

			// Pass the status delete through the client api channel for processing
			p.clientWorker.Queue(messages.FromClientAPI{
				APObjectType:   ap.ObjectNote,
				APActivityType: ap.ActivityDelete,
				GTSModel:       status,
				OriginAccount:  account,
				TargetAccount:  account,
			})

			// Look for any boosts of this status in DB.
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

				// Pass the boost delete through the client api channel for processing
				p.clientWorker.Queue(messages.FromClientAPI{
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
