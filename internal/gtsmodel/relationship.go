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

// Relationship describes a requester's relationship with another account.
type Relationship struct {
	// The account id.
	ID string
	// Are you following this user?
	Following bool
	// Are you receiving this user's boosts in your home timeline?
	ShowingReblogs bool
	// Have you enabled notifications for this user?
	Notifying bool
	// Are you followed by this user?
	FollowedBy bool
	// Are you blocking this user?
	Blocking bool
	// Is this user blocking you?
	BlockedBy bool
	// Are you muting this user?
	Muting bool
	// Are you muting notifications from this user?
	MutingNotifications bool
	// Do you have a pending follow request for this user?
	Requested bool
	// Are you blocking this user's domain?
	DomainBlocking bool
	// Are you featuring this user on your profile?
	Endorsed bool
	// Your note on this account.
	Note string
}
