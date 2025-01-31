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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// CreateOrReplace creates a Web Push subscription for the given access token,
// or entirely replaces the previously existing subscription for that token.
func (p *Processor) CreateOrReplace(
	ctx context.Context,
	accountID string,
	accessToken string,
	request *apimodel.WebPushSubscriptionCreateRequest,
) (*apimodel.WebPushSubscription, gtserror.WithCode) {
	tokenID, errWithCode := p.getTokenID(ctx, accessToken)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Clear any previous subscription.
	if err := p.state.DB.DeleteWebPushSubscriptionByTokenID(ctx, tokenID); err != nil {
		err := gtserror.Newf("couldn't delete Web Push subscription for token ID %s: %w", tokenID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Insert a new one.
	subscription := &gtsmodel.WebPushSubscription{
		ID:                id.NewULID(),
		AccountID:         accountID,
		TokenID:           tokenID,
		Endpoint:          request.Subscription.Endpoint,
		Auth:              request.Subscription.Keys.Auth,
		P256dh:            request.Subscription.Keys.P256dh,
		NotificationFlags: alertsToNotificationFlags(request.Data.Alerts),
		Policy:            typeutils.APIWebPushNotificationPolicyToWebPushNotificationPolicy(*request.Data.Policy),
	}

	if err := p.state.DB.PutWebPushSubscription(ctx, subscription); err != nil {
		err := gtserror.Newf("couldn't create Web Push subscription for token ID %s: %w", tokenID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiSubscription(ctx, subscription)
}
