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

// User represents one signed-up user of this GoToSocial instance.
//
// User may not necessarily be approved yet; in other words, this
// model is used for both active users and signed-up but not yet
// approved users.
//
// Sign-ups that have been denied rather than
// approved are stored as DeniedUser instead.
type User struct {
	ID                     string       `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt              time.Time    `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time    `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Email                  string       `bun:",nullzero,unique"`                                            // confirmed email address for this user, this should be unique -- only one email address registered per instance, multiple users per email are not supported
	AccountID              string       `bun:"type:CHAR(26),nullzero,notnull,unique"`                       // The id of the local gtsmodel.Account entry for this user.
	Account                *Account     `bun:"rel:belongs-to"`                                              // Pointer to the account of this user that corresponds to AccountID.
	EncryptedPassword      string       `bun:",nullzero,notnull"`                                           // The encrypted password of this user, generated using https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword. A salt is included so we're safe against 🌈 tables.
	SignUpIP               net.IP       `bun:",nullzero"`                                                   // IP this user used to sign up. Only stored for pending sign-ups.
	InviteID               string       `bun:"type:CHAR(26),nullzero"`                                      // id of the user who invited this user (who let this joker in?)
	Reason                 string       `bun:",nullzero"`                                                   // What reason was given for signing up when this user was created?
	Locale                 string       `bun:",nullzero"`                                                   // In what timezone/locale is this user located?
	CreatedByApplicationID string       `bun:"type:CHAR(26),nullzero"`                                      // Which application id created this user? See gtsmodel.Application
	CreatedByApplication   *Application `bun:"rel:belongs-to"`                                              // Pointer to the application corresponding to createdbyapplicationID.
	LastEmailedAt          time.Time    `bun:"type:timestamptz,nullzero"`                                   // When was this user last contacted by email.
	ConfirmationToken      string       `bun:",nullzero"`                                                   // What confirmation token did we send this user/what are we expecting back?
	ConfirmationSentAt     time.Time    `bun:"type:timestamptz,nullzero"`                                   // When did we send email confirmation to this user?
	ConfirmedAt            time.Time    `bun:"type:timestamptz,nullzero"`                                   // When did the user confirm their email address
	UnconfirmedEmail       string       `bun:",nullzero"`                                                   // Email address that hasn't yet been confirmed
	Moderator              *bool        `bun:",nullzero,notnull,default:false"`                             // Is this user a moderator?
	Admin                  *bool        `bun:",nullzero,notnull,default:false"`                             // Is this user an admin?
	Disabled               *bool        `bun:",nullzero,notnull,default:false"`                             // Is this user disabled from posting?
	Approved               *bool        `bun:",nullzero,notnull,default:false"`                             // Has this user been approved by a moderator?
	ResetPasswordToken     string       `bun:",nullzero"`                                                   // The generated token that the user can use to reset their password
	ResetPasswordSentAt    time.Time    `bun:"type:timestamptz,nullzero"`                                   // When did we email the user their reset-password email?
	ExternalID             string       `bun:",nullzero,unique"`                                            // If the login for the user is managed externally (e.g OIDC), we need to keep a stable reference to the external object (e.g OIDC sub claim)
}

// DeniedUser represents one user sign-up that
// was submitted to the instance and denied.
type DeniedUser struct {
	ID                     string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt              time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Email                  string    `bun:",nullzero,notnull"`                                           // Email address provided on the sign-up form.
	Username               string    `bun:",nullzero,notnull"`                                           // Username provided on the sign-up form.
	SignUpIP               net.IP    `bun:",nullzero"`                                                   // IP address the sign-up originated from.
	InviteID               string    `bun:"type:CHAR(26),nullzero"`                                      // Invite ID provided on the sign-up form (if applicable).
	Locale                 string    `bun:",nullzero"`                                                   // Locale provided on the sign-up form.
	CreatedByApplicationID string    `bun:"type:CHAR(26),nullzero"`                                      // ID of application used to create this sign-up.
	SignUpReason           string    `bun:",nullzero"`                                                   // Reason provided by user on the sign-up form.
	PrivateComment         string    `bun:",nullzero"`                                                   // Comment from instance admin about why this sign-up was denied.
	SendEmail              *bool     `bun:",nullzero,notnull,default:false"`                             // Send an email informing user that their sign-up has been denied.
	Message                string    `bun:",nullzero"`                                                   // Message to include when sending an email to the denied user's email address, if SendEmail is true.
}

// NewSignup models parameters for the creation
// of a new user + account on this instance.
//
// Aside from username, email, and password, it is
// fine to use zero values on fields of this struct.
//
// This struct is not stored in the database,
// it's just for passing around parameters.
type NewSignup struct {
	Username string // Username of the new account (required).
	Email    string // Email address of the user (required).
	Password string // Plaintext (not yet hashed) password for the user (required).

	Reason        string // Reason given by the user when submitting a sign up request (optional).
	PreApproved   bool   // Mark the new user/account as preapproved (optional)
	SignUpIP      net.IP // IP address from which the sign up request occurred (optional).
	Locale        string // Locale code for the new account/user (optional).
	AppID         string // ID of the application used to create this account (optional).
	EmailVerified bool   // Mark submitted email address as already verified (optional).
	ExternalID    string // ID of this user in external OIDC system (optional).
	Admin         bool   // Mark new user as an admin user (optional).
}
