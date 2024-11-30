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

package gtsmodel

import (
	"time"
)

// WebPushSubscription represents an access token's Web Push subscription.
// There can be at most one per access token.
type WebPushSubscription struct {
	// ID of this subscription in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// CreatedAt is the time this subscription was created.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// UpdatedAt is the time this subscription was last updated.
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// AccountID of the local account that created this subscription.
	AccountID string `bun:"type:CHAR(26),notnull,nullzero"`

	// TokenID is the ID of the associated access token.
	// There can be at most one subscription for any given access token,
	TokenID string `bun:"type:CHAR(26),nullzero,notnull,unique"`

	// Endpoint is the URL receiving Web Push notifications for this subscription.
	Endpoint string `bun:",nullzero,notnull"`

	// Auth is a Base64-encoded authentication secret.
	Auth string `bun:",nullzero,notnull"`

	// P256dh is a Base64-encoded Diffie-Hellman public key on the P-256 elliptic curve.
	P256dh string `bun:",nullzero,notnull"`

	// NotifyFollow and friends control which notifications are delivered to a given subscription.
	// Corresponds to NotificationType and model.PushSubscriptionAlerts.
	NotifyFollow        *bool `bun:",nullzero,notnull,default:false"`
	NotifyFollowRequest *bool `bun:",nullzero,notnull,default:false"`
	NotifyFavourite     *bool `bun:",nullzero,notnull,default:false"`
	NotifyMention       *bool `bun:",nullzero,notnull,default:false"`
	NotifyReblog        *bool `bun:",nullzero,notnull,default:false"`
	NotifyPoll          *bool `bun:",nullzero,notnull,default:false"`
	NotifyStatus        *bool `bun:",nullzero,notnull,default:false"`
	NotifyUpdate        *bool `bun:",nullzero,notnull,default:false"`
	NotifyAdminSignup   *bool `bun:",nullzero,notnull,default:false"`
	NotifyAdminReport   *bool `bun:",nullzero,notnull,default:false"`
	NotifyPendingFave   *bool `bun:",nullzero,notnull,default:false"`
	NotifyPendingReply  *bool `bun:",nullzero,notnull,default:false"`
	NotifyPendingReblog *bool `bun:",nullzero,notnull,default:false"`
}
