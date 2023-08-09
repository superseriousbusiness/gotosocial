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
func (p *Processor) timelineAndNotifyStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Ensure status fully populated; including account, mentions, etc.
	if err := p.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status with id %s: %w", status.ID, err)
	}

	// Get all local followers of the account that posted the status.
	follows, err := p.state.DB.GetAccountLocalFollowers(ctx, status.AccountID)
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
	if err := p.timelineAndNotifyStatusForFollowers(ctx, status, follows); err != nil {
		return gtserror.Newf("error timelining status %s for followers: %w", status.ID, err)
	}

	// Notify each local account that's mentioned by this status.
	if err := p.notifyMentions(ctx, status.Mentions); err != nil {
		return gtserror.Newf("error notifying status mentions for status %s: %w", status.ID, err)
	}

	return nil
}

// timelineAndNotifyStatusForFollowers iterates through the given
// slice of followers of the account that posted the given status,
// adding the status to list timelines + home timelines of each
// follower, as appropriate, and notifying each follower of the
// new status, if the status is eligible for notification.
func (p *Processor) timelineAndNotifyStatusForFollowers(
	ctx context.Context,
	status *gtsmodel.Status,
	follows []*gtsmodel.Follow,
) error {
	var (
		errs  = new(gtserror.MultiError)
		boost = status.BoostOfID != ""
		reply = status.InReplyToURI != ""
	)

	for _, follow := range follows {
		// Do an initial rough-grained check to see if the
		// status is timelineable for this follower at all
		// based on its visibility and who it replies to etc.
		timelineable, err := p.filter.StatusHomeTimelineable(
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

		if boost && !*follow.ShowReblogs {
			// Status is a boost, but the owner of
			// this follow doesn't want to see boosts
			// from this account. We can safely skip
			// everything, then, because we also know
			// that the follow owner won't want to be
			// have the status put in any list timelines,
			// or be notified about the status either.
			continue
		}

		// Add status to any relevant lists
		// for this follow, if applicable.
		p.listTimelineStatusForFollow(
			ctx,
			status,
			follow,
			errs,
		)

		// Add status to home timeline for owner
		// of this follow, if applicable.
		homeTimelined, err := p.timelineStatus(
			ctx,
			p.state.Timelines.Home.IngestOne,
			follow.AccountID, // home timelines are keyed by account ID
			follow.Account,
			status,
			stream.TimelineHome,
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
		if err := p.notify(
			ctx,
			gtsmodel.NotificationStatus,
			follow.AccountID,
			status.AccountID,
			status.ID,
		); err != nil {
			errs.Appendf("error notifying account %s about new status: %w", follow.AccountID, err)
		}
	}

	return errs.Combine()
}

// listTimelineStatusForFollow puts the given status
// in any eligible lists owned by the given follower.
func (p *Processor) listTimelineStatusForFollow(
	ctx context.Context,
	status *gtsmodel.Status,
	follow *gtsmodel.Follow,
	errs *gtserror.MultiError,
) {
	// To put this status in appropriate list timelines,
	// we need to get each listEntry that pertains to
	// this follow. Then, we want to iterate through all
	// those list entries, and add the status to the list
	// that the entry belongs to if it meets criteria for
	// inclusion in the list.

	// Get every list entry that targets this follow's ID.
	listEntries, err := p.state.DB.GetListEntriesForFollowID(
		// We only need the list IDs.
		gtscontext.SetBarebones(ctx),
		follow.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("error list timelining status: %w", err)
		return
	}

	// Check eligibility for each list entry (if any).
	for _, listEntry := range listEntries {
		eligible, err := p.listEligible(ctx, listEntry, status)
		if err != nil {
			errs.Append(err)
		}

		if !eligible {
			// Don't add this.
			continue
		}

		// At this point we are certain this status
		// should be included in the timeline of the
		// list that this list entry belongs to.
		if _, err := p.timelineStatus(
			ctx,
			p.state.Timelines.List.IngestOne,
			listEntry.ListID, // list timelines are keyed by list ID
			follow.Account,
			status,
			stream.TimelineList+":"+listEntry.ListID, // key streamType to this specific list
		); err != nil {
			errs.Appendf("error list timelining status: %w", err)
		}
	}
}

// listEligible checks if the given status is eligible
// for inclusion in the list that that the given listEntry
// belongs to, based on the replies policy of the list.
func (p *Processor) listEligible(
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
	list, err := p.state.DB.GetListByID(
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
		includes, err := p.state.DB.ListIncludesAccount(
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
		follows, err := p.state.DB.IsFollowing(
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
func (p *Processor) timelineStatus(
	ctx context.Context,
	ingest func(context.Context, string, timeline.Timelineable) (bool, error),
	timelineID string,
	account *gtsmodel.Account,
	status *gtsmodel.Status,
	streamType string,
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
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, account)
	if err != nil {
		err = gtserror.Newf("error converting status %s to frontend representation: %w", status.ID, err)
		return true, err
	}

	if err := p.stream.Update(apiStatus, account, []string{streamType}); err != nil {
		err = gtserror.Newf("error streaming update for status %s: %w", status.ID, err)
		return true, err
	}

	return true, nil
}

// wipeStatus contains common logic used to totally delete a status
// + all its attachments, notifications, boosts, and timeline entries.
func (p *Processor) wipeStatus(ctx context.Context, statusToDelete *gtsmodel.Status, deleteAttachments bool) error {
	var errs gtserror.MultiError

	// Either delete all attachments for this status,
	// or simply unattach + clean them separately later.
	//
	// Reason to unattach rather than delete is that
	// the poster might want to reattach them to another
	// status immediately (in case of delete + redraft)
	if deleteAttachments {
		// todo: p.state.DB.DeleteAttachmentsForStatus
		for _, a := range statusToDelete.AttachmentIDs {
			if err := p.media.Delete(ctx, a); err != nil {
				errs.Appendf("error deleting media: %w", err)
			}
		}
	} else {
		// todo: p.state.DB.UnattachAttachmentsForStatus
		for _, a := range statusToDelete.AttachmentIDs {
			if _, err := p.media.Unattach(ctx, statusToDelete.Account, a); err != nil {
				errs.Appendf("error unattaching media: %w", err)
			}
		}
	}

	// delete all mention entries generated by this status
	// todo: p.state.DB.DeleteMentionsForStatus
	for _, id := range statusToDelete.MentionIDs {
		if err := p.state.DB.DeleteMentionByID(ctx, id); err != nil {
			errs.Appendf("error deleting status mention: %w", err)
		}
	}

	// delete all notification entries generated by this status
	if err := p.state.DB.DeleteNotificationsForStatus(ctx, statusToDelete.ID); err != nil {
		errs.Appendf("error deleting status notifications: %w", err)
	}

	// delete all bookmarks that point to this status
	if err := p.state.DB.DeleteStatusBookmarksForStatus(ctx, statusToDelete.ID); err != nil {
		errs.Appendf("error deleting status bookmarks: %w", err)
	}

	// delete all faves of this status
	if err := p.state.DB.DeleteStatusFavesForStatus(ctx, statusToDelete.ID); err != nil {
		errs.Appendf("error deleting status faves: %w", err)
	}

	// delete all boosts for this status + remove them from timelines
	boosts, err := p.state.DB.GetStatusBoosts(
		// we MUST set a barebones context here,
		// as depending on where it came from the
		// original BoostOf may already be gone.
		gtscontext.SetBarebones(ctx),
		statusToDelete.ID)
	if err != nil {
		errs.Appendf("error fetching status boosts: %w", err)
	}
	for _, b := range boosts {
		if err := p.deleteStatusFromTimelines(ctx, b.ID); err != nil {
			errs.Appendf("error deleting boost from timelines: %w", err)
		}
		if err := p.state.DB.DeleteStatusByID(ctx, b.ID); err != nil {
			errs.Appendf("error deleting boost: %w", err)
		}
	}

	// delete this status from any and all timelines
	if err := p.deleteStatusFromTimelines(ctx, statusToDelete.ID); err != nil {
		errs.Appendf("error deleting status from timelines: %w", err)
	}

	// finally, delete the status itself
	if err := p.state.DB.DeleteStatusByID(ctx, statusToDelete.ID); err != nil {
		errs.Appendf("error deleting status: %w", err)
	}

	return errs.Combine()
}

// deleteStatusFromTimelines completely removes the given status from all timelines.
// It will also stream deletion of the status to all open streams.
func (p *Processor) deleteStatusFromTimelines(ctx context.Context, statusID string) error {
	if err := p.state.Timelines.Home.WipeItemFromAllTimelines(ctx, statusID); err != nil {
		return err
	}

	if err := p.state.Timelines.List.WipeItemFromAllTimelines(ctx, statusID); err != nil {
		return err
	}

	return p.stream.Delete(statusID)
}

// invalidateStatusFromTimelines does cache invalidation on the given status by
// unpreparing it from all timelines, forcing it to be prepared again (with updated
// stats, boost counts, etc) next time it's fetched by the timeline owner. This goes
// both for the status itself, and for any boosts of the status.
func (p *Processor) invalidateStatusFromTimelines(ctx context.Context, statusID string) {
	if err := p.state.Timelines.Home.UnprepareItemFromAllTimelines(ctx, statusID); err != nil {
		log.
			WithContext(ctx).
			WithField("statusID", statusID).
			Errorf("error unpreparing status from home timelines: %v", err)
	}

	if err := p.state.Timelines.List.UnprepareItemFromAllTimelines(ctx, statusID); err != nil {
		log.
			WithContext(ctx).
			WithField("statusID", statusID).
			Errorf("error unpreparing status from list timelines: %v", err)
	}
}
