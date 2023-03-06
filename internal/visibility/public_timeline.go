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
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// StatusPublictimelineable returns true if targetStatus should be in the public timeline of the requesting account.
//
// This function will call StatusVisible internally, so it's not necessary to call it beforehand.
func (f *Filter) StatusPublicTimelineable(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	var requesterID string

	if requester != nil {
		// Use provided account ID.
		requesterID = requester.ID
	} else {
		// Set a no-auth ID flag.
		requesterID = "noauth"
	}

	visibility, err := f.state.Caches.Visibility.Load("Type.RequesterID.ItemID", func() (*cache.CachedVisibility, error) {
		// Visibility not yet cached, perform timeline visibility lookup.
		visible, err := f.isStatusPublicTimelineable(ctx, requester, status)
		if err != nil {
			return nil, err
		}

		// Return visibility value.
		return &cache.CachedVisibility{
			ItemID:      status.ID,
			RequesterID: requesterID,
			Type:        "public",
			Value:       visible,
		}, nil
	}, "public", requesterID, status.ID)
	if err != nil {
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

	if status.InReplyToID != "" {
		// This is a reply.

		// Don't show replies not coming from original author
		// (i.e. we only show singular author status threads).
		if status.InReplyToAccountID != status.AccountID {
			return false, nil
		}

		// Get reply's parent.
		parent := status.InReplyTo

		if parent == nil {
			var err error

			// Parent of current status needs fetching from the database.
			parent, err = f.state.DB.GetStatusByID(ctx, status.InReplyToID)
			if err != nil {
				return false, fmt.Errorf("isStatusPublicTimelineable: error getting status %s: %w", status.InReplyToID, err)
			}
		}

		// Check the public timelineable-ness (?) of parent status to requester.
		visible, err := f.StatusPublicTimelineable(ctx, requester, parent)
		if err != nil {
			return false, err
		}

		if !visible {
			log.Trace(ctx, "status parent not visible")
			return false, nil
		}
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

	return true, nil
}
