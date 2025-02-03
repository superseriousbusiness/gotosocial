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

// WebPushSubscription represents an access token's Web Push subscription.
// There can be at most one per access token.
type WebPushSubscription struct {
	// ID of this subscription in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero"`

	// AccountID of the local account that created this subscription.
	AccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	// TokenID is the ID of the associated access token.
	// There can be at most one subscription for any given access token,
	TokenID string `bun:"type:CHAR(26),nullzero,notnull,unique"`

	// Endpoint is the URL receiving Web Push notifications for this subscription.
	Endpoint string `bun:",nullzero,notnull"`

	// Auth is a Base64-encoded authentication secret.
	Auth string `bun:",nullzero,notnull"`

	// P256dh is a Base64-encoded Diffie-Hellman public key on the P-256 elliptic curve.
	P256dh string `bun:",nullzero,notnull"`

	// NotificationFlags controls which notifications are delivered to this subscription.
	NotificationFlags WebPushSubscriptionNotificationFlags `bun:",notnull"`

	// Policy controls which accounts are allowed to trigger notifications for this subscription.
	Policy WebPushNotificationPolicy `bun:",nullzero,notnull,default:1"`
}

// WebPushSubscriptionNotificationFlags is a bitfield representation of a set of NotificationType.
// Corresponds to apimodel.WebPushSubscriptionAlerts.
type WebPushSubscriptionNotificationFlags int64

// WebPushSubscriptionNotificationFlagsFromSlice packs a slice of NotificationType into a WebPushSubscriptionNotificationFlags.
func WebPushSubscriptionNotificationFlagsFromSlice(notificationTypes []NotificationType) WebPushSubscriptionNotificationFlags {
	var n WebPushSubscriptionNotificationFlags
	for _, notificationType := range notificationTypes {
		n.Set(notificationType, true)
	}
	return n
}

// ToSlice unpacks a WebPushSubscriptionNotificationFlags into a slice of NotificationType.
func (n *WebPushSubscriptionNotificationFlags) ToSlice() []NotificationType {
	notificationTypes := make([]NotificationType, 0, NotificationTypeNumValues)
	for notificationType := NotificationUnknown; notificationType < NotificationTypeNumValues; notificationType++ {
		if n.Get(notificationType) {
			notificationTypes = append(notificationTypes, notificationType)
		}
	}
	return notificationTypes
}

// Get tests to see if a given NotificationType is included in this set of flags.
func (n *WebPushSubscriptionNotificationFlags) Get(notificationType NotificationType) bool {
	return *n&(1<<notificationType) != 0
}

// Set adds or removes a given NotificationType to or from this set of flags.
func (n *WebPushSubscriptionNotificationFlags) Set(notificationType NotificationType, value bool) {
	if value {
		*n |= 1 << notificationType
	} else {
		*n &= ^(1 << notificationType)
	}
}

// WebPushNotificationPolicy represents the notification policy of a Web Push subscription.
// Corresponds to apimodel.WebPushNotificationPolicy.
type WebPushNotificationPolicy enumType

const (
	// WebPushNotificationPolicyAll allows all accounts to send notifications to the subscribing user.
	WebPushNotificationPolicyAll WebPushNotificationPolicy = 1
	// WebPushNotificationPolicyFollowed allows accounts followed by the subscribing user to send notifications.
	WebPushNotificationPolicyFollowed WebPushNotificationPolicy = 2
	// WebPushNotificationPolicyFollower allows accounts following the subscribing user to send notifications.
	WebPushNotificationPolicyFollower WebPushNotificationPolicy = 3
	// WebPushNotificationPolicyNone doesn't allow any accounts to send notifications to the subscribing user.
	WebPushNotificationPolicyNone WebPushNotificationPolicy = 4
)
