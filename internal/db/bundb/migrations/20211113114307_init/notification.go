/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	ID               string           `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                                                                                                                    // id of this item in the database
	CreatedAt        time.Time        `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                                                                                                             // when was item created
	UpdatedAt        time.Time        `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                                                                                                             // when was item last updated                                                                                                                            // when was item created
	NotificationType NotificationType `validate:"oneof=follow follow_request mention reblog favourite poll status" bun:",nullzero,notnull"`                                                                                                        // Type of this notification
	TargetAccountID  string           `validate:"ulid" bun:"type:CHAR(26),nullzero,notnull"`                                                                                                                                                       // Which account does this notification target (ie., who will receive the notification?)
	TargetAccount    *Account         `validate:"-" bun:"rel:belongs-to"`                                                                                                                                                                          // Which account performed the action that created this notification?
	OriginAccountID  string           `validate:"ulid" bun:"type:CHAR(26),nullzero,notnull"`                                                                                                                                                       // ID of the account that performed the action that created the notification.
	OriginAccount    *Account         `validate:"-" bun:"rel:belongs-to"`                                                                                                                                                                          // Account corresponding to originAccountID
	StatusID         string           `validate:"required_if=NotificationType mention,required_if=NotificationType reblog,required_if=NotificationType favourite,required_if=NotificationType status,omitempty,ulid" bun:"type:CHAR(26),nullzero"` // If the notification pertains to a status, what is the database ID of that status?
	Status           *Status          `validate:"-" bun:"rel:belongs-to"`                                                                                                                                                                          // Status corresponding to statusID
	Read             bool             `validate:"-" bun:",notnull,default:false"`                                                                                                                                                                  // Notification has been seen/read
}

// NotificationType describes the reason/type of this notification.
type NotificationType string

// Notification Types
const (
	NotificationFollow        NotificationType = "follow"         // NotificationFollow -- someone followed you
	NotificationFollowRequest NotificationType = "follow_request" // NotificationFollowRequest -- someone requested to follow you
	NotificationMention       NotificationType = "mention"        // NotificationMention -- someone mentioned you in their status
	NotificationReblog        NotificationType = "reblog"         // NotificationReblog -- someone boosted one of your statuses
	NotificationFave          NotificationType = "favourite"      // NotificationFave -- someone faved/liked one of your statuses
	NotificationPoll          NotificationType = "poll"           // NotificationPoll -- a poll you voted in or created has ended
	NotificationStatus        NotificationType = "status"         // NotificationStatus -- someone you enabled notifications for has posted a status.
)
