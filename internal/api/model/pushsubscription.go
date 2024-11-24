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

// PushSubscription represents a subscription to the push streaming server.
//
// swagger:model pushSubscription
type PushSubscription struct {
	// The id of the push subscription in the database.
	ID string `json:"id"`
	// Where push alerts will be sent to.
	Endpoint string `json:"endpoint"`
	// The streaming server's VAPID public key.
	ServerKey string `json:"server_key"`
	// Which alerts should be delivered to the endpoint.
	Alerts *PushSubscriptionAlerts `json:"alerts"`
}

// PushSubscriptionAlerts represents the specific alerts that this push subscription will give.
//
// swagger:model pushSubscriptionAlerts
type PushSubscriptionAlerts struct {
	// Receive a push notification when someone has followed you?
	Follow bool `json:"follow"`
	// Receive a push notification when someone has requested to followed you?
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
}
