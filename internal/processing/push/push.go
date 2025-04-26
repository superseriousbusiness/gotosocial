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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

type Processor struct {
	state     *state.State
	converter *typeutils.Converter
}

func New(state *state.State, converter *typeutils.Converter) Processor {
	return Processor{
		state:     state,
		converter: converter,
	}
}

// getTokenID returns the token ID for a given access token.
// Since all push API calls require authentication, this should always be available.
func (p *Processor) getTokenID(ctx context.Context, accessToken string) (string, gtserror.WithCode) {
	token, err := p.state.DB.GetTokenByAccess(ctx, accessToken)
	if err != nil {
		err := gtserror.Newf("couldn't find token ID for access token: %w", err)
		return "", gtserror.NewErrorInternalError(err)
	}

	return token.ID, nil
}

// apiSubscription is a shortcut to return the API version of the given Web Push subscription,
// or return an appropriate error if conversion fails.
func (p *Processor) apiSubscription(ctx context.Context, subscription *gtsmodel.WebPushSubscription) (*apimodel.WebPushSubscription, gtserror.WithCode) {
	apiSubscription, err := p.converter.WebPushSubscriptionToAPIWebPushSubscription(ctx, subscription)
	if err != nil {
		err := gtserror.Newf("error converting Web Push subscription %s to API representation: %w", subscription.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiSubscription, nil
}

// alertsToNotificationFlags turns the alerts section of a push subscription API request into a packed bitfield.
func alertsToNotificationFlags(alerts *apimodel.WebPushSubscriptionAlerts) gtsmodel.WebPushSubscriptionNotificationFlags {
	var n gtsmodel.WebPushSubscriptionNotificationFlags

	n.Set(gtsmodel.NotificationFollow, alerts.Follow)
	n.Set(gtsmodel.NotificationFollowRequest, alerts.FollowRequest)
	n.Set(gtsmodel.NotificationFavourite, alerts.Favourite)
	n.Set(gtsmodel.NotificationMention, alerts.Mention)
	n.Set(gtsmodel.NotificationReblog, alerts.Reblog)
	n.Set(gtsmodel.NotificationPoll, alerts.Poll)
	n.Set(gtsmodel.NotificationStatus, alerts.Status)
	n.Set(gtsmodel.NotificationUpdate, alerts.Update)
	n.Set(gtsmodel.NotificationAdminSignup, alerts.AdminSignup)
	n.Set(gtsmodel.NotificationAdminReport, alerts.AdminReport)
	n.Set(gtsmodel.NotificationPendingFave, alerts.PendingFavourite)
	n.Set(gtsmodel.NotificationPendingReply, alerts.PendingReply)
	n.Set(gtsmodel.NotificationPendingReblog, alerts.PendingReblog)

	return n
}
