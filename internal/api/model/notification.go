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

package model

// Notification represents a notification of an event relevant to the user.
//
// swagger:model notification
type Notification struct {
	// REQUIRED

	// The id of the notification in the database.
	ID string `json:"id"`
	// The type of event that resulted in the notification.
	// 	follow = Someone followed you
	// 	follow_request = Someone requested to follow you
	// 	mention = Someone mentioned you in their status
	// 	reblog = Someone boosted one of your statuses
	// 	favourite = Someone favourited one of your statuses
	// 	poll = A poll you have voted in or created has ended
	// 	status = Someone you enabled notifications for has posted a status
	Type string `json:"type"`
	// The timestamp of the notification (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The account that performed the action that generated the notification.
	Account *Account `json:"account"`

	// OPTIONAL

	// Status that was the object of the notification, e.g. in mentions, reblogs, favourites, or polls.
	Status *Status `json:"status,omitempty"`
}

/*
	The below functions are added onto the apimodel notification so that it satisfies
	the Timelineable interface in internal/timeline.
*/

func (n *Notification) GetID() string {
	return n.ID
}

func (n *Notification) GetAccountID() string {
	return ""
}

func (n *Notification) GetBoostOfID() string {
	return ""
}

func (n *Notification) GetBoostOfAccountID() string {
	return ""
}
