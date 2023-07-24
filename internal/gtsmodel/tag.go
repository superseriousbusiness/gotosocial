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

// Tag represents a hashtag for gathering public statuses together.
type Tag struct {
	ID        string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Name      string    `validate:"required" bun:",unique,nullzero,notnull"`                             // (lowercase) name of the tag without the hash prefix
	Useable   *bool     `validate:"-" bun:",nullzero,notnull,default:true"`                              // Tag is useable on this instance.
	Listable  *bool     `validate:"-" bun:",nullzero,notnull,default:true"`                              // Tagged statuses can be listed on this instance.

	// Original URL of a partially-filled-out tag.
	// This SHOULD NOT be inserted in the database,
	// cached, or serialized towards a user.
	//
	// It is only to be used when parsing a hashtag
	// from a remote instance, since it may be useful
	// to dereference latest notes from the given
	// remote instance every now and then.
	URL string `validate:"-" bun:"-"`
}
