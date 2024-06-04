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
	"time"
)

// UserMute refers to the muting of one account by another.
type UserMute struct {
	ID              string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt       time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt       time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	ExpiresAt       time.Time `bun:"type:timestamptz,nullzero"`                                   // Time mute should expire. If null, should not expire.
	AccountID       string    `bun:"type:CHAR(26),unique:mutesrctarget,notnull,nullzero"`         // Who does this mute originate from?
	Account         *Account  `bun:"rel:belongs-to"`                                              // Account corresponding to accountID
	TargetAccountID string    `bun:"type:CHAR(26),unique:mutesrctarget,notnull,nullzero"`         // Who is the target of this mute?
	TargetAccount   *Account  `bun:"rel:belongs-to"`                                              // Account corresponding to targetAccountID
	Notifications   *bool     `bun:",nullzero,notnull,default:false"`                             // Apply mute to notifications as well as statuses.
}

// Expired returns whether the mute has expired at a given time.
// Mutes without an expiration timestamp never expire.
func (u *UserMute) Expired(now time.Time) bool {
	return !u.ExpiresAt.IsZero() && !u.ExpiresAt.After(now)
}
