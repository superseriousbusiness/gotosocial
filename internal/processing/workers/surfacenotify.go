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
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// notifyMentions iterates through mentions on the
// given status, and notifies each mentioned account
// that they have a new mention.
func (s *surface) notifyMentions(
	ctx context.Context,
	status *gtsmodel.Status,
) error {
	var errs gtserror.MultiError

	for _, mention := range status.Mentions {
		// Set status on the mention (stops
		// the below function populating it).
		mention.Status = status

		// Beforehand, ensure the passed mention is fully populated.
		if err := s.state.DB.PopulateMention(ctx, mention); err != nil {
			errs.Appendf("error populating mention %s: %w", mention.ID, err)
			continue
		}

		if mention.TargetAccount.IsRemote() {
			// no need to notify
			// remote accounts.
			continue
		}

		// Ensure thread not muted
		// by mentioned account.
		muted, err := s.state.DB.IsThreadMutedByAccount(
			ctx,
			status.ThreadID,
			mention.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error checking status thread mute %s: %w", status.ThreadID, err)
			continue
		}

		if muted {
			// This mentioned account
			// has muted the thread.
			// Don't pester them.
			continue
		}

		// notify mentioned
		// by status author.
		if err := s.notify(ctx,
			gtsmodel.NotificationMention,
			mention.TargetAccount,
			mention.OriginAccount,
			mention.StatusID,
		); err != nil {
			errs.Appendf("error notifying mention target %s: %w", mention.TargetAccountID, err)
			continue
		}
	}

	return errs.Combine()
}

// notifyFollowRequest notifies the target of the given
// follow request that they have a new follow request.
func (s *surface) notifyFollowRequest(
	ctx context.Context,
	followReq *gtsmodel.FollowRequest,
) error {
	// Beforehand, ensure the passed follow request is fully populated.
	if err := s.state.DB.PopulateFollowRequest(ctx, followReq); err != nil {
		return gtserror.Newf("error populating follow request %s: %w", followReq.ID, err)
	}

	if followReq.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Now notify the follow request itself.
	if err := s.notify(ctx,
		gtsmodel.NotificationFollowRequest,
		followReq.TargetAccount,
		followReq.Account,
		"",
	); err != nil {
		return gtserror.Newf("error notifying follow target %s: %w", followReq.TargetAccountID, err)
	}

	return nil
}

// notifyFollow notifies the target of the given follow that
// they have a new follow. It will also remove any previous
// notification of a follow request, essentially replacing
// that notification.
func (s *surface) notifyFollow(
	ctx context.Context,
	follow *gtsmodel.Follow,
) error {
	// Beforehand, ensure the passed follow is fully populated.
	if err := s.state.DB.PopulateFollow(ctx, follow); err != nil {
		return gtserror.Newf("error populating follow %s: %w", follow.ID, err)
	}

	if follow.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Check if previous follow req notif exists.
	prevNotif, err := s.state.DB.GetNotification(
		gtscontext.SetBarebones(ctx),
		gtsmodel.NotificationFollowRequest,
		follow.TargetAccountID,
		follow.AccountID,
		"",
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("error getting notification: %w", err)
	}

	if prevNotif != nil {
		// Previous follow request notif existed, delete it before creating new.
		if err := s.state.DB.DeleteNotificationByID(ctx, prevNotif.ID); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return gtserror.Newf("error deleting notification %s: %w", prevNotif.ID, err)
		}
	}

	// Now notify the follow itself.
	if err := s.notify(ctx,
		gtsmodel.NotificationFollow,
		follow.TargetAccount,
		follow.Account,
		"",
	); err != nil {
		return gtserror.Newf("error notifying follow target %s: %w", follow.TargetAccountID, err)
	}

	return nil
}

// notifyFave notifies the target of the given
// fave that their status has been liked/faved.
func (s *surface) notifyFave(
	ctx context.Context,
	fave *gtsmodel.StatusFave,
) error {
	if fave.TargetAccountID == fave.AccountID {
		// Self-fave, nothing to do.
		return nil
	}

	// Beforehand, ensure the passed status fave is fully populated.
	if err := s.state.DB.PopulateStatusFave(ctx, fave); err != nil {
		return gtserror.Newf("error populating fave %s: %w", fave.ID, err)
	}

	if fave.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Ensure favee hasn't
	// muted the thread.
	muted, err := s.state.DB.IsThreadMutedByAccount(
		ctx,
		fave.Status.ThreadID,
		fave.TargetAccountID,
	)
	if err != nil {
		return gtserror.Newf("error checking status thread mute %s: %w", fave.StatusID, err)
	}

	if muted {
		// Favee doesn't want
		// notifs for this thread.
		return nil
	}

	// notify status author
	// of fave by account.
	if err := s.notify(ctx,
		gtsmodel.NotificationFave,
		fave.TargetAccount,
		fave.Account,
		fave.StatusID,
	); err != nil {
		return gtserror.Newf("error notifying status author %s: %w", fave.TargetAccountID, err)
	}

	return nil
}

// notifyAnnounce notifies the status boost target
// account that their status has been boosted.
func (s *surface) notifyAnnounce(
	ctx context.Context,
	status *gtsmodel.Status,
) error {
	if status.BoostOfID == "" {
		// Not a boost, nothing to do.
		return nil
	}

	if status.BoostOfAccountID == status.AccountID {
		// Self-boost, nothing to do.
		return nil
	}

	// Beforehand, ensure the passed status is fully populated.
	if err := s.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status %s: %w", status.ID, err)
	}

	if status.BoostOfAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Ensure boostee hasn't
	// muted the thread.
	muted, err := s.state.DB.IsThreadMutedByAccount(
		ctx,
		status.BoostOf.ThreadID,
		status.BoostOfAccountID,
	)

	if err != nil {
		return gtserror.Newf("error checking status thread mute %s: %w", status.BoostOfID, err)
	}

	if muted {
		// Boostee doesn't want
		// notifs for this thread.
		return nil
	}

	// notify status author
	// of boost by account.
	if err := s.notify(ctx,
		gtsmodel.NotificationReblog,
		status.BoostOfAccount,
		status.Account,
		status.ID,
	); err != nil {
		return gtserror.Newf("error notifying status author %s: %w", status.BoostOfAccountID, err)
	}

	return nil
}

func (s *surface) notifyPollClose(ctx context.Context, status *gtsmodel.Status) error {
	// Beforehand, ensure the passed status is fully populated.
	if err := s.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status %s: %w", status.ID, err)
	}

	// Fetch all votes in the attached status poll.
	votes, err := s.state.DB.GetPollVotes(ctx, status.PollID)
	if err != nil {
		return gtserror.Newf("error getting poll %s votes: %w", status.PollID, err)
	}

	var errs gtserror.MultiError

	if status.Account.IsLocal() {
		// Send a notification to the status
		// author that their poll has closed!
		if err := s.notify(ctx,
			gtsmodel.NotificationPoll,
			status.Account,
			status.Account,
			status.ID,
		); err != nil {
			errs.Appendf("error notifying poll author: %w", err)
		}
	}

	for _, vote := range votes {
		if vote.Account.IsRemote() {
			// no need to notify
			// remote accounts.
			continue
		}

		// notify voter that
		// poll has been closed.
		if err := s.notify(ctx,
			gtsmodel.NotificationMention,
			vote.Account,
			status.Account,
			status.ID,
		); err != nil {
			errs.Appendf("error notifying poll voter %s: %w", vote.AccountID, err)
			continue
		}
	}

	return errs.Combine()
}

// notify creates, inserts, and streams a new
// notification to the target account if it
// doesn't yet exist with the given parameters.
//
// It filters out non-local target accounts, so
// it is safe to pass all sorts of notification
// targets into this function without filtering
// for non-local first.
//
// targetAccountID and originAccountID must be
// set, but statusID can be an empty string.
func (s *surface) notify(
	ctx context.Context,
	notificationType gtsmodel.NotificationType,
	targetAccount *gtsmodel.Account,
	originAccount *gtsmodel.Account,
	statusID string,
) error {
	if targetAccount.IsRemote() {
		// nothing to do.
		return nil
	}

	// Make sure a notification doesn't
	// already exist with these params.
	if _, err := s.state.DB.GetNotification(
		gtscontext.SetBarebones(ctx),
		notificationType,
		targetAccount.ID,
		originAccount.ID,
		statusID,
	); err == nil {
		// Notification exists;
		// nothing to do.
		return nil
	} else if !errors.Is(err, db.ErrNoEntries) {
		// Real error.
		return gtserror.Newf("error checking existence of notification: %w", err)
	}

	// Notification doesn't yet exist, so
	// we need to create + store one.
	notif := &gtsmodel.Notification{
		ID:               id.NewULID(),
		NotificationType: notificationType,
		TargetAccountID:  targetAccount.ID,
		TargetAccount:    targetAccount,
		OriginAccountID:  originAccount.ID,
		OriginAccount:    originAccount,
		StatusID:         statusID,
	}

	if err := s.state.DB.PutNotification(ctx, notif); err != nil {
		return gtserror.Newf("error putting notification in database: %w", err)
	}

	// Stream notification to the user.
	apiNotif, err := s.converter.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return gtserror.Newf("error converting notification to api representation: %w", err)
	}

	if err := s.stream.Notify(apiNotif, targetAccount); err != nil {
		return gtserror.Newf("error streaming notification to account: %w", err)
	}

	return nil
}
