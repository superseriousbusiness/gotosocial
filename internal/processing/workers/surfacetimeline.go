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

	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// timelineAndNotifyStatus inserts the given status into the HOME
// and/or LIST timelines of accounts that follow the status author,
// as well as the HOME timelines of accounts that follow tags used by the status.
//
// It will also handle notifications for any mentions attached to
// the account, notifications for any local accounts that want
// to know when this account posts, and conversations containing the status.
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
			Notify:      util.Ptr(false), // Account shouldn't notify itself.
			ShowReblogs: util.Ptr(true),  // Account should show own reblogs.
		})
	}

	// Timeline the status for each local follower of this account. This will
	// also handle notifying any followers with notify set to true on their follow.
	homeTimelinedAccountIDs := s.timelineAndNotifyStatusForFollowers(ctx, status, follows)

	// Timeline the status for each local account who follows a tag used by this status.
	if err := s.timelineAndNotifyStatusForTagFollowers(ctx, status, homeTimelinedAccountIDs); err != nil {
		return gtserror.Newf("error timelining status %s for tag followers: %w", status.ID, err)
	}

	// Notify each local account that's mentioned by this status.
	if err := s.notifyMentions(ctx, status); err != nil {
		return gtserror.Newf("error notifying status mentions for status %s: %w", status.ID, err)
	}

	// Update any conversations containing this status, and send conversation notifications.
	notifications, err := s.Conversations.UpdateConversationsForStatus(ctx, status)
	if err != nil {
		return gtserror.Newf("error updating conversations for status %s: %w", status.ID, err)
	}
	for _, notification := range notifications {
		s.Stream.Conversation(ctx, notification.AccountID, notification.Conversation)
	}

	return nil
}

// timelineAndNotifyStatusForFollowers iterates through the given
// slice of followers of the account that posted the given status,
// adding the status to list timelines + home timelines of each
// follower, as appropriate, and notifying each follower of the
// new status, if the status is eligible for notification.
//
// Returns a list of accounts which had this status inserted into their home timelines.
// This will be used to prevent duplicate inserts when handling followed tags.
func (s *Surface) timelineAndNotifyStatusForFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	follows []*gtsmodel.Follow,
) (homeTimelinedAccountIDs []string) {
	var (
		boost = (status.BoostOfID != "")
		reply = (status.InReplyToURI != "")
	)

	for _, follow := range follows {
		// Check to see if the status is timelineable for this follower,
		// taking account of its visibility, who it replies to, and, if
		// it's a reblog, whether follower account wants to see reblogs.
		//
		// If it's not timelineable, we can just stop early, since lists
		// are pretty much subsets of the home timeline, so if it shouldn't
		// appear there, it shouldn't appear in lists either.
		//
		// Exclusive lists don't change this:
		// if something is hometimelineable according to this filter,
		// it's also eligible to appear in exclusive lists,
		// even if it ultimately doesn't appear on the home timeline.
		timelineable, err := s.VisFilter.StatusHomeTimelineable(
			ctx, follow.Account, status,
		)
		if err != nil {
			log.Errorf(ctx, "error checking status home visibility for follow: %v", err)
			continue
		}

		if !timelineable {
			// Nothing to do.
			continue
		}

		// Get relevant filters and mutes for this follow's account.
		// (note the origin account of the follow is receiver of status).
		filters, mutes, err := s.getFiltersAndMutes(ctx, follow.AccountID)
		if err != nil {
			log.Error(ctx, err)
			continue
		}

		// Add status to any relevant lists for this follow, if applicable.
		listTimelined, exclusive, err := s.listTimelineStatusForFollow(ctx,
			status,
			follow,
			filters,
			mutes,
		)
		if err != nil {
			log.Errorf(ctx, "error list timelining status: %v", err)
			continue
		}

		var homeTimelined bool

		// If this was timelined into
		// list with exclusive flag set,
		// don't add to home timeline.
		if !exclusive {

			// Add status to home timeline for owner of
			// this follow (origin account), if applicable.
			homeTimelined, err = s.timelineStatus(ctx,
				s.State.Timelines.Home.IngestOne,
				follow.AccountID, // home timelines are keyed by account ID
				follow.Account,
				status,
				stream.TimelineHome,
				filters,
				mutes,
			)
			if err != nil {
				log.Errorf(ctx, "error home timelining status: %v", err)
				continue
			}

			if homeTimelined {
				// If hometimelined, add to list of returned account IDs.
				homeTimelinedAccountIDs = append(homeTimelinedAccountIDs, follow.AccountID)
			}
		}

		if !(homeTimelined || listTimelined) {
			// If status wasn't added to home or list
			// timelines, we shouldn't notify it.
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
		//   - This status was added to the home timeline and/or list timelines for this follower.
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
			log.Errorf(ctx, "error notifying status for account: %v", err)
			continue
		}
	}

	return homeTimelinedAccountIDs
}

// listTimelineStatusForFollow puts the given status
// in any eligible lists owned by the given follower.
//
// It returns whether the status was added to any lists,
// and whether the status author is on any exclusive lists
// (in which case the status shouldn't be added to the home timeline).
func (s *Surface) listTimelineStatusForFollow(
	ctx context.Context,
	status *gtsmodel.Status,
	follow *gtsmodel.Follow,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (timelined bool, exclusive bool, err error) {

	// Get all lists that contain this given follow.
	lists, err := s.State.DB.GetListsContainingFollowID(

		// We don't need list sub-models.
		gtscontext.SetBarebones(ctx),
		follow.ID,
	)
	if err != nil {
		return false, false, gtserror.Newf("error getting lists for follow: %w", err)
	}

	for _, list := range lists {
		// Check whether list is eligible for this status.
		eligible, err := s.listEligible(ctx, list, status)
		if err != nil {
			log.Errorf(ctx, "error checking list eligibility: %v", err)
			continue
		}

		if !eligible {
			continue
		}

		// Update exclusive flag if list is so.
		exclusive = exclusive || *list.Exclusive

		// At this point we are certain this status
		// should be included in the timeline of the
		// list that this list entry belongs to.
		listTimelined, err := s.timelineStatus(
			ctx,
			s.State.Timelines.List.IngestOne,
			list.ID, // list timelines are keyed by list ID
			follow.Account,
			status,
			stream.TimelineList+":"+list.ID, // key streamType to this specific list
			filters,
			mutes,
		)
		if err != nil {
			log.Errorf(ctx, "error adding status to list timeline: %v", err)
			continue
		}

		// Update flag based on if timelined.
		timelined = timelined || listTimelined
	}

	return timelined, exclusive, nil
}

// getFiltersAndMutes returns an account's filters and mutes.
func (s *Surface) getFiltersAndMutes(ctx context.Context, accountID string) ([]*gtsmodel.Filter, *usermute.CompiledUserMuteList, error) {
	filters, err := s.State.DB.GetFiltersForAccountID(ctx, accountID)
	if err != nil {
		return nil, nil, gtserror.Newf("couldn't retrieve filters for account %s: %w", accountID, err)
	}

	mutes, err := s.State.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), accountID, nil)
	if err != nil {
		return nil, nil, gtserror.Newf("couldn't retrieve mutes for account %s: %w", accountID, err)
	}

	compiledMutes := usermute.NewCompiledUserMuteList(mutes)
	return filters, compiledMutes, err
}

// listEligible checks if the given status is eligible
// for inclusion in the list that that the given listEntry
// belongs to, based on the replies policy of the list.
func (s *Surface) listEligible(
	ctx context.Context,
	list *gtsmodel.List,
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
		in, err := s.State.DB.IsAccountInList(ctx,
			list.ID,
			status.InReplyToAccountID,
		)
		if err != nil {
			err := gtserror.Newf("db error checking if account in list: %w", err)
			return false, err
		}
		return in, nil

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
			err := gtserror.Newf("db error checking if account followed: %w", err)
			return false, err
		}
		return follows, nil

	default:
		log.Panicf(ctx, "unknown reply policy: %s", list.RepliesPolicy)
		return false, nil // unreachable code
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
	if inserted, err := ingest(ctx, timelineID, status); err != nil &&
		!errors.Is(err, statusfilter.ErrHideStatus) {
		err := gtserror.Newf("error ingesting status %s: %w", status.ID, err)
		return false, err
	} else if !inserted {
		// Nothing more to do.
		return false, nil
	}

	// Convert updated database model to frontend model.
	apiStatus, err := s.Converter.StatusToAPIStatus(ctx,
		status,
		account,
		statusfilter.FilterContextHome,
		filters,
		mutes,
	)
	if err != nil && !errors.Is(err, statusfilter.ErrHideStatus) {
		err := gtserror.Newf("error converting status %s to frontend representation: %w", status.ID, err)
		return true, err
	}

	if apiStatus != nil {
		// The status was inserted so stream it to the user.
		s.Stream.Update(ctx, account, apiStatus, streamType)
		return true, nil
	}

	// Status was hidden.
	return false, nil
}

// timelineAndNotifyStatusForTagFollowers inserts the status into the
// home timeline of each local account which follows a useable tag from the status,
// skipping accounts for which it would have already been inserted.
func (s *Surface) timelineAndNotifyStatusForTagFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	alreadyHomeTimelinedAccountIDs []string,
) error {
	tagFollowerAccounts, err := s.tagFollowersForStatus(ctx, status, alreadyHomeTimelinedAccountIDs)
	if err != nil {
		return err
	}

	if status.BoostOf != nil {
		// Unwrap boost and work
		// with the original status.
		status = status.BoostOf
	}

	// Insert the status into the home timeline of each tag follower.
	errs := gtserror.MultiError{}
	for _, tagFollowerAccount := range tagFollowerAccounts {
		filters, mutes, err := s.getFiltersAndMutes(ctx, tagFollowerAccount.ID)
		if err != nil {
			errs.Append(err)
			continue
		}

		if _, err := s.timelineStatus(
			ctx,
			s.State.Timelines.Home.IngestOne,
			tagFollowerAccount.ID, // home timelines are keyed by account ID
			tagFollowerAccount,
			status,
			stream.TimelineHome,
			filters,
			mutes,
		); err != nil {
			errs.Appendf(
				"error inserting status %s into home timeline for account %s: %w",
				status.ID,
				tagFollowerAccount.ID,
				err,
			)
		}
	}

	return errs.Combine()
}

// tagFollowersForStatus gets local accounts which follow any useable tags from the status,
// skipping any with IDs in the provided list, and any that shouldn't be able to see it due to blocks.
func (s *Surface) tagFollowersForStatus(
	ctx context.Context,
	status *gtsmodel.Status,
	skipAccountIDs []string,
) ([]*gtsmodel.Account, error) {
	// If the status is a boost, look at the tags from the boosted status.
	taggedStatus := status
	if status.BoostOf != nil {
		taggedStatus = status.BoostOf
	}

	if taggedStatus.Visibility != gtsmodel.VisibilityPublic || len(taggedStatus.Tags) == 0 {
		// Only public statuses with tags are eligible for tag processing.
		return nil, nil
	}

	// Build list of useable tag IDs.
	useableTagIDs := make([]string, 0, len(taggedStatus.Tags))
	for _, tag := range taggedStatus.Tags {
		if *tag.Useable {
			useableTagIDs = append(useableTagIDs, tag.ID)
		}
	}
	if len(useableTagIDs) == 0 {
		return nil, nil
	}

	// Get IDs for all accounts who follow one or more of the useable tags from this status.
	allTagFollowerAccountIDs, err := s.State.DB.GetAccountIDsFollowingTagIDs(ctx, useableTagIDs)
	if err != nil {
		return nil, gtserror.Newf("DB error getting followers for tags of status %s: %w", taggedStatus.ID, err)
	}
	if len(allTagFollowerAccountIDs) == 0 {
		return nil, nil
	}

	// Build set for faster lookup of account IDs to skip.
	skipAccountIDSet := make(map[string]struct{}, len(skipAccountIDs))
	for _, accountID := range skipAccountIDs {
		skipAccountIDSet[accountID] = struct{}{}
	}

	// Build list of tag follower account IDs,
	// except those which have already had this status inserted into their timeline.
	tagFollowerAccountIDs := make([]string, 0, len(allTagFollowerAccountIDs))
	for _, accountID := range allTagFollowerAccountIDs {
		if _, skip := skipAccountIDSet[accountID]; skip {
			continue
		}
		tagFollowerAccountIDs = append(tagFollowerAccountIDs, accountID)
	}
	if len(tagFollowerAccountIDs) == 0 {
		return nil, nil
	}

	// Retrieve accounts for remaining tag followers.
	tagFollowerAccounts, err := s.State.DB.GetAccountsByIDs(ctx, tagFollowerAccountIDs)
	if err != nil {
		return nil, gtserror.Newf("DB error getting accounts for followers of tags of status %s: %w", taggedStatus.ID, err)
	}

	// Check the visibility of the *input* status for each account.
	// This accounts for the visibility of the boost as well as the original, if the input status is a boost.
	errs := gtserror.MultiError{}
	visibleTagFollowerAccounts := make([]*gtsmodel.Account, 0, len(tagFollowerAccounts))
	for _, account := range tagFollowerAccounts {
		visible, err := s.VisFilter.StatusVisible(ctx, account, status)
		if err != nil {
			errs.Appendf(
				"error checking visibility of status %s to account %s",
				status.ID,
				account.ID,
			)
		}
		if visible {
			visibleTagFollowerAccounts = append(visibleTagFollowerAccounts, account)
		}
	}

	return visibleTagFollowerAccounts, errs.Combine()
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
// that follow the the status author or tags and pushes edit messages into any
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
			Notify:      util.Ptr(false), // Account shouldn't notify itself.
			ShowReblogs: util.Ptr(true),  // Account should show own reblogs.
		})
	}

	// Push updated status to streams for each local follower of this account.
	homeTimelinedAccountIDs := s.timelineStatusUpdateForFollowers(ctx, status, follows)

	// Push updated status to streams for each local follower of tags in status, if applicable.
	if err := s.timelineStatusUpdateForTagFollowers(ctx, status, homeTimelinedAccountIDs); err != nil {
		return gtserror.Newf("error timelining status %s for tag followers: %w", status.ID, err)
	}

	return nil
}

// timelineStatusUpdateForFollowers iterates through the given
// slice of followers of the account that posted the given status,
// pushing update messages into open list/home streams of each
// follower.
//
// Returns a list of accounts which had this status updated in their home timelines.
func (s *Surface) timelineStatusUpdateForFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	follows []*gtsmodel.Follow,
) (homeTimelinedAccountIDs []string) {
	for _, follow := range follows {
		// Check to see if the status is timelineable for this follower,
		// taking account of its visibility, who it replies to, and, if
		// it's a reblog, whether follower account wants to see reblogs.
		//
		// If it's not timelineable, we can just stop early, since lists
		// are pretty much subsets of the home timeline, so if it shouldn't
		// appear there, it shouldn't appear in lists either.
		//
		// Exclusive lists don't change this:
		// if something is hometimelineable according to this filter,
		// it's also eligible to appear in exclusive lists,
		// even if it ultimately doesn't appear on the home timeline.
		timelineable, err := s.VisFilter.StatusHomeTimelineable(
			ctx, follow.Account, status,
		)
		if err != nil {
			log.Errorf(ctx, "error checking status home visibility for follow: %v", err)
			continue
		}

		if !timelineable {
			// Nothing to do.
			continue
		}

		// Get relevant filters and mutes for this follow's account.
		// (note the origin account of the follow is receiver of status).
		filters, mutes, err := s.getFiltersAndMutes(ctx, follow.AccountID)
		if err != nil {
			log.Error(ctx, err)
			continue
		}

		// Add status to relevant lists for this follow, if applicable.
		_, exclusive, err := s.listTimelineStatusUpdateForFollow(ctx,
			status,
			follow,
			filters,
			mutes,
		)
		if err != nil {
			log.Errorf(ctx, "error list timelining status: %v", err)
			continue
		}

		// If this was timelined into
		// list with exclusive flag set,
		// don't add to home timeline.
		if exclusive {
			continue
		}

		// Add status to home timeline for owner of
		// this follow (origin account), if applicable.
		homeTimelined, err := s.timelineStreamStatusUpdate(ctx,
			follow.Account,
			status,
			stream.TimelineHome,
			filters,
			mutes,
		)
		if err != nil {
			log.Errorf(ctx, "error home timelining status: %v", err)
			continue
		}

		if homeTimelined {
			// If hometimelined, add to list of returned account IDs.
			homeTimelinedAccountIDs = append(homeTimelinedAccountIDs, follow.AccountID)
		}
	}

	return homeTimelinedAccountIDs
}

// listTimelineStatusUpdateForFollow pushes edits of the given status
// into any eligible lists streams opened by the given follower.
//
// It returns whether the status author is on any exclusive lists
// (in which case the status shouldn't be added to the home timeline).
func (s *Surface) listTimelineStatusUpdateForFollow(
	ctx context.Context,
	status *gtsmodel.Status,
	follow *gtsmodel.Follow,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (bool, bool, error) {

	// Get all lists that contain this given follow.
	lists, err := s.State.DB.GetListsContainingFollowID(

		// We don't need list sub-models.
		gtscontext.SetBarebones(ctx),
		follow.ID,
	)
	if err != nil {
		return false, false, gtserror.Newf("error getting lists for follow: %w", err)
	}

	var exclusive, timelined bool
	for _, list := range lists {

		// Check whether list is eligible for this status.
		eligible, err := s.listEligible(ctx, list, status)
		if err != nil {
			log.Errorf(ctx, "error checking list eligibility: %v", err)
			continue
		}

		if !eligible {
			continue
		}

		// Update exclusive flag if list is so.
		exclusive = exclusive || *list.Exclusive

		// At this point we are certain this status
		// should be included in the timeline of the
		// list that this list entry belongs to.
		listTimelined, err := s.timelineStreamStatusUpdate(
			ctx,
			follow.Account,
			status,
			stream.TimelineList+":"+list.ID, // key streamType to this specific list
			filters,
			mutes,
		)
		if err != nil {
			log.Errorf(ctx, "error adding status to list timeline: %v", err)
			continue
		}

		// Update flag based on if timelined.
		timelined = timelined || listTimelined
	}

	return timelined, exclusive, nil
}

// timelineStatusUpdate streams the edited status to the user using the
// given streamType.
//
// Returns whether it was actually streamed.
func (s *Surface) timelineStreamStatusUpdate(
	ctx context.Context,
	account *gtsmodel.Account,
	status *gtsmodel.Status,
	streamType string,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) (bool, error) {

	// Convert updated database model to frontend model.
	apiStatus, err := s.Converter.StatusToAPIStatus(ctx,
		status,
		account,
		statusfilter.FilterContextHome,
		filters,
		mutes,
	)

	switch {
	case err == nil:
		// no issue.

	case errors.Is(err, statusfilter.ErrHideStatus):
		// Don't put this status in the stream.
		return false, nil

	default:
		return false, gtserror.Newf("error converting status: %w", err)
	}

	// The status was updated so stream it to the user.
	s.Stream.StatusUpdate(ctx, account, apiStatus, streamType)

	return true, nil
}

// timelineStatusUpdateForTagFollowers streams update notifications to the
// home timeline of each local account which follows a tag used by the status,
// skipping accounts for which it would have already been streamed.
func (s *Surface) timelineStatusUpdateForTagFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	alreadyHomeTimelinedAccountIDs []string,
) error {
	tagFollowerAccounts, err := s.tagFollowersForStatus(ctx, status, alreadyHomeTimelinedAccountIDs)
	if err != nil {
		return err
	}

	if status.BoostOf != nil {
		// Unwrap boost and work with the original status.
		status = status.BoostOf
	}

	// Stream the update to the home timeline of each tag follower.
	errs := gtserror.MultiError{}
	for _, tagFollowerAccount := range tagFollowerAccounts {
		filters, mutes, err := s.getFiltersAndMutes(ctx, tagFollowerAccount.ID)
		if err != nil {
			errs.Append(err)
			continue
		}

		if _, err := s.timelineStreamStatusUpdate(
			ctx,
			tagFollowerAccount,
			status,
			stream.TimelineHome,
			filters,
			mutes,
		); err != nil {
			errs.Appendf(
				"error updating status %s on home timeline for account %s: %w",
				status.ID,
				tagFollowerAccount.ID,
				err,
			)
		}
	}
	return errs.Combine()
}
