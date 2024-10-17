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

// Instance represents a federated instance, either local or remote.
type Instance struct {
	ID                     string       `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt              time.Time    `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time    `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Domain                 string       `bun:",nullzero,notnull,unique"`                                    // Instance domain eg example.org
	Title                  string       `bun:""`                                                            // Title of this instance as it would like to be displayed.
	URI                    string       `bun:",nullzero,notnull,unique"`                                    // base URI of this instance eg https://example.org
	SuspendedAt            time.Time    `bun:"type:timestamptz,nullzero"`                                   // When was this instance suspended, if at all?
	DomainBlockID          string       `bun:"type:CHAR(26),nullzero"`                                      // ID of any existing domain block for this instance in the database
	DomainBlock            *DomainBlock `bun:"rel:belongs-to"`                                              // Domain block corresponding to domainBlockID
	ShortDescription       string       `bun:""`                                                            // Short description of this instance
	ShortDescriptionText   string       `bun:""`                                                            // Raw text version of short description (before parsing).
	Description            string       `bun:""`                                                            // Longer description of this instance.
	DescriptionText        string       `bun:""`                                                            // Raw text version of long description (before parsing).
	CustomCSS              string       `bun:",nullzero"`                                                   // Custom CSS for the instance.
	Terms                  string       `bun:""`                                                            // Terms and conditions of this instance.
	TermsText              string       `bun:""`                                                            // Raw text version of terms (before parsing).
	ContactEmail           string       `bun:""`                                                            // Contact email address for this instance
	ContactAccountUsername string       `bun:",nullzero"`                                                   // Username of the contact account for this instance
	ContactAccountID       string       `bun:"type:CHAR(26),nullzero"`                                      // Contact account ID in the database for this instance
	ContactAccount         *Account     `bun:"rel:belongs-to"`                                              // account corresponding to contactAccountID
	Reputation             int64        `bun:",notnull,default:0"`                                          // Reputation score of this instance
	Version                string       `bun:",nullzero"`                                                   // Version of the software used on this instance
	Rules                  []Rule       `bun:"-"`                                                           // List of instance rules
}
