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

type HeaderFilterAllow struct{ HeaderFilter }

type HeaderFilterBlock struct{ HeaderFilter }

type HeaderFilter struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    //
	Key       string    `bun:","`                                                           // Request header key this filter pertains to
	Regex     string    `bun:","`                                                           // Request header value matching regular expression
	AuthorID  string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this filter
	Author    *Account  `bun:"-"`                                                           // Account corresponding to AuthorID
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
}
