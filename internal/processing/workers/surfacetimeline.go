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

package workers

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
)

// timelineAndNotifyStatus inserts the given status into the HOME
// and LIST timelines of accounts that follow the status author.
//
// It will also handle notifications for any mentions attached to
// the account, and notifications for any local accounts that want
// to know when this account posts.
func (s *Surface) timelineAndNotifyStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Ensure status fully populated; including account, mentions, etc.
	if err := s.State.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status with id %s: %w", status.ID, err)
	}

	// Get all local followers of the account that posted the status.
	follows, err := s.State.DB.GetAccountLocalFollowers(ctx, status.AccountID)
	if err != nil {
		return gtserror.Newf("error getting local followers of account %s: %w", status.AccountID, err)
	}

	// If the poster is also local, add a fake entry for them
	// so they can see their own status in their timeline.
	if status.Account.IsLocal() {
		follows = append(follows, &gtsmodel.Follow{
			AccountID:   status.AccountID,
			Account:     status.Account,
			Notify:      func() *bool { b := false; return &b }(), // Account shouldn't notify itself.
			ShowReblogs: func() *bool { b := true; return &b }(),  // Account should show own reblogs.
		})
	}

	// Timeline the status for each local follower of this account.
	// This will also handle notifying any followers with notify
	// set to true on their follow.
	if err := s.timelineAndNotifyStatusForFollowers(ctx, status, follows); err != nil {
		return gtserror.Newf("error timelining status %s for followers: %w", status.ID, err)
	}

	// Notify each local account that's mentioned by this status.
	if err := s.notifyMentions(ctx, status); err != nil {
		return gtserror.Newf("error notifying status mentions for status %s: %w", status.ID, err)
	}

	return nil
}

// timelineAndNotifyStatusForFollowers iterates through the given
// slice of followers of the account that posted the given status,
// adding the status to list timelines + home timelines of each
// follower, as appropriate, and notifying each follower of the
// new status, if the status is eligible for notification.
func (s *Surface) timelineAndNotifyStatusForFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	follows []*gtsmodel.Follow,
) error {
	var (
		errs  gtserror.MultiError
		boost = status.BoostOfID != ""
		reply = status.InReplyToURI != ""
	)

	for _, follow := range follows {
		// Check to see if the status is timelineable for this follower,
		// taking account of its visibility, who it replies to, and, if
		// it's a reblog, whether follower account wants to see reblogs.
		//
		// If it's not timelineable, we can just stop early, since lists
		// are prettymuch subsets of the home timeline, so if it shouldn't
		// appear there, it shouldn't appear in lists either.
		timelineable, err := s.VisFilter.StatusHomeTimelineable(
			ctx, follow.Account, status,
		)
		if err != nil {
			errs.Appendf("error checking status %s hometimelineability: %w", status.ID, err)
			continue
		}

		if !timelineable {
			// Nothing to do.
			continue
		}

		filters, err := s.State.DB.GetFiltersForAccountID(ctx, follow.AccountID)
		if err != nil {
			return gtserror.Newf("couldn't retrieve filters for account %s: %w", follow.AccountID, err)
		}

		mutes, err := s.State.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), follow.AccountID, nil)
		if err != nil {
			return gtserror.Newf("couldn't retrieve mutes for account %s: %w", follow.AccountID, err)
		}
		compiledMutes := usermute.NewCompiledUserMuteList(mutes)

		// Add status to any relevant lists
		// for this follow, if applicable.
		s.listTimelineStatusForFollow(
			ctx,
			status,
			follow,
			&errs,
			filters,
			compiledMutes,
		)

		// Add status to home timeline for owner
		// of this follow, if applicable.
		homeTimelined, err := s.timelineStatus(
			ctx,
			s.State.Timelines.Home.IngestOne,
			follow.AccountID, // home timelines are keyed by account ID
			follow.Account,
			status,
			stream.TimelineHome,
			filters,
			compiledMutes,
		)
		if err != nil {
			errs.Appendf("error home timelining status: %w", err)
			continue
		}

		if !homeTimelined {
			// If status wasn't added to home
			// timeline, we shouldn't notify it.
			continue
		}

		if !*follow.Notify {
			// This follower doesn't have notifs
			// set for this account's new posts.
			continue
		}

		if boost || reply {
			// Don't notify for boosts or replies.
			continue
		}

		// If we reach here, we know:
		//
		//   - This status is hometimelineable.
		//   - This status was added to the home timeline for this follower.
		//   - This follower wants to be notified when this account posts.
		//   - This is a top-level post (not a reply or boost).
		//
		// That means we can officially notify this one.
		if err := s.Notify(ctx,
			gtsmodel.NotificationStatus,
			follow.Account,
			status.Account,
			status.ID,
		); err != nil {
			errs.Appendf("error notifying account %s about new status: %w", follow.AccountID, err)
		}
	}

	return errs.Combine()
}

// listTimelineStatusForFollow puts the given status
// in any eligible lists owned by the given follower.
func (s *Surface) listTimelineStatusForFollow(
	ctx context.Context,
	status *gtsmodel.Status,
	follow *gtsmodel.Follow,
	errs *gtserror.MultiError,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) {
	// To put this status in appropriate list timelines,
	// we need to get each listEntry that pertains to
	// this follow. Then, we want to iterate through all
	// those list entries, and add the status to the list
	// that the entry belongs to if it meets criteria for
	// inclusion in the list.

	// Get every list entry that targets this follow's ID.
	listEntries, err := s.State.DB.GetListEntriesForFollowID(
		// We only need the list IDs.
		gtscontext.SetBarebones(ctx),
		follow.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("error getting list entries: %w", err)
		return
	}

	// Check eligibility for each list entry (if any).
	for _, listEntry := range listEntries {
		eligible, err := s.listEligible(ctx, listEntry, status)
		if err != nil {
			errs.Appendf("error checking list eligibility: %w", err)
			continue
		}

		if !eligible {
			// Don't add this.
			continue
		}

		// At this point we are certain this status
		// should be included in the timeline of the
		// list that this list entry belongs to.
		if _, err := s.timelineStatus(
			ctx,
			s.State.Timelines.List.IngestOne,
			listEntry.ListID, // list timelines are keyed by list ID
			follow.Account,
			status,
			stream.TimelineList+":"+listEntry.ListID, // key streamType to this specific list
			filters,
			mutes,
		); err != nil {
			errs.Appendf("error adding status to timeline for list %s: %w", listEntry.ListID, err)
			// implicit continue
		}
	}
}

// listEligible checks if the given status is eligible
// for inclusion in the list that that the given listEntry
// belongs to, based on the replies policy of the list.
func (s *Surface) listEligible(
	ctx context.Context,
	listEntry *gtsmodel.ListEntry,
	status *gtsmodel.Status,
) (bool, error) {
	if status.InReplyToURI == "" {
		// If status is not a reply,
		// then it's all gravy baby.
		return true, nil
	}

	if status.InReplyToID == "" {
		// Status is a reply but we don't
		// have the replied-to account!
		return false, nil
	}

	// Status is a reply to a known account.
	// We need to fetch the list that this
	// entry belongs to, in order to check
	// the list's replies policy.
	list, err := s.State.DB.GetListByID(
		ctx, listEntry.ListID,
	)
	if err != nil {
		err := gtserror.Newf("db error getting list %s: %w", listEntry.ListID, err)
		return false, err
	}

	switch list.RepliesPolicy {
	case gtsmodel.RepliesPolicyNone:
		// This list should not show
		// replies at all, so skip it.
		return false, nil

	case gtsmodel.RepliesPolicyList:
		// This list should show replies
		// only to other people in the list.
		//
		// Check if replied-to account is
		// also included in this list.
		includes, err := s.State.DB.ListIncludesAccount(
			ctx,
			list.ID,
			status.InReplyToAccountID,
		)
		if err != nil {
			err := gtserror.Newf(
				"db error checking if account %s in list %s: %w",
				status.InReplyToAccountID, listEntry.ListID, err,
			)
			return false, err
		}

		return includes, nil

	case gtsmodel.RepliesPolicyFollowed:
		// This list should show replies
		// only to people that the list
		// owner also follows.
		//
		// Check if replied-to account is
		// followed by list owner account.
		follows, err := s.State.DB.IsFollowing(
			ctx,
			list.AccountID,
			status.InReplyToAccountID,
		)
		if err != nil {
			err := gtserror.Newf(
				"db error checking if account %s is followed by %s: %w",
				status.InReplyToAccountID, list.AccountID, err,
			)
			return false, err
		}

		return follows, nil

	default:
		// HUH??
		err := gtserror.Newf(
			"reply policy '%s' not recognized on list %s",
			list.RepliesPolicy, list.ID,
		)
		return false, err
	}
}

// timelineStatus uses the provided ingest function to put the given
// status in a timeline with the given ID, if it's timelineable.
//
// If the status was inserted into the timeline, true will be returned
// + it will also be streamed to the user using the given streamType.
func (s *Surface) timelineStatus(
	ctx context.Context,
	ingest func(context.Context, string, timeline.Timelineable) (bool, error),
	timelineID string,
	account *gtsmodel.Account,
	status *gtsmodel.Status,
	streamType string,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (bool, error) {
	// Ingest status into given timeline using provided function.
	if inserted, err := ingest(ctx, timelineID, status); err != nil {
		err = gtserror.Newf("error ingesting status %s: %w", status.ID, err)
		return false, err
	} else if !inserted {
		// Nothing more to do.
		return false, nil
	}

	// The status was inserted so stream it to the user.
	apiStatus, err := s.Converter.StatusToAPIStatus(ctx,
		status,
		account,
		statusfilter.FilterContextHome,
		filters,
		mutes,
	)
	if err != nil {
		err = gtserror.Newf("error converting status %s to frontend representation: %w", status.ID, err)
		return true, err
	}
	s.Stream.Update(ctx, account, apiStatus, streamType)

	return true, nil
}

// deleteStatusFromTimelines completely removes the given status from all timelines.
// It will also stream deletion of the status to all open streams.
func (s *Surface) deleteStatusFromTimelines(ctx context.Context, statusID string) error {
	if err := s.State.Timelines.Home.WipeItemFromAllTimelines(ctx, statusID); err != nil {
		return err
	}
	if err := s.State.Timelines.List.WipeItemFromAllTimelines(ctx, statusID); err != nil {
		return err
	}
	s.Stream.Delete(ctx, statusID)
	return nil
}

// invalidateStatusFromTimelines does cache invalidation on the given status by
// unpreparing it from all timelines, forcing it to be prepared again (with updated
// stats, boost counts, etc) next time it's fetched by the timeline owner. This goes
// both for the status itself, and for any boosts of the status.
func (s *Surface) invalidateStatusFromTimelines(ctx context.Context, statusID string) {
	if err := s.State.Timelines.Home.UnprepareItemFromAllTimelines(ctx, statusID); err != nil {
		log.
			WithContext(ctx).
			WithField("statusID", statusID).
			Errorf("error unpreparing status from home timelines: %v", err)
	}

	if err := s.State.Timelines.List.UnprepareItemFromAllTimelines(ctx, statusID); err != nil {
		log.
			WithContext(ctx).
			WithField("statusID", statusID).
			Errorf("error unpreparing status from list timelines: %v", err)
	}
}

// timelineStatusUpdate looks up HOME and LIST timelines of accounts
// that follow the the status author and pushes edit messages into any
// active streams.
// Note that calling invalidateStatusFromTimelines takes care of the
// state in general, we just need to do this for any streams that are
// open right now.
func (s *Surface) timelineStatusUpdate(ctx context.Context, status *gtsmodel.Status) error {
	// Ensure status fully populated; including account, mentions, etc.
	if err := s.State.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status with id %s: %w", status.ID, err)
	}

	// Get all local followers of the account that posted the status.
	follows, err := s.State.DB.GetAccountLocalFollowers(ctx, status.AccountID)
	if err != nil {
		return gtserror.Newf("error getting local followers of account %s: %w", status.AccountID, err)
	}

	// If the poster is also local, add a fake entry for them
	// so they can see their own status in their timeline.
	if status.Account.IsLocal() {
		follows = append(follows, &gtsmodel.Follow{
			AccountID:   status.AccountID,
			Account:     status.Account,
			Notify:      func() *bool { b := false; return &b }(), // Account shouldn't notify itself.
			ShowReblogs: func() *bool { b := true; return &b }(),  // Account should show own reblogs.
		})
	}

	// Push to streams for each local follower of this account.
	if err := s.timelineStatusUpdateForFollowers(ctx, status, follows); err != nil {
		return gtserror.Newf("error timelining status %s for followers: %w", status.ID, err)
	}

	return nil
}

// timelineStatusUpdateForFollowers iterates through the given
// slice of followers of the account that posted the given status,
// pushing update messages into open list/home streams of each
// follower.
func (s *Surface) timelineStatusUpdateForFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	follows []*gtsmodel.Follow,
) error {
	var (
		errs gtserror.MultiError
	)

	for _, follow := range follows {
		// Check to see if the status is timelineable for this follower,
		// taking account of its visibility, who it replies to, and, if
		// it's a reblog, whether follower account wants to see reblogs.
		//
		// If it's not timelineable, we can just stop early, since lists
		// are prettymuch subsets of the home timeline, so if it shouldn't
		// appear there, it shouldn't appear in lists either.
		timelineable, err := s.VisFilter.StatusHomeTimelineable(
			ctx, follow.Account, status,
		)
		if err != nil {
			errs.Appendf("error checking status %s hometimelineability: %w", status.ID, err)
			continue
		}

		if !timelineable {
			// Nothing to do.
			continue
		}

		filters, err := s.State.DB.GetFiltersForAccountID(ctx, follow.AccountID)
		if err != nil {
			return gtserror.Newf("couldn't retrieve filters for account %s: %w", follow.AccountID, err)
		}

		mutes, err := s.State.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), follow.AccountID, nil)
		if err != nil {
			return gtserror.Newf("couldn't retrieve mutes for account %s: %w", follow.AccountID, err)
		}
		compiledMutes := usermute.NewCompiledUserMuteList(mutes)

		// Add status to any relevant lists
		// for this follow, if applicable.
		s.listTimelineStatusUpdateForFollow(
			ctx,
			status,
			follow,
			&errs,
			filters,
			compiledMutes,
		)

		// Add status to home timeline for owner
		// of this follow, if applicable.
		err = s.timelineStreamStatusUpdate(
			ctx,
			follow.Account,
			status,
			stream.TimelineHome,
			filters,
			compiledMutes,
		)
		if err != nil {
			errs.Appendf("error home timelining status: %w", err)
			continue
		}
	}

	return errs.Combine()
}

// listTimelineStatusUpdateForFollow pushes edits of the given status
// into any eligible lists streams opened by the given follower.
func (s *Surface) listTimelineStatusUpdateForFollow(
	ctx context.Context,
	status *gtsmodel.Status,
	follow *gtsmodel.Follow,
	errs *gtserror.MultiError,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) {
	// To put this status in appropriate list timelines,
	// we need to get each listEntry that pertains to
	// this follow. Then, we want to iterate through all
	// those list entries, and add the status to the list
	// that the entry belongs to if it meets criteria for
	// inclusion in the list.

	// Get every list entry that targets this follow's ID.
	listEntries, err := s.State.DB.GetListEntriesForFollowID(
		// We only need the list IDs.
		gtscontext.SetBarebones(ctx),
		follow.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("error getting list entries: %w", err)
		return
	}

	// Check eligibility for each list entry (if any).
	for _, listEntry := range listEntries {
		eligible, err := s.listEligible(ctx, listEntry, status)
		if err != nil {
			errs.Appendf("error checking list eligibility: %w", err)
			continue
		}

		if !eligible {
			// Don't add this.
			continue
		}

		// At this point we are certain this status
		// should be included in the timeline of the
		// list that this list entry belongs to.
		if err := s.timelineStreamStatusUpdate(
			ctx,
			follow.Account,
			status,
			stream.TimelineList+":"+listEntry.ListID, // key streamType to this specific list
			filters,
			mutes,
		); err != nil {
			errs.Appendf("error adding status to timeline for list %s: %w", listEntry.ListID, err)
			// implicit continue
		}
	}
}

// timelineStatusUpdate streams the edited status to the user using the
// given streamType.
func (s *Surface) timelineStreamStatusUpdate(
	ctx context.Context,
	account *gtsmodel.Account,
	status *gtsmodel.Status,
	streamType string,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) error {
	apiStatus, err := s.Converter.StatusToAPIStatus(ctx, status, account, statusfilter.FilterContextHome, filters, mutes)
	if errors.Is(err, statusfilter.ErrHideStatus) {
		// Don't put this status in the stream.
		return nil
	}
	if err != nil {
		err = gtserror.Newf("error converting status %s to frontend representation: %w", status.ID, err)
		return err
	}
	s.Stream.StatusUpdate(ctx, account, apiStatus, streamType)
	return nil
}
