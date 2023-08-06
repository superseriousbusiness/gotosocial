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

import (
	"net"
	"time"
)

// AdminAccountAction models an action taken by an instance administrator on an account.
type AdminAccountAction struct {
	ID              string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt       time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt       time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	AccountID       string          `bun:"type:CHAR(26),notnull,nullzero"`                              // Who performed this admin action.
	Account         *Account        `bun:"rel:has-one"`                                                 // Account corresponding to accountID
	TargetAccountID string          `bun:"type:CHAR(26),notnull,nullzero"`                              // Who is the target of this action
	TargetAccount   *Account        `bun:"rel:has-one"`                                                 // Account corresponding to targetAccountID
	Text            string          `bun:""`                                                            // text explaining why this action was taken
	Type            AdminActionType `bun:",nullzero,notnull"`                                           // type of action that was taken
	SendEmail       bool            `bun:""`                                                            // should an email be sent to the account owner to explain what happened
	ReportID        string          `bun:"type:CHAR(26),nullzero"`                                      // id of a report connected to this action, if it exists
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

// NewSignup models parameters for the creation
// of a new user + account on this instance.
//
// Aside from username, email, and password, it is
// fine to use zero values on fields of this struct.
type NewSignup struct {
	Username string // Username of the new account.
	Email    string // Email address of the user.
	Password string // Plaintext (not yet hashed) password for the user.

	Reason        string // Reason given by the user when submitting a sign up request (optional).
	PreApproved   bool   // Mark the new user/account as preapproved (optional)
	SignUpIP      net.IP // IP address from which the sign up request occurred (optional).
	Locale        string // Locale code for the new account/user (optional).
	AppID         string // ID of the application used to create this account (optional).
	EmailVerified bool   // Mark submitted email address as already verified (optional).
	ExternalID    string // ID of this user in external OIDC system (optional).
	Admin         bool   // Mark new user as an admin user (optional).
}
