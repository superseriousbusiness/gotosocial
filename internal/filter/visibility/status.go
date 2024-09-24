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
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// StatusesVisible calls StatusVisible for each status in the statuses slice, and returns a slice of only statuses which are visible to the requester.
func (f *Filter) StatusesVisible(ctx context.Context, requester *gtsmodel.Account, statuses []*gtsmodel.Status) ([]*gtsmodel.Status, error) {
	var errs gtserror.MultiError
	filtered := slices.DeleteFunc(statuses, func(status *gtsmodel.Status) bool {
		visible, err := f.StatusVisible(ctx, requester, status)
		if err != nil {
			errs.Append(err)
			return true
		}
		return !visible
	})
	return filtered, errs.Combine()
}

// StatusVisible will check if status is visible to requester,
// accounting for requester with no auth (i.e is nil), suspensions,
// disabled local users, pending approvals, account blocks,
// and status visibility settings.
func (f *Filter) StatusVisible(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (bool, error) {
	const vtype = cache.VisibilityTypeStatus

	// By default we assume no auth.
	requesterID := NoAuth

	if requester != nil {
		// Use provided account ID.
		requesterID = requester.ID
	}

	visibility, err := f.state.Caches.Visibility.LoadOne("Type,RequesterID,ItemID", func() (*cache.CachedVisibility, error) {
		// Visibility not yet cached, perform visibility lookup.
		visible, err := f.isStatusVisible(ctx, requester, status)
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
		return false, err
	}

	return visibility.Value, nil
}

// isStatusVisible will check if status is visible to requester.
// It is the "meat" of the logic to Filter{}.StatusVisible()
// which is called within cache loader callback.
func (f *Filter) isStatusVisible(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (bool, error) {
	// Ensure that status is fully populated for further processing.
	if err := f.state.DB.PopulateStatus(ctx, status); err != nil {
		return false, gtserror.Newf("error populating status %s: %w", status.ID, err)
	}

	// Check whether status accounts are visible to the requester.
	acctsVisible, err := f.areStatusAccountsVisible(ctx, requester, status)
	if err != nil {
		return false, gtserror.Newf("error checking status %s account visibility: %w", status.ID, err)
	} else if !acctsVisible {
		return false, nil
	}

	if util.PtrOrZero(status.PendingApproval) {
		// Use a different visibility heuristic
		// for pending approval statuses.
		return isPendingStatusVisible(
			requester, status,
		), nil
	}

	if requester == nil {
		// Use a different visibility
		// heuristic for unauthed requests.
		return f.isStatusVisibleUnauthed(
			ctx, status,
		)
	}

	/*
		From this point down we know the request is authed.
	*/

	if requester.IsRemote() && status.IsLocalOnly() {
		// Remote accounts can't see local-only
		// posts regardless of their visibility.
		return false, nil
	}

	if status.Visibility == gtsmodel.VisibilityPublic ||
		status.Visibility == gtsmodel.VisibilityUnlocked {
		// This status is visible to all auth'd accounts
		// (pending blocks, which we already checked above).
		return true, nil
	}

	/*
		From this point down we know the request
		is of visibility followers-only or below.
	*/

	if requester.ID == status.AccountID {
		// Author can always see their own status.
		return true, nil
	}

	if status.MentionsAccount(requester.ID) {
		// Status mentions the requesting account.
		return true, nil
	}

	if status.BoostOf != nil {
		if !status.BoostOf.MentionsPopulated() {
			// Boosted status needs its mentions populating, fetch these from database.
			status.BoostOf.Mentions, err = f.state.DB.GetMentions(ctx, status.BoostOf.MentionIDs)
			if err != nil {
				return false, gtserror.Newf("error populating boosted status %s mentions: %w", status.BoostOfID, err)
			}
		}

		if status.BoostOf.MentionsAccount(requester.ID) {
			// Boosted status mentions the requesting account.
			return true, nil
		}
	}

	switch status.Visibility {
	case gtsmodel.VisibilityFollowersOnly:
		// Check requester follows status author.
		follows, err := f.state.DB.IsFollowing(ctx,
			requester.ID,
			status.AccountID,
		)
		if err != nil {
			return false, gtserror.Newf("error checking follow %s->%s: %w", requester.ID, status.AccountID, err)
		}

		if !follows {
			log.Trace(ctx, "follow-only status not visible to requester")
			return false, nil
		}

		return true, nil

	case gtsmodel.VisibilityMutualsOnly:
		// Check mutual following between requester and author.
		mutuals, err := f.state.DB.IsMutualFollowing(ctx,
			requester.ID,
			status.AccountID,
		)
		if err != nil {
			return false, gtserror.Newf("error checking mutual follow %s<->%s: %w", requester.ID, status.AccountID, err)
		}

		if !mutuals {
			log.Trace(ctx, "mutual-only status not visible to requester")
			return false, nil
		}

		return true, nil

	case gtsmodel.VisibilityDirect:
		log.Trace(ctx, "direct status not visible to requester")
		return false, nil

	default:
		log.Warnf(ctx, "unexpected status visibility %s for %s", status.Visibility, status.URI)
		return false, nil
	}
}

// isPendingStatusVisible returns whether a status pending approval is visible to requester.
func isPendingStatusVisible(requester *gtsmodel.Account, status *gtsmodel.Status) bool {
	if requester == nil {
		// Any old tom, dick, and harry can't
		// see pending-approval statuses,
		// no matter what their visibility.
		return false
	}

	if status.AccountID == requester.ID {
		// This is requester's status,
		// so they can always see it.
		return true
	}

	if status.InReplyToAccountID == requester.ID {
		// This status replies to requester,
		// so they can always see it (else
		// they can't approve it).
		return true
	}

	if status.BoostOfAccountID == requester.ID {
		// This status boosts requester,
		// so they can always see it.
		return true
	}

	// Nobody else
	// can see this.
	return false
}

// isStatusVisibleUnauthed returns whether status is visible without any unauthenticated account.
func (f *Filter) isStatusVisibleUnauthed(ctx context.Context, status *gtsmodel.Status) (bool, error) {

	// For remote accounts, only show
	// Public statuses via the web.
	if status.Account.IsRemote() {
		return status.Visibility == gtsmodel.VisibilityPublic, nil
	}

	// If status is local only,
	// never show via the web.
	if status.IsLocalOnly() {
		return false, nil
	}

	// Check account's settings to see
	// what they expose. Populate these
	// from the DB if necessary.
	if status.Account.Settings == nil {
		var err error
		status.Account.Settings, err = f.state.DB.GetAccountSettings(ctx, status.Account.ID)
		if err != nil {
			return false, gtserror.Newf(
				"error getting settings for account %s: %w",
				status.Account.ID, err,
			)
		}
	}

	switch webvis := status.Account.Settings.WebVisibility; webvis {

	// public_only: status must be Public.
	case gtsmodel.VisibilityPublic:
		return status.Visibility == gtsmodel.VisibilityPublic, nil

	// unlisted: status must be Public or Unlocked.
	case gtsmodel.VisibilityUnlocked:
		visible := status.Visibility == gtsmodel.VisibilityPublic ||
			status.Visibility == gtsmodel.VisibilityUnlocked
		return visible, nil

	// none: never show via the web.
	case gtsmodel.VisibilityNone:
		return false, nil

	// Huh?
	default:
		return false, gtserror.Newf(
			"unrecognized web visibility for account %s: %s",
			status.Account.ID, webvis,
		)
	}
}

// areStatusAccountsVisible calls Filter{}.AccountVisible() on status author and the status boost-of (if set) author, returning visibility of status (and boost-of) to requester.
func (f *Filter) areStatusAccountsVisible(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	// Check whether status author's account is visible to requester.
	visible, err := f.AccountVisible(ctx, requester, status.Account)
	if err != nil {
		return false, gtserror.Newf("error checking status author visibility: %w", err)
	}

	if !visible {
		log.Trace(ctx, "status author not visible to requester")
		return false, nil
	}

	if status.BoostOfID != "" {
		// This is a boosted status.

		if status.AccountID == status.BoostOfAccountID {
			// Some clout-chaser boosted their own status, tch.
			return true, nil
		}

		// Check whether boosted status author's account is visible to requester.
		visible, err := f.AccountVisible(ctx, requester, status.BoostOfAccount)
		if err != nil {
			return false, gtserror.Newf("error checking boosted author visibility: %w", err)
		}

		if !visible {
			log.Trace(ctx, "boosted status author not visible to requester")
			return false, nil
		}
	}

	return true, nil
}
