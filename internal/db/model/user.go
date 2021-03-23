/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package model

import (
	"net"
	"time"
)

// User represents an actual human user of gotosocial. Note, this is a LOCAL gotosocial user, not a remote account.
// To cross reference this local user with their account (which can be local or remote), use the AccountID field.
type User struct {
	/*
		BASIC INFO
	*/

	// id of this user in the local database; the end-user will never need to know this, it's strictly internal
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// confirmed email address for this user, this should be unique -- only one email address registered per instance, multiple users per email are not supported
	Email string `pg:",notnull,unique"`
	// The id of the local gtsmodel.Account entry for this user, if it exists (unconfirmed users don't have an account yet)
	AccountID string `pg:"default:'',notnull,unique"`
	// The encrypted password of this user, generated using https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword. A salt is included so we're safe against 🌈 tables
	EncryptedPassword string `pg:",notnull"`

	/*
		USER METADATA
	*/

	// When was this user created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// From what IP was this user created?
	SignUpIP net.IP
	// When was this user updated (eg., password changed, email address changed)?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When did this user sign in for their current session?
	CurrentSignInAt time.Time `pg:"type:timestamp"`
	// What's the most recent IP of this user
	CurrentSignInIP net.IP
	// When did this user last sign in?
	LastSignInAt time.Time `pg:"type:timestamp"`
	// What's the previous IP of this user?
	LastSignInIP net.IP
	// How many times has this user signed in?
	SignInCount int
	// id of the user who invited this user (who let this guy in?)
	InviteID string
	// What languages does this user want to see?
	ChosenLanguages []string
	// What languages does this user not want to see?
	FilteredLanguages []string
	// In what timezone/locale is this user located?
	Locale string
	// Which application id created this user? See gtsmodel.Application
	CreatedByApplicationID string
	// When did we last contact this user
	LastEmailedAt time.Time `pg:"type:timestamp"`

	/*
		USER CONFIRMATION
	*/

	// What confirmation token did we send this user/what are we expecting back?
	ConfirmationToken string
	// When did the user confirm their email address
	ConfirmedAt time.Time `pg:"type:timestamp"`
	// When did we send email confirmation to this user?
	ConfirmationSentAt time.Time `pg:"type:timestamp"`
	// Email address that hasn't yet been confirmed
	UnconfirmedEmail string

	/*
		ACL FLAGS
	*/

	// Is this user a moderator?
	Moderator bool
	// Is this user an admin?
	Admin bool
	// Is this user disabled from posting?
	Disabled bool
	// Has this user been approved by a moderator?
	Approved bool

	/*
		USER SECURITY
	*/

	// The generated token that the user can use to reset their password
	ResetPasswordToken string
	// When did we email the user their reset-password email?
	ResetPasswordSentAt time.Time `pg:"type:timestamp"`

	EncryptedOTPSecret     string
	EncryptedOTPSecretIv   string
	EncryptedOTPSecretSalt string
	OTPRequiredForLogin    bool
	OTPBackupCodes         []string
	ConsumedTimestamp      int
	RememberToken          string
	SignInToken            string
	SignInTokenSentAt      time.Time `pg:"type:timestamp"`
	WebauthnID             string
}
