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

// AdminAccountAction models an action taken by an instance administrator on an account.
type AdminAccountAction struct {
	ID              string          `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt       time.Time       `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt       time.Time       `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	AccountID       string          `validate:"required,ulid" bun:"type:CHAR(26),notnull,nullzero"`                  // Who performed this admin action.
	Account         *Account        `validate:"-" bun:"rel:has-one"`                                                 // Account corresponding to accountID
	TargetAccountID string          `validate:"required,ulid" bun:"type:CHAR(26),notnull,nullzero"`                  // Who is the target of this action
	TargetAccount   *Account        `validate:"-" bun:"rel:has-one"`                                                 // Account corresponding to targetAccountID
	Text            string          `validate:"-" bun:""`                                                            // text explaining why this action was taken
	Type            AdminActionType `validate:"oneof=disable silence suspend" bun:",nullzero,notnull"`               // type of action that was taken
	SendEmail       bool            `validate:"-" bun:""`                                                            // should an email be sent to the account owner to explain what happened
	ReportID        string          `validate:",omitempty,ulid" bun:"type:CHAR(26),nullzero"`                        // id of a report connected to this action, if it exists
}

// AdminActionType describes a type of action taken on an entity by an admin
type AdminActionType string

const (
	// AdminActionDisable -- the account or application etc has been disabled but not deleted.
	AdminActionDisable AdminActionType = "disable"
	// AdminActionSilence -- the account or application etc has been silenced.
	AdminActionSilence AdminActionType = "silence"
	// AdminActionSuspend -- the account or application etc has been deleted.
	AdminActionSuspend AdminActionType = "suspend"
)
