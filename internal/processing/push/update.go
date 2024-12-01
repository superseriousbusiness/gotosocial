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

package push

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// Update updates the Web Push subscription for the given access token.
func (p *Processor) Update(
	ctx context.Context,
	accessToken string,
	request *apimodel.WebPushSubscriptionUpdateRequest,
) (*apimodel.WebPushSubscription, gtserror.WithCode) {
	tokenID, errWithCode := p.getTokenID(ctx, accessToken)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get existing subscription.
	subscription, err := p.state.DB.GetWebPushSubscriptionByTokenID(ctx, tokenID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("couldn't get Web Push subscription for token ID %s: %w", tokenID, err),
		)
	}
	if subscription == nil {
		return nil, gtserror.NewErrorNotFound(errors.New("no Web Push subscription exists for this access token"))
	}

	// Update it.
	subscription.NotifyFollow = &request.Data.Alerts.Follow
	subscription.NotifyFollowRequest = &request.Data.Alerts.FollowRequest
	subscription.NotifyFavourite = &request.Data.Alerts.Favourite
	subscription.NotifyMention = &request.Data.Alerts.Mention
	subscription.NotifyReblog = &request.Data.Alerts.Reblog
	subscription.NotifyPoll = &request.Data.Alerts.Poll
	subscription.NotifyStatus = &request.Data.Alerts.Status
	subscription.NotifyUpdate = &request.Data.Alerts.Update
	subscription.NotifyAdminSignup = &request.Data.Alerts.AdminSignup
	subscription.NotifyAdminReport = &request.Data.Alerts.AdminReport
	subscription.NotifyPendingFavourite = &request.Data.Alerts.PendingFavourite
	subscription.NotifyPendingReply = &request.Data.Alerts.PendingReply
	subscription.NotifyPendingReblog = &request.Data.Alerts.PendingReblog
	if err = p.state.DB.UpdateWebPushSubscription(
		ctx,
		subscription,
		"notify_follow",
		"notify_follow_request",
		"notify_favourite",
		"notify_mention",
		"notify_reblog",
		"notify_poll",
		"notify_status",
		"notify_update",
		"notify_admin_signup",
		"notify_admin_report",
		"notify_pending_favourite",
		"notify_pending_reply",
		"notify_pending_reblog",
	); err != nil {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("couldn't update Web Push subscription for token ID %s: %w", tokenID, err),
		)
	}

	return p.apiSubscription(ctx, subscription)
}
