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

package trans

import "time"

// DomainBlock represents a domain block as serialized in an exported file.
type DomainBlock struct {
	Type               Type       `json:"type" bun:"-"`
	ID                 string     `json:"id" bun:",nullzero"`
	CreatedAt          *time.Time `json:"createdAt" bun:",nullzero"`
	Domain             string     `json:"domain" bun:",nullzero"`
	CreatedByAccountID string     `json:"createdByAccountID" bun:",nullzero"`
	PrivateComment     string     `json:"privateComment,omitempty" bun:",nullzero"`
	PublicComment      string     `json:"publicComment,omitempty" bun:",nullzero"`
	Obfuscate          *bool      `json:"obfuscate" bun:",nullzero,notnull,default:false"`
	SubscriptionID     string     `json:"subscriptionID,omitempty" bun:",nullzero"`
}
