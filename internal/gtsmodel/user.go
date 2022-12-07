/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

import (
	"net"
	"time"
)

// User represents an actual human user of gotosocial. Note, this is a LOCAL gotosocial user, not a remote account.
// To cross reference this local user with their account (which can be local or remote), use the AccountID field.
type User struct {
	ID                     string       `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt              time.Time    `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time    `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Email                  string       `validate:"required_with=ConfirmedAt" bun:",nullzero,unique"`                    // confirmed email address for this user, this should be unique -- only one email address registered per instance, multiple users per email are not supported
	AccountID              string       `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull,unique"`           // The id of the local gtsmodel.Account entry for this user.
	Account                *Account     `validate:"-" bun:"rel:belongs-to"`                                              // Pointer to the account of this user that corresponds to AccountID.
	EncryptedPassword      string       `validate:"required" bun:",nullzero,notnull"`                                    // The encrypted password of this user, generated using https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword. A salt is included so we're safe against 🌈 tables.
	SignUpIP               net.IP       `validate:"-" bun:",nullzero"`                                                   // From what IP was this user created?
	CurrentSignInAt        time.Time    `validate:"-" bun:"type:timestamptz,nullzero"`                                   // When did the user sign in with their current session.
	CurrentSignInIP        net.IP       `validate:"-" bun:",nullzero"`                                                   // What's the most recent IP of this user
	LastSignInAt           time.Time    `validate:"-" bun:"type:timestamptz,nullzero"`                                   // When did this user last sign in?
	LastSignInIP           net.IP       `validate:"-" bun:",nullzero"`                                                   // What's the previous IP of this user?
	SignInCount            int          `validate:"min=0" bun:",notnull,default:0"`                                      // How many times has this user signed in?
	InviteID               string       `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                         // id of the user who invited this user (who let this joker in?)
	ChosenLanguages        []string     `validate:"-" bun:",nullzero"`                                                   // What languages does this user want to see?
	FilteredLanguages      []string     `validate:"-" bun:",nullzero"`                                                   // What languages does this user not want to see?
	Locale                 string       `validate:"-" bun:",nullzero"`                                                   // In what timezone/locale is this user located?
	CreatedByApplicationID string       `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                         // Which application id created this user? See gtsmodel.Application
	CreatedByApplication   *Application `validate:"-" bun:"rel:belongs-to"`                                              // Pointer to the application corresponding to createdbyapplicationID.
	LastEmailedAt          time.Time    `validate:"-" bun:"type:timestamptz,nullzero"`                                   // When was this user last contacted by email.
	ConfirmationToken      string       `validate:"required_with=ConfirmationSentAt" bun:",nullzero"`                    // What confirmation token did we send this user/what are we expecting back?
	ConfirmationSentAt     time.Time    `validate:"required_with=ConfirmationToken" bun:"type:timestamptz,nullzero"`     // When did we send email confirmation to this user?
	ConfirmedAt            time.Time    `validate:"required_with=Email" bun:"type:timestamptz,nullzero"`                 // When did the user confirm their email address
	UnconfirmedEmail       string       `validate:"required_without=Email" bun:",nullzero"`                              // Email address that hasn't yet been confirmed
	Moderator              *bool        `validate:"-" bun:",nullzero,notnull,default:false"`                             // Is this user a moderator?
	Admin                  *bool        `validate:"-" bun:",nullzero,notnull,default:false"`                             // Is this user an admin?
	Disabled               *bool        `validate:"-" bun:",nullzero,notnull,default:false"`                             // Is this user disabled from posting?
	Approved               *bool        `validate:"-" bun:",nullzero,notnull,default:false"`                             // Has this user been approved by a moderator?
	ResetPasswordToken     string       `validate:"required_with=ResetPasswordSentAt" bun:",nullzero"`                   // The generated token that the user can use to reset their password
	ResetPasswordSentAt    time.Time    `validate:"required_with=ResetPasswordToken" bun:"type:timestamptz,nullzero"`    // When did we email the user their reset-password email?
	ExternalID             string       `validate:"-" bun:",nullzero,unique"`                                            // If the login for the user is managed externally (e.g OIDC), we need to keep a stable reference to the external object (e.g OIDC sub claim)
}
