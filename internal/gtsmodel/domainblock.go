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

// DomainBlock represents a federation block against a particular domain
type DomainBlock struct {
	ID                 string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Domain             string    `bun:",nullzero,notnull"`                                           // domain to block. Eg. 'whatever.com'
	CreatedByAccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this block
	CreatedByAccount   *Account  `bun:"rel:belongs-to"`                                              // Account corresponding to createdByAccountID
	PrivateComment     string    `bun:""`                                                            // Private comment on this block, viewable to admins
	PublicComment      string    `bun:""`                                                            // Public comment on this block, viewable (optionally) by everyone
	Obfuscate          *bool     `bun:",nullzero,notnull,default:false"`                             // whether the domain name should appear obfuscated when displaying it publicly
	SubscriptionID     string    `bun:"type:CHAR(26),nullzero"`                                      // if this block was created through a subscription, what's the subscription ID?
}
