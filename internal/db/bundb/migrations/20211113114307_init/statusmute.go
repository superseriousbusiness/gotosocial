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

// StatusMute refers to one account having muted the status of another account or its own.
type StatusMute struct {
	ID              string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt       time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt       time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	AccountID       string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // id of the account that created ('did') the mute
	Account         *Account  `validate:"-" bun:"rel:belongs-to"`                                              // pointer to the account specified by accountID
	TargetAccountID string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // id the account owning the muted status (can be the same as accountID)
	TargetAccount   *Account  `validate:"-" bun:"rel:belongs-to"`                                              // pointer to the account specified by targetAccountID
	StatusID        string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // database id of the status that has been muted
	Status          *Status   `validate:"-" bun:"rel:belongs-to"`                                              // pointer to the muted status specified by statusID
}
