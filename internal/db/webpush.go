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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// WebPush contains functions related to Web Push notifications.
type WebPush interface {
	// GetVAPIDKeyPair retrieves the server's existing VAPID key pair, if there is one.
	// If there isn't, it returns nil.
	GetVAPIDKeyPair(ctx context.Context) (*gtsmodel.VAPIDKeyPair, error)

	// PutVAPIDKeyPair stores the server's VAPID key pair.
	// This should be called at most once, during server startup.
	PutVAPIDKeyPair(ctx context.Context, vapidKeyPair *gtsmodel.VAPIDKeyPair) error

	// GetWebPushSubscriptionByTokenID retrieves an access token's Web Push subscription, if there is one.
	GetWebPushSubscriptionByTokenID(ctx context.Context, tokenID string) (*gtsmodel.WebPushSubscription, error)

	// PutWebPushSubscription creates an access token's Web Push subscription.
	PutWebPushSubscription(ctx context.Context, subscription *gtsmodel.WebPushSubscription) error

	// UpdateWebPushSubscription updates an access token's Web Push subscription.
	UpdateWebPushSubscription(ctx context.Context, subscription *gtsmodel.WebPushSubscription, columns ...string) error

	// DeleteWebPushSubscriptionByTokenID deletes an access token's Web Push subscription, if there is one.
	DeleteWebPushSubscriptionByTokenID(ctx context.Context, tokenID string) error

	// GetWebPushSubscriptionsByAccountID retrieves an account's list of Web Push subscriptions.
	GetWebPushSubscriptionsByAccountID(ctx context.Context, accountID string) ([]*gtsmodel.WebPushSubscription, error)

	// DeleteWebPushSubscriptionsByAccountID deletes an account's list of Web Push subscriptions.
	DeleteWebPushSubscriptionsByAccountID(ctx context.Context, accountID string) error
}
