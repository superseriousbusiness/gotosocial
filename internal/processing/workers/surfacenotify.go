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

// notifyMentions notifies each targeted account in
// the given mentions that they have a new mention.
func (s *surface) notifyMentions(
	ctx context.Context,
	mentions []*gtsmodel.Mention,
) error {
	errs := gtserror.NewMultiError(len(mentions))

	for _, mention := range mentions {
		if err := s.notify(
			ctx,
			gtsmodel.NotificationMention,
			mention.TargetAccountID,
			mention.OriginAccountID,
			mention.StatusID,
		); err != nil {
			errs.Append(err)
		}
	}

	return errs.Combine()
}

// notifyFollowRequest notifies the target of the given
// follow request that they have a new follow request.
func (s *surface) notifyFollowRequest(
	ctx context.Context,
	followRequest *gtsmodel.FollowRequest,
) error {
	return s.notify(
		ctx,
		gtsmodel.NotificationFollowRequest,
		followRequest.TargetAccountID,
		followRequest.AccountID,
		"",
	)
}

// notifyFollow notifies the target of the given follow that
// they have a new follow. It will also remove any previous
// notification of a follow request, essentially replacing
// that notification.
func (s *surface) notifyFollow(
	ctx context.Context,
	follow *gtsmodel.Follow,
) error {
	// Check if previous follow req notif exists.
	prevNotif, err := s.state.DB.GetNotification(
		gtscontext.SetBarebones(ctx),
		gtsmodel.NotificationFollowRequest,
		follow.TargetAccountID,
		follow.AccountID,
		"",
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error checking for previous follow request notification: %w", err)
	}

	if prevNotif != nil {
		// Previous notif existed, delete it.
		if err := s.state.DB.DeleteNotificationByID(ctx, prevNotif.ID); err != nil {
			return gtserror.Newf("db error removing previous follow request notification %s: %w", prevNotif.ID, err)
		}
	}

	// Now notify the follow itself.
	return s.notify(
		ctx,
		gtsmodel.NotificationFollow,
		follow.TargetAccountID,
		follow.AccountID,
		"",
	)
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

	return s.notify(
		ctx,
		gtsmodel.NotificationFave,
		fave.TargetAccountID,
		fave.AccountID,
		fave.StatusID,
	)
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

	return s.notify(
		ctx,
		gtsmodel.NotificationReblog,
		status.BoostOfAccountID,
		status.AccountID,
		status.ID,
	)
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
	targetAccountID string,
	originAccountID string,
	statusID string,
) error {
	targetAccount, err := s.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		return gtserror.Newf("error getting target account %s: %w", targetAccountID, err)
	}

	if !targetAccount.IsLocal() {
		// Nothing to do.
		return nil
	}

	// Make sure a notification doesn't
	// already exist with these params.
	if _, err := s.state.DB.GetNotification(
		gtscontext.SetBarebones(ctx),
		notificationType,
		targetAccountID,
		originAccountID,
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
		TargetAccountID:  targetAccountID,
		OriginAccountID:  originAccountID,
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
