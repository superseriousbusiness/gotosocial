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
	"errors"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/cache"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// StatusHomeTimelineable checks if given status should be included on owner's home timeline. Primarily relying on status visibility to owner and the AP visibility setting, but also taking into account thread replies etc.
// Despite the name, statuses that ultimately end up in exclusive lists also need to be home-timelineable.
func (f *Filter) StatusHomeTimelineable(ctx context.Context, owner *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	const vtype = cache.VisibilityTypeHome

	// By default we assume no auth.
	requesterID := NoAuth

	if owner != nil {
		// Use provided account ID.
		requesterID = owner.ID
	}

	visibility, err := f.state.Caches.Visibility.LoadOne("Type,RequesterID,ItemID", func() (*cache.CachedVisibility, error) {
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
		// iterated-over
		// loop status.
		next = status

		// assume one author
		// until proven otherwise.
		oneAuthor = true
	)

	for {
		// Populate account mention objects before account mention checks.
		next.Mentions, err = f.state.DB.GetMentions(ctx, next.MentionIDs)
		if err != nil {
			return false, gtserror.Newf("error populating status %s mentions: %w", next.ID, err)
		}

		if (next.AccountID == owner.ID) ||
			next.MentionsAccount(owner.ID) {
			// Owner is in / mentioned in
			// this status thread. They can
			// see future visible statuses.
			visible = true
			break
		}

		var notVisible bool

		// Check whether status in conversation is explicitly relevant to timeline
		// owner (i.e. includes mutals), or is explicitly invisible (i.e. blocked).
		visible, notVisible, err = f.isVisibleConversation(ctx, owner, next)
		if err != nil {
			return false, gtserror.Newf("error checking conversation visibility: %w", err)
		}

		if notVisible {
			log.Tracef(ctx, "conversation not visible to timeline owner")
			return false, nil
		}

		if visible {
			// Conversation relevant
			// to timeline owner!
			break
		}

		if oneAuthor {
			// Check if this continues to be a single-author thread.
			oneAuthor = (next.AccountID == status.AccountID)
		}

		if next.InReplyToURI == "" {
			// Reached the top of the thread.
			break
		}

		// Check parent is deref'd.
		if next.InReplyToID == "" {
			log.Debugf(ctx, "status not (yet) deref'd: %s", next.InReplyToURI)
			return false, cache.SentinelError
		}

		// Check if parent is set.
		inReplyTo := next.InReplyTo
		if inReplyTo == nil {

			// Fetch next parent in conversation.
			inReplyTo, err = f.state.DB.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				next.InReplyToID,
			)
			if err != nil {
				return false, gtserror.Newf("error getting status parent %s: %w", next.InReplyToURI, err)
			}
		}

		// Set next status.
		next = inReplyTo
	}

	if next != status && !oneAuthor && !visible {
		log.Trace(ctx, "ignoring visible reply in conversation irrelevant to owner")
		return false, nil
	}

	// At this point status is either a top-level status, a reply in a single
	// author thread (e.g. "this is my weird-ass take and here is why 1/10 🧵"),
	// a status thread *including* the owner, or a conversation thread between
	// accounts the timeline owner follows.

	// Ensure owner follows author.
	follow, err := f.state.DB.GetFollow(ctx,
		owner.ID,
		status.AccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, gtserror.Newf("error retrieving follow %s->%s: %w", owner.ID, status.AccountID, err)
	}

	if follow == nil {
		log.Trace(ctx, "ignoring status from unfollowed author")
		return false, nil
	}

	if status.BoostOfID != "" && !*follow.ShowReblogs {
		// Status is a boost, but the owner of this follow
		// doesn't want to see boosts from this account.
		return false, nil
	}

	return true, nil
}

func (f *Filter) isVisibleConversation(
	ctx context.Context,
	owner *gtsmodel.Account,
	status *gtsmodel.Status,
) (
	bool, // explicitly IS visible
	bool, // explicitly NOT visible
	error, // err
) {
	// Check if status is visible to the timeline owner.
	visible, err := f.StatusVisible(ctx, owner, status)
	if err != nil {
		return false, false, err
	}

	if !visible {
		// Explicitly NOT visible
		// to the timeline owner.
		return false, true, nil
	}

	if status.Visibility == gtsmodel.VisibilityUnlocked ||
		status.Visibility == gtsmodel.VisibilityPublic {
		// NOTE: there is no need to check in the case of
		// direct / follow-only / mutual-only visibility statuses
		// as the above visibility check already handles this.

		// Check owner follows the status author.
		follow, err := f.state.DB.IsFollowing(ctx,
			owner.ID,
			status.AccountID,
		)
		if err != nil {
			return false, false, gtserror.Newf("error checking follow %s->%s: %w", owner.ID, status.AccountID, err)
		}

		if !follow {
			// Not explicitly visible
			// status to timeline owner.
			return false, false, nil
		}
	}

	var follow bool

	for _, mention := range status.Mentions {
		// Check block between timeline owner and mention.
		block, err := f.state.DB.IsEitherBlocked(ctx,
			owner.ID,
			mention.TargetAccountID,
		)
		if err != nil {
			return false, false, gtserror.Newf("error checking mention block %s<->%s: %w", owner.ID, mention.TargetAccountID, err)
		}

		if block {
			// Invisible conversation.
			return false, true, nil
		}

		if !follow {
			// See if tl owner follows any of mentions.
			follow, err = f.state.DB.IsFollowing(ctx,
				owner.ID,
				mention.TargetAccountID,
			)
			if err != nil {
				return false, false, gtserror.Newf("error checking mention follow %s->%s: %w", owner.ID, mention.TargetAccountID, err)
			}
		}
	}

	return follow, false, nil
}
