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

import (
	"time"
)

// Instance represents an instance entry as serialized in an export file.
type Instance struct {
	Type                   Type       `json:"type" bun:"-"`
	ID                     string     `json:"id" bun:",nullzero"`
	CreatedAt              *time.Time `json:"createdAt" bun:",nullzero"`
	Domain                 string     `json:"domain" bun:",nullzero"`
	Title                  string     `json:"title,omitempty" bun:",nullzero"`
	URI                    string     `json:"uri" bun:",nullzero"`
	SuspendedAt            *time.Time `json:"suspendedAt,omitempty" bun:",nullzero"`
	DomainBlockID          string     `json:"domainBlockID,omitempty" bun:",nullzero"`
	ShortDescription       string     `json:"shortDescription,omitempty" bun:",nullzero"`
	Description            string     `json:"description,omitempty" bun:",nullzero"`
	Terms                  string     `json:"terms,omitempty" bun:",nullzero"`
	ContactEmail           string     `json:"contactEmail,omitempty" bun:",nullzero"`
	ContactAccountUsername string     `json:"contactAccountUsername,omitempty" bun:",nullzero"`
	ContactAccountID       string     `json:"contactAccountID,omitempty" bun:",nullzero"`
	Reputation             int64      `json:"reputation"`
	Version                string     `json:"version,omitempty" bun:",nullzero"`
}
