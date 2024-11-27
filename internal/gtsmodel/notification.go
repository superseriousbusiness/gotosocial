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
	"strings"
	"time"
)

// Notification models an alert/notification sent to an account about something like a reblog, like, new follow request, etc.
type Notification struct {
	ID               string           `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt        time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt        time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	NotificationType NotificationType `bun:",nullzero,notnull"`                                           // Type of this notification
	TargetAccountID  string           `bun:"type:CHAR(26),nullzero,notnull"`                              // ID of the account targeted by the notification (ie., who will receive the notification?)
	TargetAccount    *Account         `bun:"-"`                                                           // Account corresponding to TargetAccountID. Can be nil, always check first + select using ID if necessary.
	OriginAccountID  string           `bun:"type:CHAR(26),nullzero,notnull"`                              // ID of the account that performed the action that created the notification.
	OriginAccount    *Account         `bun:"-"`                                                           // Account corresponding to OriginAccountID. Can be nil, always check first + select using ID if necessary.
	StatusID         string           `bun:"type:CHAR(26),nullzero"`                                      // If the notification pertains to a status, what is the database ID of that status?
	Status           *Status          `bun:"-"`                                                           // Status corresponding to StatusID. Can be nil, always check first + select using ID if necessary.
	Read             *bool            `bun:",nullzero,notnull,default:false"`                             // Notification has been seen/read
}

// NotificationType describes the
// reason/type of this notification.
type NotificationType enumType

const (
	// Notification Types
	NotificationUnknown       NotificationType = 0  // NotificationUnknown -- unknown notification type, error if this occurs
	NotificationFollow        NotificationType = 1  // NotificationFollow -- someone followed you
	NotificationFollowRequest NotificationType = 2  // NotificationFollowRequest -- someone requested to follow you
	NotificationMention       NotificationType = 3  // NotificationMention -- someone mentioned you in their status
	NotificationReblog        NotificationType = 4  // NotificationReblog -- someone boosted one of your statuses
	NotificationFave          NotificationType = 5  // NotificationFave -- someone faved/liked one of your statuses
	NotificationPoll          NotificationType = 6  // NotificationPoll -- a poll you voted in or created has ended
	NotificationStatus        NotificationType = 7  // NotificationStatus -- someone you enabled notifications for has posted a status.
	NotificationSignup        NotificationType = 8  // NotificationSignup -- someone has submitted a new account sign-up to the instance.
	NotificationPendingFave   NotificationType = 9  // Someone has faved a status of yours, which requires approval by you.
	NotificationPendingReply  NotificationType = 10 // Someone has replied to a status of yours, which requires approval by you.
	NotificationPendingReblog NotificationType = 11 // Someone has boosted a status of yours, which requires approval by you.
)

// String returns a stringified, frontend API compatible form of NotificationType.
func (t NotificationType) String() string {
	switch t {
	case NotificationFollow:
		return "follow"
	case NotificationFollowRequest:
		return "follow_request"
	case NotificationMention:
		return "mention"
	case NotificationReblog:
		return "reblog"
	case NotificationFave:
		return "favourite"
	case NotificationPoll:
		return "poll"
	case NotificationStatus:
		return "status"
	case NotificationSignup:
		return "admin.sign_up"
	case NotificationPendingFave:
		return "pending.favourite"
	case NotificationPendingReply:
		return "pending.reply"
	case NotificationPendingReblog:
		return "pending.reblog"
	default:
		panic("invalid notification type")
	}
}

// ParseNotificationType returns a notification type from the given value.
func ParseNotificationType(in string) NotificationType {
	switch strings.ToLower(in) {
	case "follow":
		return NotificationFollow
	case "follow_request":
		return NotificationFollowRequest
	case "mention":
		return NotificationMention
	case "reblog":
		return NotificationReblog
	case "favourite":
		return NotificationFave
	case "poll":
		return NotificationPoll
	case "status":
		return NotificationStatus
	case "admin.sign_up":
		return NotificationSignup
	case "pending.favourite":
		return NotificationPendingFave
	case "pending.reply":
		return NotificationPendingReply
	case "pending.reblog":
		return NotificationPendingReblog
	default:
		return NotificationUnknown
	}
}
