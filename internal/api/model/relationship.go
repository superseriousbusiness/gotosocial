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

// Relationship represents a relationship between accounts.
//
// swagger:model accountRelationship
type Relationship struct {
	// The account id.
	// example: 01FBW9XGEP7G6K88VY4S9MPE1R
	ID string `json:"id"`
	// You are following this account.
	Following bool `json:"following"`
	// You are seeing reblogs/boosts from this account in your home timeline.
	ShowingReblogs bool `json:"showing_reblogs"`
	// You are seeing notifications when this account posts.
	Notifying bool `json:"notifying"`
	// This account follows you.
	FollowedBy bool `json:"followed_by"`
	// You are blocking this account.
	Blocking bool `json:"blocking"`
	// This account is blocking you.
	BlockedBy bool `json:"blocked_by"`
	// You are muting this account.
	Muting bool `json:"muting"`
	// You are muting notifications from this account.
	MutingNotifications bool `json:"muting_notifications"`
	// You have requested to follow this account, and the request is pending.
	Requested bool `json:"requested"`
	// You are blocking this account's domain.
	DomainBlocking bool `json:"domain_blocking"`
	// You are featuring this account on your profile.
	Endorsed bool `json:"endorsed"`
	// Your note on this account.
	Note string `json:"note"`
}
