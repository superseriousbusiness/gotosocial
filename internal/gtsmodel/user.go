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
	// Database ID of the user.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Datetime when the user was created.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Datetime when was the user was last updated.
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Confirmed email address for this user.
	//
	// This should be unique, ie., only one email
	// address registered per instance. Multiple
	// users per email are not (yet) supported.
	Email string `bun:",nullzero,unique"`

	// Database ID of the Account for this user.
	AccountID string `bun:"type:CHAR(26),nullzero,notnull,unique"`

	// Account corresponding to AccountID.
	Account *Account `bun:"-"`

	// Bcrypt-encrypted password of this user, generated using
	// https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword.
	//
	// A salt is included so we're safe against 🌈 tables.
	EncryptedPassword string `bun:",nullzero,notnull"`

	// 2FA secret for this user.
	//
	// Null if 2FA is not enabled for this user.
	TwoFactorSecret string `bun:",nullzero"`

	// Slice of bcrypt-encrypted backup/recovery codes that a
	// user can use if they lose their 2FA authenticator app.
	//
	// Null if 2FA is not enabled for this user.
	TwoFactorBackups []string `bun:",nullzero,array"`

	// Datetime when 2fa was enabled.
	//
	// Null if 2fa is not enabled for this user.
	TwoFactorEnabledAt time.Time `bun:"type:timestamptz,nullzero"`

	// IP this user used to sign up.
	//
	// Only stored for pending sign-ups.
	SignUpIP net.IP `bun:",nullzero"`

	// Database ID of the invite that this
	// user used to sign up, if applicable.
	InviteID string `bun:"type:CHAR(26),nullzero"`

	// Reason given for signing up
	// when this user was created.
	Reason string `bun:",nullzero"`

	// Timezone/locale in which
	// this user is located.
	Locale string `bun:",nullzero"`

	// Database ID of the Application used to create this user.
	CreatedByApplicationID string `bun:"type:CHAR(26),nullzero"`

	// Application corresponding to ApplicationID.
	CreatedByApplication *Application `bun:"-"`

	// Datetime when this user was last contacted by email.
	LastEmailedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Confirmation token emailed to this user.
	//
	// Only set if user's email not yet confirmed.
	ConfirmationToken string `bun:",nullzero"`

	// Datetime when confirmation token was emailed to user.
	ConfirmationSentAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime when user confirmed
	// their email address, if applicable.
	ConfirmedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Email address that hasn't yet been confirmed.
	UnconfirmedEmail string `bun:",nullzero"`

	// True if user has moderator role.
	Moderator *bool `bun:",nullzero,notnull,default:false"`

	// True if user has admin role.
	Admin *bool `bun:",nullzero,notnull,default:false"`

	// True if user is disabled from posting.
	Disabled *bool `bun:",nullzero,notnull,default:false"`

	// True if this user's sign up has
	// been approved by a moderator or admin.
	Approved *bool `bun:",nullzero,notnull,default:false"`

	// Reset password token that the user
	// can use to reset their password.
	ResetPasswordToken string `bun:",nullzero"`

	// Datetime when reset password token was emailed to user.
	ResetPasswordSentAt time.Time `bun:"type:timestamptz,nullzero"`

	// If the login for the user is managed
	// externally (e.g., via OIDC), this is a stable
	// reference to the external object (e.g OIDC sub claim).
	ExternalID string `bun:",nullzero,unique"`
}

func (u *User) TwoFactorEnabled() bool {
	return !u.TwoFactorEnabledAt.IsZero()
}

// DeniedUser represents one user sign-up that
// was submitted to the instance and denied.
type DeniedUser struct {
	// Database ID of the user.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Datetime when the user was denied.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Datetime when the denied user was last updated.
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Email address provided on the sign-up form.
	Email string `bun:",nullzero,notnull"`

	// Username provided on the sign-up form.
	Username string `bun:",nullzero,notnull"`

	// IP address the sign-up originated from.
	SignUpIP net.IP `bun:",nullzero"`

	// Invite ID provided on the sign-up form (if applicable).
	InviteID string `bun:"type:CHAR(26),nullzero"`

	// Locale provided on the sign-up form.
	Locale string `bun:",nullzero"`

	// ID of application used to create this sign-up.
	CreatedByApplicationID string `bun:"type:CHAR(26),nullzero"`

	// Reason provided by user on the sign-up form.
	SignUpReason string `bun:",nullzero"`

	// Comment from instance admin about why this sign-up was denied.
	PrivateComment string `bun:",nullzero"`

	// Send an email informing user that their sign-up has been denied.
	SendEmail *bool `bun:",nullzero,notnull,default:false"`

	// Message to include when sending an email to the
	// denied user's email address, if SendEmail is true.
	Message string `bun:",nullzero"`
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
