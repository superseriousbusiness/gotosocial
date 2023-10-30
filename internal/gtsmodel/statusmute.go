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

// StatusMute IS DEPRECATED -- USE THREADMUTE INSTEAD NOW! THIS TABLE DOESN'T EXIST ANYMORE!
type StatusMute struct {
	ID              string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt       time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt       time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	AccountID       string    `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that created ('did') the mute
	Account         *Account  `bun:"rel:belongs-to"`                                              // pointer to the account specified by accountID
	TargetAccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                              // id the account owning the muted status (can be the same as accountID)
	TargetAccount   *Account  `bun:"rel:belongs-to"`                                              // pointer to the account specified by targetAccountID
	StatusID        string    `bun:"type:CHAR(26),nullzero,notnull"`                              // database id of the status that has been muted
	Status          *Status   `bun:"rel:belongs-to"`                                              // pointer to the muted status specified by statusID
}
