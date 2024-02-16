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
	"net/url"
	"time"
)

// Move represents an ActivityPub "Move" activity
// received (or created) by this instance.
type Move struct {
	ID          string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	CreatedAt   time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // When was item created.
	UpdatedAt   time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // When was item last updated.
	AttemptedAt time.Time `bun:"type:timestamptz,nullzero"`                                   // When was processing of the Move to TargetURI last attempted by our instance (zero if not yet attempted).
	SucceededAt time.Time `bun:"type:timestamptz,nullzero"`                                   // When did the processing of the Move to TargetURI succeed according to our criteria (zero if not yet complete).
	OriginURI   string    `bun:",nullzero,notnull,unique:moveorigintarget"`                   // OriginURI of the Move. Ie., the Move Object.
	Origin      *url.URL  `bun:"-"`                                                           // URL corresponding to OriginURI. Not stored in the database.
	TargetURI   string    `bun:",nullzero,notnull,unique:moveorigintarget"`                   // TargetURI of the Move. Ie., the Move Target.
	Target      *url.URL  `bun:"-"`                                                           // URL corresponding to TargetURI. Not stored in the database.
	URI         string    `bun:",nullzero,notnull,unique"`                                    // ActivityPub ID/URI of the Move Activity itself.
}
