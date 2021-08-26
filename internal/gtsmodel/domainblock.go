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
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// blocked domain
	Domain string `bun:",pk,notnull,unique"`
	// When was this block created
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this block updated
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// Account ID of the creator of this block
	CreatedByAccountID string   `bun:"type:CHAR(26),notnull"`
	CreatedByAccount   *Account `bun:"rel:belongs-to"`
	// Private comment on this block, viewable to admins
	PrivateComment string `bun:",nullzero"`
	// Public comment on this block, viewable (optionally) by everyone
	PublicComment string `bun:",nullzero"`
	// whether the domain name should appear obfuscated when displaying it publicly
	Obfuscate bool
	// if this block was created through a subscription, what's the subscription ID?
	SubscriptionID string `bun:"type:CHAR(26),nullzero"`
}
