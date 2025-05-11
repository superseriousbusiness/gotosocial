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

// StatusFave refers to a 'fave' or 'like' in the database, from one account, targeting the status of another account
type StatusFave struct {
	ID              string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                      // id of this item in the database
	CreatedAt       time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`   // when was item created
	UpdatedAt       time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`   // when was item last updated
	AccountID       string    `bun:"type:CHAR(26),unique:statusfaveaccountstatus,nullzero,notnull"` // id of the account that created ('did') the fave
	Account         *Account  `bun:"-"`                                                             // account that created the fave
	TargetAccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                                // id the account owning the faved status
	TargetAccount   *Account  `bun:"-"`                                                             // account owning the faved status
	StatusID        string    `bun:"type:CHAR(26),unique:statusfaveaccountstatus,nullzero,notnull"` // database id of the status that has been 'faved'
	Status          *Status   `bun:"-"`                                                             // the faved status
	URI             string    `bun:",nullzero,notnull,unique"`                                      // ActivityPub URI of this fave
	PendingApproval *bool     `bun:",nullzero,notnull,default:false"`                               // If true then Like must be Approved by the like-ee before being fully distributed.
	PreApproved     bool      `bun:"-"`                                                             // If true, then fave targets a status on our instance, has permission to do the interaction, and an Accept should be sent out for it immediately. Field not stored in the DB.
	ApprovedByURI   string    `bun:",nullzero"`                                                     // URI of an Accept Activity that approves this Like.
}

// GetAccount returns the account that owns
// this fave. May be nil if fave not populated.
// Fulfils Interaction interface.
func (f *StatusFave) GetAccount() *Account {
	return f.Account
}
