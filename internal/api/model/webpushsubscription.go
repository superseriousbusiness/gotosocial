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

package model

// WebPushSubscription represents a subscription to a Web Push server.
//
// swagger:model webPushSubscription
type WebPushSubscription struct {
	// The id of the push subscription in the database.
	ID string `json:"id"`

	// Where push alerts will be sent to.
	Endpoint string `json:"endpoint"`

	// The streaming server's VAPID public key.
	ServerKey string `json:"server_key"`

	// Which alerts should be delivered to the endpoint.
	Alerts WebPushSubscriptionAlerts `json:"alerts"`

	// Which accounts should generate notifications.
	Policy WebPushNotificationPolicy `json:"policy"`

	// Whether the subscription uses RFC or pre-RFC Web Push standards.
	// For GotoSocial, this is always true.
	Standard bool `json:"standard"`
}

// WebPushSubscriptionAlerts represents the specific events that this Web Push subscription will receive.
//
// swagger:model webPushSubscriptionAlerts
type WebPushSubscriptionAlerts struct {
	// Receive a push notification when someone has followed you?
	Follow bool `json:"follow"`

	// Receive a push notification when someone has requested to follow you?
	FollowRequest bool `json:"follow_request"`

	// Receive a push notification when a status you created has been favourited by someone else?
	Favourite bool `json:"favourite"`

	// Receive a push notification when someone else has mentioned you in a status?
	Mention bool `json:"mention"`

	// Receive a push notification when a status you created has been boosted by someone else?
	Reblog bool `json:"reblog"`

	// Receive a push notification when a poll you voted in or created has ended?
	Poll bool `json:"poll"`

	// Receive a push notification when a subscribed account posts a status?
	Status bool `json:"status"`

	// Receive a push notification when a status you interacted with has been edited?
	Update bool `json:"update"`

	// Receive a push notification when a new user has signed up?
	AdminSignup bool `json:"admin.sign_up"`

	// Receive a push notification when a new report has been filed?
	AdminReport bool `json:"admin.report"`

	// Receive a push notification when a fave is pending?
	PendingFavourite bool `json:"pending.favourite"`

	// Receive a push notification when a reply is pending?
	PendingReply bool `json:"pending.reply"`

	// Receive a push notification when a boost is pending?
	PendingReblog bool `json:"pending.reblog"`
}

// WebPushSubscriptionCreateRequest captures params for creating or replacing a Web Push subscription.
//
// swagger:ignore
type WebPushSubscriptionCreateRequest struct {
	Subscription *WebPushSubscriptionRequestSubscription `form:"-" json:"subscription"`

	SubscriptionEndpoint   *string `form:"subscription[endpoint]" json:"-"`
	SubscriptionKeysAuth   *string `form:"subscription[keys][auth]" json:"-"`
	SubscriptionKeysP256dh *string `form:"subscription[keys][p256dh]" json:"-"`

	WebPushSubscriptionUpdateRequest
}

// WebPushSubscriptionRequestSubscription is the part of a Web Push subscription that is fixed at creation.
//
// swagger:ignore
type WebPushSubscriptionRequestSubscription struct {
	// Endpoint is the URL to which Web Push notifications will be sent.
	Endpoint string `json:"endpoint"`

	Keys WebPushSubscriptionRequestSubscriptionKeys `json:"keys"`
}

// WebPushSubscriptionRequestSubscriptionKeys is the part of a Web Push subscription that contains auth secrets.
//
// swagger:ignore
type WebPushSubscriptionRequestSubscriptionKeys struct {
	// Auth is the auth secret, a Base64 encoded string of 16 bytes of random data.
	Auth string `json:"auth"`

	// P256dh is the user agent public key, a Base64 encoded string of a public key from an ECDH keypair using the prime256v1 curve.
	P256dh string `json:"p256dh"`
}

// WebPushSubscriptionUpdateRequest captures params for updating a Web Push subscription.
//
// swagger:ignore
type WebPushSubscriptionUpdateRequest struct {
	Data *WebPushSubscriptionRequestData `form:"-" json:"data"`

	DataAlertsFollow           *bool `form:"data[alerts][follow]" json:"-"`
	DataAlertsFollowRequest    *bool `form:"data[alerts][follow_request]" json:"-"`
	DataAlertsFavourite        *bool `form:"data[alerts][favourite]" json:"-"`
	DataAlertsMention          *bool `form:"data[alerts][mention]" json:"-"`
	DataAlertsReblog           *bool `form:"data[alerts][reblog]" json:"-"`
	DataAlertsPoll             *bool `form:"data[alerts][poll]" json:"-"`
	DataAlertsStatus           *bool `form:"data[alerts][status]" json:"-"`
	DataAlertsUpdate           *bool `form:"data[alerts][update]" json:"-"`
	DataAlertsAdminSignup      *bool `form:"data[alerts][admin.sign_up]" json:"-"`
	DataAlertsAdminReport      *bool `form:"data[alerts][admin.report]" json:"-"`
	DataAlertsPendingFavourite *bool `form:"data[alerts][pending.favourite]" json:"-"`
	DataAlertsPendingReply     *bool `form:"data[alerts][pending.reply]" json:"-"`
	DataAlertsPendingReblog    *bool `form:"data[alerts][pending.reblog]" json:"-"`

	DataPolicy *WebPushNotificationPolicy `form:"data[policy]" json:"-"`
}

// WebPushSubscriptionRequestData is the part of a Web Push subscription that can be changed after creation.
//
// swagger:ignore
type WebPushSubscriptionRequestData struct {
	// Alerts selects the specific events that this Web Push subscription will receive.
	Alerts *WebPushSubscriptionAlerts `form:"-" json:"alerts"`

	// Policy selects which accounts will trigger Web Push notifications.
	Policy *WebPushNotificationPolicy `form:"-" json:"policy"`
}

// WebPushNotificationPolicy names sets of accounts that can generate notifications.
type WebPushNotificationPolicy string

const (
	// WebPushNotificationPolicyAll allows all accounts to send notifications to the subscribing user.
	WebPushNotificationPolicyAll WebPushNotificationPolicy = "all"
	// WebPushNotificationPolicyFollowed allows accounts followed by the subscribing user to send notifications.
	WebPushNotificationPolicyFollowed WebPushNotificationPolicy = "followed"
	// WebPushNotificationPolicyFollower allows accounts following the subscribing user to send notifications.
	WebPushNotificationPolicyFollower WebPushNotificationPolicy = "follower"
	// WebPushNotificationPolicyNone doesn't allow any acounts to send notifications to the subscribing user.
	WebPushNotificationPolicyNone WebPushNotificationPolicy = "none"
)
