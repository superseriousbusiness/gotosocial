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

package model

// Relationship represents a relationship between accounts. See https://docs.joinmastodon.org/entities/relationship/
type Relationship struct {
	// The account id.
	ID string `json:"id"`
	// Are you following this user?
	Following bool `json:"following"`
	// Are you receiving this user's boosts in your home timeline?
	ShowingReblogs bool `json:"showing_reblogs"`
	// Have you enabled notifications for this user?
	Notifying bool `json:"notifying"`
	// Are you followed by this user?
	FollowedBy bool `json:"followed_by"`
	// Are you blocking this user?
	Blocking bool `json:"blocking"`
	// Is this user blocking you?
	BlockedBy bool `json:"blocked_by"`
	// Are you muting this user?
	Muting bool `json:"muting"`
	// Are you muting notifications from this user?
	MutingNotifications bool `json:"muting_notifications"`
	// Do you have a pending follow request for this user?
	Requested bool `json:"requested"`
	// Are you blocking this user's domain?
	DomainBlocking bool `json:"domain_blocking"`
	// Are you featuring this user on your profile?
	Endorsed bool `json:"endorsed"`
	// Your note on this account.
	Note string `json:"note"`
}
