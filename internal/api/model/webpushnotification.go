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

// WebPushNotification represents a notification summary delivered to the client by the Web Push server.
// It does not contain an entire Notification, just the NotificationID and some preview information.
// It is not used in the client API directly, but is included in the API doc for decoding Web Push notifications.
//
// swagger:model webPushNotification
type WebPushNotification struct {
	// NotificationID is the Notification.ID of the referenced Notification.
	NotificationID string `json:"notification_id"`

	// NotificationType is the Notification.Type of the referenced Notification.
	NotificationType string `json:"notification_type"`

	// Title is a title for the notification,
	// generally describing an action taken by a user.
	Title string `json:"title"`

	// Body is a preview of the notification body,
	// such as the first line of a status's CW or text,
	// or the first line of an account bio.
	Body string `json:"body"`

	// Icon is an image URL that can be displayed with the notification,
	// normally the account's avatar.
	Icon string `json:"icon"`

	// PreferredLocale is a BCP 47 language tag for the receiving user's locale.
	PreferredLocale string `json:"preferred_locale"`

	// AccessToken is the access token associated with the Web Push subscription.
	// I don't know why this is sent, given that the client should know that already,
	// but Feditext does use it.
	AccessToken string `json:"access_token"`
}
