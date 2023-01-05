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

// Instance represents a federated instance, either local or remote.
type Instance struct {
	ID                     string       `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                     // id of this item in the database
	CreatedAt              time.Time    `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`              // when was item created
	UpdatedAt              time.Time    `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`              // when was item last updated
	Domain                 string       `validate:"required,fqdn" bun:",nullzero,notnull,unique"`                                     // Instance domain eg example.org
	Title                  string       `validate:"-" bun:""`                                                                         // Title of this instance as it would like to be displayed.
	URI                    string       `validate:"required,url" bun:",nullzero,notnull,unique"`                                      // base URI of this instance eg https://example.org
	SuspendedAt            time.Time    `validate:"-" bun:"type:timestamptz,nullzero"`                                                // When was this instance suspended, if at all?
	DomainBlockID          string       `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                      // ID of any existing domain block for this instance in the database
	DomainBlock            *DomainBlock `validate:"-" bun:"rel:belongs-to"`                                                           // Domain block corresponding to domainBlockID
	ShortDescription       string       `validate:"-" bun:""`                                                                         // Short description of this instance
	Description            string       `validate:"-" bun:""`                                                                         // Longer description of this instance
	Terms                  string       `validate:"-" bun:""`                                                                         // Terms and conditions of this instance
	ContactEmail           string       `validate:"omitempty,email" bun:""`                                                           // Contact email address for this instance
	ContactAccountUsername string       `validate:"required_with=ContactAccountID" bun:",nullzero"`                                   // Username of the contact account for this instance
	ContactAccountID       string       `validate:"required_with=ContactAccountUsername,omitempty,ulid" bun:"type:CHAR(26),nullzero"` // Contact account ID in the database for this instance
	ContactAccount         *Account     `validate:"-" bun:"rel:belongs-to"`                                                           // account corresponding to contactAccountID
	Reputation             int64        `validate:"-" bun:",notnull,default:0"`                                                       // Reputation score of this instance
	Version                string       `validate:"-" bun:",nullzero"`                                                                // Version of the software used on this instance
}
