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
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
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
		err := gtserror.Newf("couldn't get Web Push subscription for token ID %s: %w", tokenID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	if subscription == nil {
		err := errors.New("no Web Push subscription exists for this access token")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Update it.
	subscription.NotificationFlags = alertsToNotificationFlags(request.Data.Alerts)
	subscription.Policy = typeutils.APIWebPushNotificationPolicyToWebPushNotificationPolicy(*request.Data.Policy)

	if err = p.state.DB.UpdateWebPushSubscription(
		ctx,
		subscription,
		"notification_flags",
		"policy",
	); err != nil {
		err := gtserror.Newf("couldn't update Web Push subscription for token ID %s: %w", tokenID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiSubscription(ctx, subscription)
}
