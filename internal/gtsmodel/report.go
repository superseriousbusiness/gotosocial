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

// Report models a user-created reported about an account, which should be reviewed
// and acted upon by instance admins.
//
// This can be either a report created locally (on this instance) about a user on this
// or another instance, OR a report that was created remotely (on another instance)
// about a user on this instance, and received via the federated (s2s) API.
type Report struct {
	ID                     string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt              time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	URI                    string    `validate:"required,url" bun:",unique,nullzero,notnull"`                         // activitypub URI of this report
	AccountID              string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // which account created this report
	Account                *Account  `validate:"-" bun:"-"`                                                           // account corresponding to AccountID
	TargetAccountID        string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // which account is targeted by this report
	TargetAccount          *Account  `validate:"-" bun:"-"`                                                           // account corresponding to TargetAccountID
	Comment                string    `validate:"-" bun:",nullzero"`                                                   // comment / explanation for this report, by the reporter
	StatusIDs              []string  `validate:"dive,ulid" bun:"statuses,array"`                                      // database IDs of any statuses referenced by this report
	Statuses               []*Status `validate:"-" bun:"-"`                                                           // statuses corresponding to StatusIDs
	Forwarded              *bool     `validate:"-" bun:",nullzero,notnull,default:false"`                             // flag to indicate report should be forwarded to remote instance
	ActionTaken            string    `validate:"-" bun:",nullzero"`                                                   // string description of what action was taken in response to this report
	ActionTakenAt          time.Time `validate:"-" bun:"type:timestamptz,nullzero"`                                   // time at which action was taken, if any
	ActionTakenByAccountID string    `validate:",omitempty,ulid" bun:"type:CHAR(26),nullzero"`                        // database ID of account which took action, if any
	ActionTakenByAccount   *Account  `validate:"-" bun:"-"`                                                           // account corresponding to ActionTakenByID, if any
}
