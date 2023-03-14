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
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// StatusHomeTimelineable checks if given status should be included on owner's home timeline. Primarily relying on status visibility to owner and the AP visibility setting, but also taking into account thread replies etc.
func (f *Filter) StatusHomeTimelineable(ctx context.Context, owner *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	// By default we assume no auth.
	requesterID := noauth

	if owner != nil {
		// Use provided account ID.
		requesterID = owner.ID
	}

	visibility, err := f.state.Caches.Visibility.Load("Type.RequesterID.ItemID", func() (*cache.CachedVisibility, error) {
		// Visibility not yet cached, perform timeline visibility lookup.
		visible, err := f.isStatusHomeTimelineable(ctx, owner, status)
		if err != nil {
			return nil, err
		}

		// Return visibility value.
		return &cache.CachedVisibility{
			ItemID:      status.ID,
			RequesterID: requesterID,
			Type:        "home",
			Value:       visible,
		}, nil
	}, "home", requesterID, status.ID)
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

func (f *Filter) isStatusHomeTimelineable(ctx context.Context, owner *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	if status.CreatedAt.After(time.Now().Add(24 * time.Hour)) {
		// Statuses made over 1 day in the future we don't show...
		log.Warnf(ctx, "status >24hrs in the future: %+v", status)
		return false, nil
	}

	// Check whether status is visible to timeline owner.
	visible, err := f.StatusVisible(ctx, owner, status)
	if err != nil {
		return false, err
	}

	if !visible {
		log.Trace(ctx, "status not visible to timeline owner")
		return false, nil
	}

	if status.AccountID == owner.ID {
		// Author can always see their status.
		return true, nil
	}

	if status.MentionsAccount(owner.ID) {
		// Can always see when you are mentioned.
		return true, nil
	}

	var (
		parent    *gtsmodel.Status
		included  bool
		oneAuthor bool
	)

	for parent = status; parent.InReplyToURI != ""; {
		// Fetch next parent to lookup.
		parentID := parent.InReplyToID
		if parentID == "" {
			log.Warnf(ctx, "status not yet deref'd: %s", parent.InReplyToURI)
			return false, cache.SentinelError
		}

		// Get the next parent in the chain from DB.
		parent, err = f.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			parentID,
		)
		if err != nil {
			return false, fmt.Errorf("isStatusHomeTimelineable: error getting status parent %s: %w", parentID, err)
		}

		if (parent.AccountID == owner.ID) ||
			parent.MentionsAccount(owner.ID) {
			// Owner is in / mentioned in
			// this status thread.
			included = true
			break
		}

		if oneAuthor {
			// Check if this is a single-author status thread.
			oneAuthor = (parent.AccountID == status.AccountID)
		}
	}

	if parent != status && !included && !oneAuthor {
		log.Trace(ctx, "ignoring visible reply to conversation thread excluding owner")
		return false, nil
	}

	// At this point status is either a top-level status, a reply in a single
	// author thread (e.g. "this is my weird-ass take and here is why 1/10 🧵"),
	// or a thread mentioning / including timeline owner.

	if status.Visibility == gtsmodel.VisibilityFollowersOnly ||
		status.Visibility == gtsmodel.VisibilityMutualsOnly {
		// Followers/mutuals only post that already passed the status
		// visibility check, (i.e. we follow / mutuals with author).
		return true, nil
	}

	// Ensure owner follows author of public/unlocked status.
	follow, err := f.state.DB.IsFollowing(ctx,
		owner.ID,
		status.AccountID,
	)
	if err != nil {
		return false, fmt.Errorf("isStatusHomeTimelineable: error checking follow %s->%s: %w", owner.ID, status.AccountID, err)
	}

	if !follow {
		log.Trace(ctx, "ignoring visible status from unfollowed author")
		return false, nil
	}

	return true, nil
}
