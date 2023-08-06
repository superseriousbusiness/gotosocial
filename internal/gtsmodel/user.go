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

// User represents an actual human user of gotosocial. Note, this is a LOCAL gotosocial user, not a remote account.
// To cross reference this local user with their account (which can be local or remote), use the AccountID field.
type User struct {
	ID                     string       `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt              time.Time    `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt              time.Time    `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Email                  string       `bun:",nullzero,unique"`                                            // confirmed email address for this user, this should be unique -- only one email address registered per instance, multiple users per email are not supported
	AccountID              string       `bun:"type:CHAR(26),nullzero,notnull,unique"`                       // The id of the local gtsmodel.Account entry for this user.
	Account                *Account     `bun:"rel:belongs-to"`                                              // Pointer to the account of this user that corresponds to AccountID.
	EncryptedPassword      string       `bun:",nullzero,notnull"`                                           // The encrypted password of this user, generated using https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword. A salt is included so we're safe against 🌈 tables.
	SignUpIP               net.IP       `bun:",nullzero"`                                                   // From what IP was this user created?
	CurrentSignInAt        time.Time    `bun:"type:timestamptz,nullzero"`                                   // When did the user sign in with their current session.
	CurrentSignInIP        net.IP       `bun:",nullzero"`                                                   // What's the most recent IP of this user
	LastSignInAt           time.Time    `bun:"type:timestamptz,nullzero"`                                   // When did this user last sign in?
	LastSignInIP           net.IP       `bun:",nullzero"`                                                   // What's the previous IP of this user?
	SignInCount            int          `bun:",notnull,default:0"`                                          // How many times has this user signed in?
	InviteID               string       `bun:"type:CHAR(26),nullzero"`                                      // id of the user who invited this user (who let this joker in?)
	ChosenLanguages        []string     `bun:",nullzero"`                                                   // What languages does this user want to see?
	FilteredLanguages      []string     `bun:",nullzero"`                                                   // What languages does this user not want to see?
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
