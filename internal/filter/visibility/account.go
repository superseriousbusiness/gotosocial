// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package visibility

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/cache"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// AccountVisible will check if given account is visible to requester, accounting for requester with no auth (i.e is nil), suspensions, disabled local users and account blocks.
func (f *Filter) AccountVisible(ctx context.Context, requester *gtsmodel.Account, account *gtsmodel.Account) (bool, error) {
	const vtype = cache.VisibilityTypeAccount

	// By default we assume no auth.
	requesterID := NoAuth

	if requester != nil {
		// Use provided account ID.
		requesterID = requester.ID
	}

	visibility, err := f.state.Caches.Visibility.LoadOne("Type,RequesterID,ItemID", func() (*cache.CachedVisibility, error) {
		// Visibility not yet cached, perform visibility lookup.
		visible, err := f.isAccountVisibleTo(ctx, requester, account)
		if err != nil {
			return nil, err
		}

		// Return visibility value.
		return &cache.CachedVisibility{
			ItemID:      account.ID,
			RequesterID: requesterID,
			Type:        vtype,
			Value:       visible,
		}, nil
	}, vtype, requesterID, account.ID)
	if err != nil {
		return false, err
	}

	return visibility.Value, nil
}

// isAccountVisibleTo will check if account is visible to requester. It is the "meat" of the logic to Filter{}.AccountVisible() which is called within cache loader callback.
func (f *Filter) isAccountVisibleTo(ctx context.Context, requester *gtsmodel.Account, account *gtsmodel.Account) (bool, error) {
	// Check whether target account is visible to anyone.
	visible, err := f.isAccountVisible(ctx, account)
	if err != nil {
		return false, gtserror.Newf("error checking account %s visibility: %w", account.ID, err)
	}

	if !visible {
		log.Trace(ctx, "target account is not visible to anyone")
		return false, nil
	}

	if requester == nil {
		// It seems stupid, but when un-authed all accounts are
		// visible to allow for federation to work correctly.
		return true, nil
	}

	// If requester is not visible, they cannot *see* either.
	visible, err = f.isAccountVisible(ctx, requester)
	if err != nil {
		return false, gtserror.Newf("error checking account %s visibility: %w", account.ID, err)
	}

	if !visible {
		log.Trace(ctx, "requesting account cannot see other accounts")
		return false, nil
	}

	// Check whether either blocks the other.
	blocked, err := f.state.DB.IsEitherBlocked(ctx,
		requester.ID,
		account.ID,
	)
	if err != nil {
		return false, gtserror.Newf("error checking account blocks: %w", err)
	}

	if blocked {
		log.Trace(ctx, "block exists between accounts")
		return false, nil
	}

	return true, nil
}

// isAccountVisible will check if given account should be visible at all, e.g. it may not be if suspended or disabled.
func (f *Filter) isAccountVisible(ctx context.Context, account *gtsmodel.Account) (bool, error) {
	if account.IsLocal() {
		// This is a local account.

		if account.Username == config.GetHost() {
			// This is the instance actor account.
			return true, nil
		}

		// Fetch the local user model for this account.
		user, err := f.state.DB.GetUserByAccountID(ctx, account.ID)
		if err != nil {
			err := gtserror.Newf("db error getting user for account %s: %w", account.ID, err)
			return false, err
		}

		// Make sure that user is active (i.e. not disabled, not approved etc).
		if *user.Disabled || !*user.Approved || user.ConfirmedAt.IsZero() {
			log.Trace(ctx, "local account not active")
			return false, nil
		}
	} else {
		// This is a remote account.

		// Check whether remote account's domain is blocked.
		blocked, err := f.state.DB.IsDomainBlocked(ctx, account.Domain)
		if err != nil {
			return false, err
		}

		if blocked {
			log.Trace(ctx, "remote account domain blocked")
			return false, nil
		}
	}

	if !account.SuspendedAt.IsZero() {
		log.Trace(ctx, "account suspended")
		return false, nil
	}

	return true, nil
}
