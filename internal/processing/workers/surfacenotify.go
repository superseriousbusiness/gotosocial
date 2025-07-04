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
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
)

// notifyPendingReply notifies the account replied-to
// by the given status that they have a new reply,
// and that approval is pending.
func (s *Surface) notifyPendingReply(
	ctx context.Context,
	status *gtsmodel.Status,
) error {
	// Beforehand, ensure the passed status is fully populated.
	if err := s.State.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status %s: %w", status.ID, err)
	}

	if status.InReplyToAccount.IsRemote() {
		// Don't notify
		// remote accounts.
		return nil
	}

	if status.AccountID == status.InReplyToAccountID {
		// Don't notify
		// self-replies.
		return nil
	}

	// Ensure thread not muted
	// by replied-to account.
	muted, err := s.State.DB.IsThreadMutedByAccount(ctx,
		status.ThreadID,
		status.InReplyToAccountID,
	)
	if err != nil {
		return gtserror.Newf("error checking status thread mute %s: %w", status.ThreadID, err)
	}

	if muted {
		// The replied-to account
		// has muted the thread.
		// Don't pester them.
		return nil
	}

	// notify mentioned
	// by status author.
	if err := s.Notify(ctx,
		gtsmodel.NotificationPendingReply,
		status.InReplyToAccount,
		status.Account,
		status,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying replied-to account %s: %w", status.InReplyToAccountID, err)
	}

	return nil
}

// notifyMentions iterates through mentions on the
// given status, and notifies each mentioned account
// that they have a new mention.
func (s *Surface) notifyMentions(
	ctx context.Context,
	status *gtsmodel.Status,
) error {
	var errs gtserror.MultiError

	for _, mention := range status.Mentions {
		// Set status on the mention (stops
		// notifyMention having to populate it).
		mention.Status = status

		// Do the thing.
		if err := s.notifyMention(ctx, mention); err != nil {
			errs = append(errs, err)
		}
	}

	return errs.Combine()
}

// notifyMention notifies the target
// of the given mention that they've
// been mentioned in a status.
func (s *Surface) notifyMention(
	ctx context.Context,
	mention *gtsmodel.Mention,
) error {
	// Beforehand, ensure the passed mention is fully populated.
	if err := s.State.DB.PopulateMention(ctx, mention); err != nil {
		return gtserror.Newf(
			"error populating mention %s: %w",
			mention.ID, err,
		)
	}

	if mention.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Ensure thread not muted
	// by mentioned account.
	muted, err := s.State.DB.IsThreadMutedByAccount(ctx,
		mention.Status.ThreadID,
		mention.TargetAccountID,
	)
	if err != nil {
		return gtserror.Newf(
			"error checking status thread mute %s: %w",
			mention.Status.ThreadID, err,
		)
	}

	if muted {
		// This mentioned account
		// has muted the thread.
		// Don't pester them.
		return nil
	}

	// Notify mentioned
	// by status author.
	if err := s.Notify(ctx,
		gtsmodel.NotificationMention,
		mention.TargetAccount,
		mention.OriginAccount,
		mention.Status,
		nil,
	); err != nil {
		return gtserror.Newf(
			"error notifying mention target %s: %w",
			mention.TargetAccountID, err,
		)
	}

	return nil
}

// notifyFollowRequest notifies the target of the given
// follow request that they have a new follow request.
func (s *Surface) notifyFollowRequest(
	ctx context.Context,
	followReq *gtsmodel.FollowRequest,
) error {
	// Beforehand, ensure the passed follow request is fully populated.
	if err := s.State.DB.PopulateFollowRequest(ctx, followReq); err != nil {
		return gtserror.Newf("error populating follow request %s: %w", followReq.ID, err)
	}

	if followReq.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Now notify the follow request itself.
	if err := s.Notify(ctx,
		gtsmodel.NotificationFollowRequest,
		followReq.TargetAccount,
		followReq.Account,
		nil,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying follow target %s: %w", followReq.TargetAccountID, err)
	}

	return nil
}

// notifyFollow notifies the target of the given follow that
// they have a new follow. It will also remove any previous
// notification of a follow request, essentially replacing
// that notification.
func (s *Surface) notifyFollow(
	ctx context.Context,
	follow *gtsmodel.Follow,
) error {
	// Beforehand, ensure the passed follow is fully populated.
	if err := s.State.DB.PopulateFollow(ctx, follow); err != nil {
		return gtserror.Newf("error populating follow %s: %w", follow.ID, err)
	}

	if follow.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return nil
	}

	// Check if previous follow req notif exists.
	prevNotif, err := s.State.DB.GetNotification(
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
		if err := s.State.DB.DeleteNotificationByID(ctx, prevNotif.ID); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return gtserror.Newf("error deleting notification %s: %w", prevNotif.ID, err)
		}
	}

	// Now notify the follow itself.
	if err := s.Notify(ctx,
		gtsmodel.NotificationFollow,
		follow.TargetAccount,
		follow.Account,
		nil,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying follow target %s: %w", follow.TargetAccountID, err)
	}

	return nil
}

// notifyFave notifies the target of the given
// fave that their status has been liked/faved.
func (s *Surface) notifyFave(
	ctx context.Context,
	fave *gtsmodel.StatusFave,
) error {
	notifyable, err := s.notifyableFave(ctx, fave)
	if err != nil {
		return err
	}

	if !notifyable {
		// Nothing to do.
		return nil
	}

	// notify status author
	// of fave by account.
	if err := s.Notify(ctx,
		gtsmodel.NotificationFavourite,
		fave.TargetAccount,
		fave.Account,
		fave.Status,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying status author %s: %w", fave.TargetAccountID, err)
	}

	return nil
}

// notifyPendingFave notifies the target of the
// given fave that their status has been faved
// and that approval is required.
func (s *Surface) notifyPendingFave(
	ctx context.Context,
	fave *gtsmodel.StatusFave,
) error {
	notifyable, err := s.notifyableFave(ctx, fave)
	if err != nil {
		return err
	}

	if !notifyable {
		// Nothing to do.
		return nil
	}

	// notify status author
	// of fave by account.
	if err := s.Notify(ctx,
		gtsmodel.NotificationPendingFave,
		fave.TargetAccount,
		fave.Account,
		fave.Status,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying status author %s: %w", fave.TargetAccountID, err)
	}

	return nil
}

// notifyableFave checks that the given
// fave should be notified, taking account
// of localness of receiving account, and mutes.
func (s *Surface) notifyableFave(
	ctx context.Context,
	fave *gtsmodel.StatusFave,
) (bool, error) {
	if fave.TargetAccountID == fave.AccountID {
		// Self-fave, nothing to do.
		return false, nil
	}

	// Beforehand, ensure the passed status fave is fully populated.
	if err := s.State.DB.PopulateStatusFave(ctx, fave); err != nil {
		return false, gtserror.Newf("error populating fave %s: %w", fave.ID, err)
	}

	if fave.TargetAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return false, nil
	}

	// Ensure favee hasn't
	// muted the thread.
	muted, err := s.State.DB.IsThreadMutedByAccount(ctx,
		fave.Status.ThreadID,
		fave.TargetAccountID,
	)
	if err != nil {
		return false, gtserror.Newf("error checking status thread mute %s: %w", fave.StatusID, err)
	}

	if muted {
		// Favee doesn't want
		// notifs for this thread.
		return false, nil
	}

	return true, nil
}

// notifyAnnounce notifies the status boost target
// account that their status has been boosted.
func (s *Surface) notifyAnnounce(
	ctx context.Context,
	boost *gtsmodel.Status,
) error {
	notifyable, err := s.notifyableAnnounce(ctx, boost)
	if err != nil {
		return err
	}

	if !notifyable {
		// Nothing to do.
		return nil
	}

	// notify status author
	// of boost by account.
	if err := s.Notify(ctx,
		gtsmodel.NotificationReblog,
		boost.BoostOfAccount,
		boost.Account,
		boost,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying boost target %s: %w", boost.BoostOfAccountID, err)
	}

	return nil
}

// notifyPendingAnnounce notifies the status boost
// target account that their status has been boosted,
// and that the boost requires approval.
func (s *Surface) notifyPendingAnnounce(
	ctx context.Context,
	boost *gtsmodel.Status,
) error {
	notifyable, err := s.notifyableAnnounce(ctx, boost)
	if err != nil {
		return err
	}

	if !notifyable {
		// Nothing to do.
		return nil
	}

	// notify status author
	// of boost by account.
	if err := s.Notify(ctx,
		gtsmodel.NotificationPendingReblog,
		boost.BoostOfAccount,
		boost.Account,
		boost,
		nil,
	); err != nil {
		return gtserror.Newf("error notifying boost target %s: %w", boost.BoostOfAccountID, err)
	}

	return nil
}

// notifyableAnnounce checks that the given
// announce should be notified, taking account
// of localness of receiving account, and mutes.
func (s *Surface) notifyableAnnounce(
	ctx context.Context,
	status *gtsmodel.Status,
) (bool, error) {
	if status.BoostOfID == "" {
		// Not a boost, nothing to do.
		return false, nil
	}

	if status.BoostOfAccountID == status.AccountID {
		// Self-boost, nothing to do.
		return false, nil
	}

	// Beforehand, ensure the passed status is fully populated.
	if err := s.State.DB.PopulateStatus(ctx, status); err != nil {
		return false, gtserror.Newf("error populating status %s: %w", status.ID, err)
	}

	if status.BoostOfAccount.IsRemote() {
		// no need to notify
		// remote accounts.
		return false, nil
	}

	// Ensure boostee hasn't
	// muted the thread.
	muted, err := s.State.DB.IsThreadMutedByAccount(ctx,
		status.BoostOf.ThreadID,
		status.BoostOfAccountID,
	)

	if err != nil {
		return false, gtserror.Newf("error checking status thread mute %s: %w", status.BoostOfID, err)
	}

	if muted {
		// Boostee doesn't want
		// notifs for this thread.
		return false, nil
	}

	return true, nil
}

func (s *Surface) notifyPollClose(ctx context.Context, status *gtsmodel.Status) error {
	// Beforehand, ensure the passed status is fully populated.
	if err := s.State.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status %s: %w", status.ID, err)
	}

	// Fetch all votes in the attached status poll.
	votes, err := s.State.DB.GetPollVotes(ctx, status.PollID)
	if err != nil {
		return gtserror.Newf("error getting poll %s votes: %w", status.PollID, err)
	}

	var errs gtserror.MultiError

	if status.Account.IsLocal() {
		// Send a notification to the status
		// author that their poll has closed!
		if err := s.Notify(ctx,
			gtsmodel.NotificationPoll,
			status.Account,
			status.Account,
			status,
			nil,
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
		if err := s.Notify(ctx,
			gtsmodel.NotificationPoll,
			vote.Account,
			status.Account,
			status,
			nil,
		); err != nil {
			errs.Appendf("error notifying poll voter %s: %w", vote.AccountID, err)
			continue
		}
	}

	return errs.Combine()
}

func (s *Surface) notifySignup(ctx context.Context, newUser *gtsmodel.User) error {
	modAccounts, err := s.State.DB.GetInstanceModerators(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No registered
			// mod accounts.
			return nil
		}

		// Real error.
		return gtserror.Newf("error getting instance moderator accounts: %w", err)
	}

	// Ensure user + account populated.
	if err := s.State.DB.PopulateUser(ctx, newUser); err != nil {
		return gtserror.Newf("db error populating new user: %w", err)
	}

	if err := s.State.DB.PopulateAccount(ctx, newUser.Account); err != nil {
		return gtserror.Newf("db error populating new user's account: %w", err)
	}

	// Notify each moderator.
	var errs gtserror.MultiError
	for _, mod := range modAccounts {
		if err := s.Notify(ctx,
			gtsmodel.NotificationAdminSignup,
			mod,
			newUser.Account,
			nil,
			nil,
		); err != nil {
			errs.Appendf("error notifying moderator %s: %w", mod.ID, err)
			continue
		}
	}

	return errs.Combine()
}

func (s *Surface) notifyStatusEdit(
	ctx context.Context,
	status *gtsmodel.Status,
	edit *gtsmodel.StatusEdit,
) error {
	// Get local-only interactions (we can't/don't notify remotes).
	interactions, err := s.State.DB.GetStatusInteractions(ctx, status.ID, true)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error getting status interactions: %w", err)
	}

	// Deduplicate interactions by account ID,
	// we don't need to notify someone twice
	// if they've both boosted *and* replied
	// to an edited status, for example.
	interactions = xslices.DeduplicateFunc(
		interactions,
		func(v gtsmodel.Interaction) string {
			return v.GetAccount().ID
		},
	)

	// Notify each account that's
	// interacted with the status.
	var errs gtserror.MultiError
	for _, i := range interactions {
		targetAcct := i.GetAccount()
		if targetAcct.ID == status.AccountID {
			// Don't notify an account
			// if they've interacted
			// with their *own* status.
			continue
		}

		if err := s.Notify(ctx,
			gtsmodel.NotificationUpdate,
			targetAcct,
			status.Account,
			status,
			edit,
		); err != nil {
			errs.Appendf("error notifying status edit: %w", err)
			continue
		}
	}

	return errs.Combine()
}

func getNotifyLockURI(
	notificationType gtsmodel.NotificationType,
	targetAccount *gtsmodel.Account,
	originAccount *gtsmodel.Account,
	statusOrEditID string,
) string {
	builder := strings.Builder{}
	builder.WriteString("notification:?")
	builder.WriteString("type=" + notificationType.String())
	builder.WriteString("&targetAcct=" + targetAccount.URI)
	builder.WriteString("&originAcct=" + originAccount.URI)
	if statusOrEditID != "" {
		builder.WriteString("&statusOrEditID=" + statusOrEditID)
	}
	return builder.String()
}

// Notify creates, inserts, and streams a new
// notification to the target account if it
// doesn't yet exist with the given parameters.
//
// It filters out non-local target accounts, so
// it is safe to pass all sorts of notification
// targets into this function without filtering
// for non-local first.
//
// targetAccount and originAccount must be
// set, but statusOrEditID can be empty.
func (s *Surface) Notify(
	ctx context.Context,
	notificationType gtsmodel.NotificationType,
	targetAccount *gtsmodel.Account,
	originAccount *gtsmodel.Account,
	status *gtsmodel.Status,
	edit *gtsmodel.StatusEdit,
) error {
	if targetAccount.IsRemote() {
		// nothing to do.
		return nil
	}

	// Get status / edit ID
	// if either was provided.
	// (prefer edit though!)
	var statusOrEditID string
	if edit != nil {
		statusOrEditID = edit.ID
	} else if status != nil {
		statusOrEditID = status.ID
	}

	// We're doing state-y stuff so get a
	// lock on this combo of notif params.
	unlock := s.State.ProcessingLocks.Lock(getNotifyLockURI(
		notificationType,
		targetAccount,
		originAccount,
		statusOrEditID,
	))

	// Wrap the unlock so we
	// can do granular unlocking.
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Make sure a notification doesn't
	// already exist with these params.
	if _, err := s.State.DB.GetNotification(
		gtscontext.SetBarebones(ctx),
		notificationType,
		targetAccount.ID,
		originAccount.ID,
		statusOrEditID,
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
		StatusOrEditID:   statusOrEditID,
	}

	if err := s.State.DB.PutNotification(ctx, notif); err != nil {
		return gtserror.Newf("error putting notification in database: %w", err)
	}

	// Unlock already, we're done
	// with the state-y stuff.
	unlock()

	// Check whether origin account is muted by target account.
	muted, err := s.MuteFilter.AccountNotificationsMuted(ctx,
		targetAccount,
		originAccount,
	)
	if err != nil {
		return gtserror.Newf("error checking account mute: %w", err)
	}

	if muted {
		// Don't notify.
		return nil
	}

	var filtered []apimodel.FilterResult

	if status != nil {
		// Check whether status is muted by the target account.
		muted, err := s.MuteFilter.StatusNotificationsMuted(ctx,
			targetAccount,
			status,
		)
		if err != nil {
			return gtserror.Newf("error checking status mute: %w", err)
		}

		if muted {
			// Don't notify.
			return nil
		}

		var hide bool

		// Check whether notification status is filtered by requester in notifs.
		filtered, hide, err = s.StatusFilter.StatusFilterResultsInContext(ctx,
			targetAccount,
			status,
			gtsmodel.FilterContextNotifications,
		)
		if err != nil {
			return gtserror.Newf("error checking status filtering: %w", err)
		}

		if hide {
			// Don't notify.
			return nil
		}
	}

	// Convert notification to frontend API model for streaming / web push.
	apiNotif, err := s.Converter.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return gtserror.Newf("error converting notification to api representation: %w", err)
	}

	if apiNotif.Status != nil {
		// Set filter results on status,
		// in case any were set above.
		apiNotif.Status.Filtered = filtered
	}

	// Stream notification to the user.
	s.Stream.Notify(ctx, targetAccount, apiNotif)

	// Send Web Push notification to the user.
	if err = s.WebPushSender.Send(ctx, notif, apiNotif); err != nil {
		return gtserror.Newf("error sending Web Push notifications: %w", err)
	}

	return nil
}
