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
	const vtype = cache.VisibilityTypeHome

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
		next      *gtsmodel.Status
		oneAuthor = true // Assume one author until proven otherwise.
		included  bool
		converstn bool
	)

	for next = status; next.InReplyToURI != ""; {
		// Fetch next parent to lookup.
		parentID := next.InReplyToID
		if parentID == "" {
			log.Warnf(ctx, "status not yet deref'd: %s", next.InReplyToURI)
			return false, cache.SentinelError
		}

		// Get the next parent in the chain from DB.
		next, err = f.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			parentID,
		)
		if err != nil {
			return false, fmt.Errorf("isStatusHomeTimelineable: error getting status parent %s: %w", parentID, err)
		}

		// Populate account mention objects before account mention checks.
		next.Mentions, err = f.state.DB.GetMentions(ctx, next.MentionIDs)
		if err != nil {
			return false, fmt.Errorf("isStatusHomeTimelineable: error populating status parent %s mentions: %w", parentID, err)
		}

		if (next.AccountID == owner.ID) ||
			next.MentionsAccount(owner.ID) {
			// Owner is in / mentioned in
			// this status thread. They can
			// see all future visible statuses.
			included = true
			break
		}

		// Check whether this should be a visible conversation, i.e.
		// is it between accounts on owner timeline that they follow?
		converstn, err = f.isVisibleConversation(ctx, owner, next)
		if err != nil {
			return false, fmt.Errorf("isStatusHomeTimelineable: error checking conversation visibility: %w", err)
		}

		if converstn {
			// Owner is relevant to this conversation,
			// i.e. between follows / mutuals they know.
			break
		}

		if oneAuthor {
			// Check if this continues to be a single-author thread.
			oneAuthor = (next.AccountID == status.AccountID)
		}
	}

	if next != status && !oneAuthor && !included && !converstn {
		log.Trace(ctx, "ignoring visible reply in conversation irrelevant to owner")
		return false, nil
	}

	// At this point status is either a top-level status, a reply in a single
	// author thread (e.g. "this is my weird-ass take and here is why 1/10 🧵"),
	// a status thread *including* the owner, or a conversation thread between
	// accounts the timeline owner follows.

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

func (f *Filter) isVisibleConversation(ctx context.Context, owner *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	// Check if status is visible to the timeline owner.
	visible, err := f.StatusVisible(ctx, owner, status)
	if err != nil {
		return false, err
	}

	if !visible {
		// Invisible to
		// timeline owner.
		return false, nil
	}

	if status.Visibility == gtsmodel.VisibilityUnlocked ||
		status.Visibility == gtsmodel.VisibilityPublic {
		// NOTE: there is no need to check in the case of
		// direct / follow-only / mutual-only visibility statuses
		// as the above visibility check already handles this.

		// Check if owner follows the status author.
		followAuthor, err := f.state.DB.IsFollowing(ctx,
			owner.ID,
			status.AccountID,
		)
		if err != nil {
			return false, fmt.Errorf("error checking follow %s->%s: %w", owner.ID, status.AccountID, err)
		}

		if !followAuthor {
			// Not a visible status
			// in conversation thread.
			return false, nil
		}
	}

	for _, mention := range status.Mentions {
		// Check if timeline owner follows target.
		follow, err := f.state.DB.IsFollowing(ctx,
			owner.ID,
			mention.TargetAccountID,
		)
		if err != nil {
			return false, fmt.Errorf("error checking mention follow %s->%s: %w", owner.ID, mention.TargetAccountID, err)
		}

		if follow {
			// Confirmed conversation.
			return true, nil
		}
	}

	return false, nil
}
