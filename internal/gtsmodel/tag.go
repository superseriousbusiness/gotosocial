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

// Tag represents a hashtag for gathering public statuses together.
type Tag struct {
	ID                     string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt              time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	URL                    string    `validate:"required,url" bun:",nullzero,notnull"`                                // Href/web address of this tag, eg https://example.org/tags/somehashtag
	Name                   string    `validate:"required" bun:",unique,nullzero,notnull"`                             // name of this tag -- the tag without the hash part
	FirstSeenFromAccountID string    `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                         // Which account ID is the first one we saw using this tag?
	Useable                *bool     `validate:"-" bun:",nullzero,notnull,default:true"`                              // can our instance users use this tag?
	Listable               *bool     `validate:"-" bun:",nullzero,notnull,default:true"`                              // can our instance users look up this tag?
	LastStatusAt           time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was this tag last used?
}
