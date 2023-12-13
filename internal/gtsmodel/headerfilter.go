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
	"unsafe"
)

func init() {
	// Note that since all of the below calculations are
	// constant, these should be optimized out of builds.
	const filterSz = unsafe.Sizeof(HeaderFilter{})
	if unsafe.Sizeof(HeaderFilterAllow{}) != filterSz {
		panic("HeaderFilterAllow{} needs to have the same in-memory size / layout as HeaderFilter{}")
	}
	if unsafe.Sizeof(HeaderFilterBlock{}) != filterSz {
		panic("HeaderFilterBlock{} needs to have the same in-memory size / layout as HeaderFilter{}")
	}
}

// HeaderFilterAllow represents an allow HTTP header filter in the database.
type HeaderFilterAllow struct{ HeaderFilter }

// HeaderFilterBlock represents a block HTTP header filter in the database.
type HeaderFilterBlock struct{ HeaderFilter }

// HeaderFilter represents an HTTP request filter in
// the database, with a header to match against, value
// matching regex, and details about its creation.
type HeaderFilter struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database
	Header    string    `bun:",nullzero,notnull"`                                           // Request header this filter pertains to
	Regex     string    `bun:",nullzero,notnull"`                                           // Request header value matching regular expression
	AuthorID  string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this filter
	Author    *Account  `bun:"-"`                                                           // Account corresponding to AuthorID
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
}
