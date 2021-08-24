/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package gtsmodel

import "time"

// Notification models an alert/notification sent to an account about something like a reblog, like, new follow request, etc.
type Notification struct {
	// ID of this notification in the database
	ID string `bun:"type:CHAR(26),pk,notnull"`
	// Type of this notification
	NotificationType NotificationType `bun:",notnull"`
	// Creation time of this notification
	CreatedAt time.Time `bun:"type:timestamp,notnull,default:current_timestamp"`
	// Which account does this notification target (ie., who will receive the notification?)
	TargetAccountID string   `bun:"type:CHAR(26),notnull"`
	TargetAccount   *Account `bun:"rel:belongs-to"`
	// Which account performed the action that created this notification?
	OriginAccountID string   `bun:"type:CHAR(26),notnull"`
	OriginAccount   *Account `bun:"rel:belongs-to"`
	// If the notification pertains to a status, what is the database ID of that status?
	StatusID string  `bun:"type:CHAR(26)"`
	Status   *Status `bun:"rel:belongs-to"`
	// Has this notification been read already?
	Read bool
}

// NotificationType describes the reason/type of this notification.
type NotificationType string

const (
	// NotificationFollow -- someone followed you
	NotificationFollow NotificationType = "follow"
	// NotificationFollowRequest -- someone requested to follow you
	NotificationFollowRequest NotificationType = "follow_request"
	// NotificationMention -- someone mentioned you in their status
	NotificationMention NotificationType = "mention"
	// NotificationReblog -- someone boosted one of your statuses
	NotificationReblog NotificationType = "reblog"
	// NotificationFave -- someone faved/liked one of your statuses
	NotificationFave NotificationType = "favourite"
	// NotificationPoll -- a poll you voted in or created has ended
	NotificationPoll NotificationType = "poll"
	// NotificationStatus -- someone you enabled notifications for has posted a status.
	NotificationStatus NotificationType = "status"
)
