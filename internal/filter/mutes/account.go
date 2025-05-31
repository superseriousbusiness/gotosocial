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

package mutes

import (
	"context"
	"errors"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// NOTE:
// we don't bother using the Mutes cache for any
// of the accounts functions below, as there's only
// a single cache load required of any UserMute.

// AccountMuted returns whether given target account is muted by requester.
func (f *Filter) AccountMuted(ctx context.Context, requester *gtsmodel.Account, account *gtsmodel.Account) (bool, error) {
	mute, expired, err := f.getUserMute(ctx, requester, account)
	if err != nil {
		return false, err
	} else if mute == nil {
		return false, nil
	}
	return !expired, nil
}

// AccountNotificationsMuted returns whether notifications are muted for requester when incoming from given target account.
func (f *Filter) AccountNotificationsMuted(ctx context.Context, requester *gtsmodel.Account, account *gtsmodel.Account) (bool, error) {
	mute, expired, err := f.getUserMute(ctx, requester, account)
	if err != nil {
		return false, err
	} else if mute == nil {
		return false, nil
	}
	return *mute.Notifications && !expired, nil
}

func (f *Filter) getUserMute(ctx context.Context, requester *gtsmodel.Account, account *gtsmodel.Account) (*gtsmodel.UserMute, bool, error) {
	if requester == nil {
		// Un-authed so no account
		// is possible to be muted.
		return nil, false, nil
	}

	// Look for mute against target.
	mute, err := f.state.DB.GetMute(
		gtscontext.SetBarebones(ctx),
		requester.ID,
		account.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, false, gtserror.Newf("db error getting user mute: %w", err)
	}

	if mute == nil {
		// No user mute exists!
		return nil, false, nil
	}

	// Get current time.
	now := time.Now()

	// Return whether mute is expired.
	return mute, mute.Expired(now), nil
}
