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
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Name      string    `bun:",unique,nullzero,notnull"`                                    // (lowercase) name of the tag without the hash prefix
	Useable   *bool     `bun:",nullzero,notnull,default:true"`                              // Tag is useable on this instance.
	Listable  *bool     `bun:",nullzero,notnull,default:true"`                              // Tagged statuses can be listed on this instance.
	Href      string    `bun:"-"`                                                           // Href of the hashtag. Will only be set on freshly-extracted hashtags from remote AP messages. Not stored in the database.
}

// FollowedTag represents a user following a tag.
type FollowedTag struct {
	// ID of the account that follows the tag.
	AccountID string `bun:"type:CHAR(26),pk,nullzero"`

	// ID of the tag.
	TagID string `bun:"type:CHAR(26),pk,nullzero"`
}
