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

import "time"

// EmailDomainBlock represents a domain that the server should automatically reject sign-up requests from.
type EmailDomainBlock struct {
	// ID of this block in the database
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// Email domain to block. Eg. 'gmail.com' or 'hotmail.com'
	Domain string `bun:",notnull"`
	// When was this block created
	CreatedAt time.Time `bun:"type:timestamp,notnull,default:current_timestamp"`
	// When was this block updated
	UpdatedAt time.Time `bun:"type:timestamp,notnull,default:current_timestamp"`
	// Account ID of the creator of this block
	CreatedByAccountID string   `bun:"type:CHAR(26),notnull"`
	CreatedByAccount   *Account `bun:"rel:belongs-to"`
}
