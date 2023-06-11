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

package processing

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
)

// timelineAndNotifyStatus processes the given new status and inserts it into
// the HOME and LIST timelines of accounts that follow the status author.
//
// It will also handle notifications for any mentions attached to the account, and
// also notifications for any local accounts that want to know when this account posts.
func (p *Processor) timelineAndNotifyStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Ensure status fully populated; including account, mentions, etc.
	if err := p.state.DB.PopulateStatus(ctx, status); err != nil {
		return fmt.Errorf("timelineAndNotifyStatus: error populating status with id %s: %w", status.ID, err)
	}

	// Get local followers of the account that posted the status.
	follows, err := p.state.DB.GetAccountLocalFollowers(ctx, status.AccountID)
	if err != nil {
		return fmt.Errorf("timelineAndNotifyStatus: error getting local followers for account id %s: %w", status.AccountID, err)
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
		return fmt.Errorf("timelineAndNotifyStatus: error timelining status %s for followers: %w", status.ID, err)
	}

	// Notify each local account that's mentioned by this status.
	if err := p.notifyStatusMentions(ctx, status); err != nil {
		return fmt.Errorf("timelineAndNotifyStatus: error notifying status mentions for status %s: %w", status.ID, err)
	}

	return nil
}

func (p *Processor) timelineAndNotifyStatusForFollowers(ctx context.Context, status *gtsmodel.Status, follows []*gtsmodel.Follow) error {
	var (
		errs  = make(gtserror.MultiError, 0, len(follows))
		boost = status.BoostOfID != ""
		reply = status.InReplyToURI != ""
	)

	for _, follow := range follows {
		if sr := follow.ShowReblogs; boost && (sr == nil || !*sr) {
			// This is a boost, but this follower
			// doesn't want to see those from this
			// account, so just skip everything.
			continue
		}

		// Add status to each list that this follow
		// is included in, and stream it if applicable.
		listEntries, err := p.state.DB.GetListEntriesForFollowID(
			// We only need the list IDs.
			gtscontext.SetBarebones(ctx),
			follow.ID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Append(fmt.Errorf("timelineAndNotifyStatusForFollowers: error list timelining status: %w", err))
			continue
		}

		for _, listEntry := range listEntries {
			if _, err := p.timelineStatus(
				ctx,
				p.state.Timelines.List.IngestOne,
				listEntry.ListID, // list timelines are keyed by list ID
				follow.Account,
				status,
				stream.TimelineList+":"+listEntry.ListID, // key streamType to this specific list
			); err != nil {
				errs.Append(fmt.Errorf("timelineAndNotifyStatusForFollowers: error list timelining status: %w", err))
				continue
			}
		}

		// Add status to home timeline for this
		// follower, and stream it if applicable.
		if timelined, err := p.timelineStatus(
			ctx,
			p.state.Timelines.Home.IngestOne,
			follow.AccountID, // home timelines are keyed by account ID
			follow.Account,
			status,
			stream.TimelineHome,
		); err != nil {
			errs.Append(fmt.Errorf("timelineAndNotifyStatusForFollowers: error home timelining status: %w", err))
			continue
		} else if !timelined {
			// Status wasn't added to home tomeline,
			// so we shouldn't notify it either.
			continue
		}

		if n := follow.Notify; n == nil || !*n {
			// This follower doesn't have notifications
			// set for this account's new posts, so bail.
			continue
		}

		if boost || reply {
			// Don't notify for boosts or replies.
			continue
		}

		// If we reach here, we know:
		//
		//   - This follower wants to be notified when this account posts.
		//   - This is a top-level post (not a reply).
		//   - This is not a boost of another post.
		//   - The post is visible in this follower's home timeline.
		//
		// That means we can officially notify this one.
		if err := p.notify(
			ctx,
			gtsmodel.NotificationStatus,
			follow.AccountID,
			status.AccountID,
			status.ID,
		); err != nil {
			errs.Append(fmt.Errorf("timelineAndNotifyStatusForFollowers: error notifying account %s about new status: %w", follow.AccountID, err))
		}
	}

	return errs.Combine()
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
	// Make sure the status is timelineable.
	// This works for both home and list timelines.
	if timelineable, err := p.filter.StatusHomeTimelineable(ctx, account, status); err != nil {
		err = fmt.Errorf("timelineStatusForAccount: error getting timelineability for status for timeline with id %s: %w", account.ID, err)
		return false, err
	} else if !timelineable {
		// Nothing to do.
		return false, nil
	}

	// Ingest status into given timeline using provided function.
	if inserted, err := ingest(ctx, timelineID, status); err != nil {
		err = fmt.Errorf("timelineStatusForAccount: error ingesting status %s: %w", status.ID, err)
		return false, err
	} else if !inserted {
		// Nothing more to do.
		return false, nil
	}

	// The status was inserted so stream it to the user.
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, account)
	if err != nil {
		err = fmt.Errorf("timelineStatusForAccount: error converting status %s to frontend representation: %w", status.ID, err)
		return true, err
	}

	if err := p.stream.Update(apiStatus, account, []string{streamType}); err != nil {
		err = fmt.Errorf("timelineStatusForAccount: error streaming update for status %s: %w", status.ID, err)
		return true, err
	}

	return true, nil
}

func (p *Processor) notifyStatusMentions(ctx context.Context, status *gtsmodel.Status) error {
	errs := make(gtserror.MultiError, 0, len(status.Mentions))

	for _, m := range status.Mentions {
		if err := p.notify(
			ctx,
			gtsmodel.NotificationMention,
			m.TargetAccountID,
			m.OriginAccountID,
			m.StatusID,
		); err != nil {
			errs.Append(err)
		}
	}

	return errs.Combine()
}

func (p *Processor) notifyFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest) error {
	return p.notify(
		ctx,
		gtsmodel.NotificationFollowRequest,
		followRequest.TargetAccountID,
		followRequest.AccountID,
		"",
	)
}

func (p *Processor) notifyFollow(ctx context.Context, follow *gtsmodel.Follow, targetAccount *gtsmodel.Account) error {
	// Remove previous follow request notification, if it exists.
	prevNotif, err := p.state.DB.GetNotification(
		gtscontext.SetBarebones(ctx),
		gtsmodel.NotificationFollowRequest,
		targetAccount.ID,
		follow.AccountID,
		"",
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Proper error while checking.
		return fmt.Errorf("notifyFollow: db error checking for previous follow request notification: %w", err)
	}

	if prevNotif != nil {
		// Previous notification existed, delete.
		if err := p.state.DB.DeleteNotificationByID(ctx, prevNotif.ID); err != nil {
			return fmt.Errorf("notifyFollow: db error removing previous follow request notification %s: %w", prevNotif.ID, err)
		}
	}

	// Now notify the follow itself.
	return p.notify(
		ctx,
		gtsmodel.NotificationFollow,
		targetAccount.ID,
		follow.AccountID,
		"",
	)
}

func (p *Processor) notifyFave(ctx context.Context, fave *gtsmodel.StatusFave) error {
	if fave.TargetAccountID == fave.AccountID {
		// Self-fave, nothing to do.
		return nil
	}

	return p.notify(
		ctx,
		gtsmodel.NotificationFave,
		fave.TargetAccountID,
		fave.AccountID,
		fave.StatusID,
	)
}

func (p *Processor) notifyAnnounce(ctx context.Context, status *gtsmodel.Status) error {
	if status.BoostOfID == "" {
		// Not a boost, nothing to do.
		return nil
	}

	if status.BoostOfAccountID == status.AccountID {
		// Self-boost, nothing to do.
		return nil
	}

	return p.notify(
		ctx,
		gtsmodel.NotificationReblog,
		status.BoostOfAccountID,
		status.AccountID,
		status.ID,
	)
}

func (p *Processor) notify(
	ctx context.Context,
	notificationType gtsmodel.NotificationType,
	targetAccountID string,
	originAccountID string,
	statusID string,
) error {
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		return fmt.Errorf("notify: error getting target account %s: %w", targetAccountID, err)
	}

	if !targetAccount.IsLocal() {
		// Nothing to do.
		return nil
	}

	// Make sure a notification doesn't
	// already exist with these params.
	if _, err := p.state.DB.GetNotification(
		ctx,
		notificationType,
		targetAccountID,
		originAccountID,
		statusID,
	); err == nil {
		// Notification exists, nothing to do.
		return nil
	} else if !errors.Is(err, db.ErrNoEntries) {
		// Real error.
		return fmt.Errorf("notify: error checking existence of notification: %w", err)
	}

	// Notification doesn't yet exist, so
	// we need to create + store one.
	notif := &gtsmodel.Notification{
		ID:               id.NewULID(),
		NotificationType: notificationType,
		TargetAccountID:  targetAccountID,
		OriginAccountID:  originAccountID,
		StatusID:         statusID,
	}

	if err := p.state.DB.PutNotification(ctx, notif); err != nil {
		return fmt.Errorf("notify: error putting notification in database: %w", err)
	}

	// Stream notification to the user.
	apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return fmt.Errorf("notify: error converting notification to api representation: %w", err)
	}

	if err := p.stream.Notify(apiNotif, targetAccount); err != nil {
		return fmt.Errorf("notify: error streaming notification to account: %w", err)
	}

	return nil
}

// wipeStatus contains common logic used to totally delete a status
// + all its attachments, notifications, boosts, and timeline entries.
func (p *Processor) wipeStatus(ctx context.Context, statusToDelete *gtsmodel.Status, deleteAttachments bool) error {
	// either delete all attachments for this status, or simply
	// unattach all attachments for this status, so they'll be
	// cleaned later by a separate process; reason to unattach rather
	// than delete is that the poster might want to reattach them
	// to another status immediately (in case of delete + redraft)
	if deleteAttachments {
		// todo: p.state.DB.DeleteAttachmentsForStatus
		for _, a := range statusToDelete.AttachmentIDs {
			if err := p.media.Delete(ctx, a); err != nil {
				return err
			}
		}
	} else {
		// todo: p.state.DB.UnattachAttachmentsForStatus
		for _, a := range statusToDelete.AttachmentIDs {
			if _, err := p.media.Unattach(ctx, statusToDelete.Account, a); err != nil {
				return err
			}
		}
	}

	// delete all mention entries generated by this status
	// todo: p.state.DB.DeleteMentionsForStatus
	for _, id := range statusToDelete.MentionIDs {
		if err := p.state.DB.DeleteMentionByID(ctx, id); err != nil {
			return err
		}
	}

	// delete all notification entries generated by this status
	if err := p.state.DB.DeleteNotificationsForStatus(ctx, statusToDelete.ID); err != nil {
		return err
	}

	// delete all bookmarks that point to this status
	if err := p.state.DB.DeleteStatusBookmarksForStatus(ctx, statusToDelete.ID); err != nil {
		return err
	}

	// delete all faves of this status
	if err := p.state.DB.DeleteStatusFavesForStatus(ctx, statusToDelete.ID); err != nil {
		return err
	}

	// delete all boosts for this status + remove them from timelines
	if boosts, err := p.state.DB.GetStatusReblogs(ctx, statusToDelete); err == nil {
		for _, b := range boosts {
			if err := p.deleteStatusFromTimelines(ctx, b.ID); err != nil {
				return err
			}
			if err := p.state.DB.DeleteStatusByID(ctx, b.ID); err != nil {
				return err
			}
		}
	}

	// delete this status from any and all timelines
	if err := p.deleteStatusFromTimelines(ctx, statusToDelete.ID); err != nil {
		return err
	}

	// delete the status itself
	return p.state.DB.DeleteStatusByID(ctx, statusToDelete.ID)
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

/*
	EMAIL FUNCTIONS
*/

func (p *Processor) emailReport(ctx context.Context, report *gtsmodel.Report) error {
	instance, err := p.state.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return fmt.Errorf("emailReport: error getting instance: %w", err)
	}

	toAddresses, err := p.state.DB.GetInstanceModeratorAddresses(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No registered moderator addresses.
			return nil
		}
		return fmt.Errorf("emailReport: error getting instance moderator addresses: %w", err)
	}

	if report.Account == nil {
		report.Account, err = p.state.DB.GetAccountByID(ctx, report.AccountID)
		if err != nil {
			return fmt.Errorf("emailReport: error getting report account: %w", err)
		}
	}

	if report.TargetAccount == nil {
		report.TargetAccount, err = p.state.DB.GetAccountByID(ctx, report.TargetAccountID)
		if err != nil {
			return fmt.Errorf("emailReport: error getting report target account: %w", err)
		}
	}

	reportData := email.NewReportData{
		InstanceURL:        instance.URI,
		InstanceName:       instance.Title,
		ReportURL:          instance.URI + "/settings/admin/reports/" + report.ID,
		ReportDomain:       report.Account.Domain,
		ReportTargetDomain: report.TargetAccount.Domain,
	}

	if err := p.emailSender.SendNewReportEmail(toAddresses, reportData); err != nil {
		return fmt.Errorf("emailReport: error emailing instance moderators: %w", err)
	}

	return nil
}

func (p *Processor) emailReportClosed(ctx context.Context, report *gtsmodel.Report) error {
	user, err := p.state.DB.GetUserByAccountID(ctx, report.Account.ID)
	if err != nil {
		return fmt.Errorf("emailReportClosed: db error getting user: %w", err)
	}

	if user.ConfirmedAt.IsZero() || !*user.Approved || *user.Disabled || user.Email == "" {
		// Only email users who:
		// - are confirmed
		// - are approved
		// - are not disabled
		// - have an email address
		return nil
	}

	instance, err := p.state.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return fmt.Errorf("emailReportClosed: db error getting instance: %w", err)
	}

	if report.Account == nil {
		report.Account, err = p.state.DB.GetAccountByID(ctx, report.AccountID)
		if err != nil {
			return fmt.Errorf("emailReportClosed: error getting report account: %w", err)
		}
	}

	if report.TargetAccount == nil {
		report.TargetAccount, err = p.state.DB.GetAccountByID(ctx, report.TargetAccountID)
		if err != nil {
			return fmt.Errorf("emailReportClosed: error getting report target account: %w", err)
		}
	}

	reportClosedData := email.ReportClosedData{
		Username:             report.Account.Username,
		InstanceURL:          instance.URI,
		InstanceName:         instance.Title,
		ReportTargetUsername: report.TargetAccount.Username,
		ReportTargetDomain:   report.TargetAccount.Domain,
		ActionTakenComment:   report.ActionTaken,
	}

	return p.emailSender.SendReportClosedEmail(user.Email, reportClosedData)
}
