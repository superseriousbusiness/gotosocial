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

package visibility

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// StatusHometimelineable returns true if targetStatus should be in the home timeline of the requesting account.
//
// This function will call StatusVisible internally, so it's not necessary to call it beforehand.
func (f *Filter) StatusHomeTimelineable(ctx context.Context, owner *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	visibility, err := f.state.Caches.Visibility.Load("Type.RequesterID.ItemID", func() (*cache.CachedVisibility, error) {
		// Visibility not yet cached, perform timeline visibility lookup.
		visible, err := f.isStatusHomeTimelineable(ctx, owner, status)
		if err != nil {
			return nil, err
		}

		var ownerID string

		if owner != nil {
			// Use provided account ID.
			ownerID = owner.ID
		} else {
			// Set a no-auth ID flag.
			ownerID = "noauth"
		}

		// Return visibility value.
		return &cache.CachedVisibility{
			ItemID:      status.ID,
			RequesterID: ownerID,
			Type:        "home",
			Value:       visible,
		}, nil
	}, "home", owner.ID, status.ID)
	if err != nil {
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
		// Status author can always see their status.
		return true, nil
	}

	if status.InReplyToID != "" {
		var oldest *gtsmodel.Status

		// Iteratively get to the oldest status in the reply-chain.
		for oldest = status.InReplyTo; oldest.InReplyToID != ""; {
			oldest, err = f.state.DB.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				oldest.InReplyToID,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return false, fmt.Errorf("isStatusHomeTimelineable: error getting status %s parent: %w", oldest.InReplyToID, err)
			}
		}

		if oldest != status {
			// Check whether owner can see the oldest parent in reply chain.
			// (this prevents conversation snippets on the home timeline).
			visible, err := f.StatusHomeTimelineable(ctx, owner, oldest)
			if err != nil {
				return false, fmt.Errorf("isStatusHomeTimelineable: error checking grandest parent %s: %w", oldest.ID, err)
			}

			if !visible {
				log.Trace(ctx, "ignoring visible reply to invisible grandest parent")
				return false, nil
			}
		}
	}

	if status.Visibility == gtsmodel.VisibilityFollowersOnly ||
		status.Visibility == gtsmodel.VisibilityMutualsOnly {
		// Followers/mutuals only post that already passed the status
		// visibility check, (i.e. we follow / mutuals with author).
		return true, nil
	}

	// Ensure owner follows the status author.
	follow, err := f.state.DB.IsFollowing(ctx,
		owner,
		status.Account,
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
