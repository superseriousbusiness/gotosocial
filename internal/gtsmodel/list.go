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

import "time"

// List refers to a list of follows for which the owning account wants to view a timeline of posts.
type List struct {
	ID            string        `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt     time.Time     `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt     time.Time     `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Title         string        `bun:",nullzero,notnull,unique:listaccounttitle"`                   // Title of this list.
	AccountID     string        `bun:"type:CHAR(26),notnull,nullzero,unique:listaccounttitle"`      // Account that created/owns the list
	Account       *Account      `bun:"-"`                                                           // Account corresponding to accountID
	RepliesPolicy RepliesPolicy `bun:",nullzero,notnull,default:'followed'"`                        // RepliesPolicy for this list.
	Exclusive     *bool         `bun:",nullzero,notnull,default:false"`                             // Hide posts from members of this list from your home timeline.
}

// ListEntry refers to a single follow entry in a list.
type ListEntry struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	ListID    string    `bun:"type:CHAR(26),notnull,nullzero,unique:listentrylistfollow"`   // ID of the list that this entry belongs to.
	FollowID  string    `bun:"type:CHAR(26),notnull,nullzero,unique:listentrylistfollow"`   // Follow that the account owning this entry wants to see posts of in the timeline.
	Follow    *Follow   `bun:"-"`                                                           // Follow corresponding to followID.
}

// RepliesPolicy denotes which replies should be shown in the list.
type RepliesPolicy string

const (
	RepliesPolicyFollowed RepliesPolicy = "followed" // Show replies to any followed user.
	RepliesPolicyList     RepliesPolicy = "list"     // Show replies to members of the list only.
	RepliesPolicyNone     RepliesPolicy = "none"     // Don't show replies.
)
