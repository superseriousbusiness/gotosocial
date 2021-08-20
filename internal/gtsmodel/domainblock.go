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

// DomainBlock represents a federation block against a particular domain
type DomainBlock struct {
	// ID of this block in the database
	ID string `pg:"type:CHAR(26),pk,notnull,unique"`
	// blocked domain
	Domain string `pg:",pk,notnull,unique"`
	// When was this block created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this block updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Account ID of the creator of this block
	CreatedByAccountID string   `pg:"type:CHAR(26),notnull"`
	CreatedByAccount   *Account `pg:"rel:belongs-to"`
	// Private comment on this block, viewable to admins
	PrivateComment string
	// Public comment on this block, viewable (optionally) by everyone
	PublicComment string
	// whether the domain name should appear obfuscated when displaying it publicly
	Obfuscate bool
	// if this block was created through a subscription, what's the subscription ID?
	SubscriptionID string `pg:"type:CHAR(26)"`
}
