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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/cache"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// StatusHomeTimelineable checks if given status should be included on requester's public timeline. Primarily relying on status visibility to requester and the AP visibility setting, and ignoring conversation threads.
func (f *Filter) StatusPublicTimelineable(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	const vtype = cache.VisibilityTypePublic

	// By default we assume no auth.
	requesterID := NoAuth

	if requester != nil {
		// Use provided account ID.
		requesterID = requester.ID
	}

	visibility, err := f.state.Caches.Visibility.LoadOne("Type,RequesterID,ItemID", func() (*cache.CachedVisibility, error) {
		// Visibility not yet cached, perform timeline visibility lookup.
		visible, err := f.isStatusPublicTimelineable(ctx, requester, status)
		if err != nil {
			return nil, err
		}

		// Return visibility value.
		return &cache.CachedVisibility{
			ItemID:      status.ID,
			RequesterID: requesterID,
			Type:        vtype,
			Value:       visible,
		}, nil
	}, vtype, requesterID, status.ID)
	if err != nil {
		if err == cache.SentinelError {
			// Filter-out our temporary
			// race-condition error.
			return false, nil
		}

		return false, err
	}

	return visibility.Value, nil
}

func (f *Filter) isStatusPublicTimelineable(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	if status.CreatedAt.After(time.Now().Add(24 * time.Hour)) {
		// Statuses made over 1 day in the future we don't show...
		log.Warnf(ctx, "status >24hrs in the future: %+v", status)
		return false, nil
	}

	// Don't show boosts on timeline.
	if status.BoostOfID != "" {
		return false, nil
	}

	// Check whether status is visible to requesting account.
	visible, err := f.StatusVisible(ctx, requester, status)
	if err != nil {
		return false, err
	}

	if !visible {
		log.Trace(ctx, "status not visible to timeline requester")
		return false, nil
	}

	for parent := status; parent.InReplyToURI != ""; {
		// Fetch next parent to lookup.
		parentID := parent.InReplyToID
		if parentID == "" {
			log.Debugf(ctx, "status not (yet) deref'd: %s", parent.InReplyToURI)
			return false, cache.SentinelError
		}

		// Get the next parent in the chain from DB.
		parent, err = f.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			parentID,
		)
		if err != nil {
			return false, gtserror.Newf("error getting status parent %s: %w", parentID, err)
		}

		if parent.AccountID != status.AccountID {
			// This is not a single author reply-chain-thread,
			// instead is an actualy conversation. Don't timeline.
			log.Trace(ctx, "ignoring multi-author reply-chain")
			return false, nil
		}
	}

	// This is either a visible status in a
	// single-author thread, or a visible top
	// level status. Show on public timeline.
	return true, nil
}
