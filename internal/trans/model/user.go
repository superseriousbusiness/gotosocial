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

package trans

import (
	"time"
)

// User represents a local instance user as serialized to an export file.
type User struct {
	Type                Type       `json:"type" bun:"-"`
	ID                  string     `json:"id" bun:",nullzero"`
	CreatedAt           *time.Time `json:"createdAt" bun:",nullzero"`
	Email               string     `json:"email,omitempty" bun:",nullzero"`
	AccountID           string     `json:"accountID" bun:",nullzero"`
	EncryptedPassword   string     `json:"encryptedPassword" bun:",nullzero"`
	CurrentSignInAt     *time.Time `json:"currentSignInAt,omitempty" bun:",nullzero"`
	LastSignInAt        *time.Time `json:"lastSignInAt,omitempty" bun:",nullzero"`
	InviteID            string     `json:"inviteID,omitempty" bun:",nullzero"`
	ChosenLanguages     []string   `json:"chosenLanguages,omitempty" bun:",nullzero"`
	FilteredLanguages   []string   `json:"filteredLanguage,omitempty" bun:",nullzero"`
	Locale              string     `json:"locale" bun:",nullzero"`
	LastEmailedAt       time.Time  `json:"lastEmailedAt,omitempty" bun:",nullzero"`
	ConfirmationToken   string     `json:"confirmationToken,omitempty" bun:",nullzero"`
	ConfirmationSentAt  *time.Time `json:"confirmationTokenSentAt,omitempty" bun:",nullzero"`
	ConfirmedAt         *time.Time `json:"confirmedAt,omitempty" bun:",nullzero"`
	UnconfirmedEmail    string     `json:"unconfirmedEmail,omitempty" bun:",nullzero"`
	Moderator           *bool      `json:"moderator" bun:",nullzero,notnull,default:false"`
	Admin               *bool      `json:"admin" bun:",nullzero,notnull,default:false"`
	Disabled            *bool      `json:"disabled" bun:",nullzero,notnull,default:false"`
	Approved            *bool      `json:"approved" bun:",nullzero,notnull,default:false"`
	ResetPasswordToken  string     `json:"resetPasswordToken,omitempty" bun:",nullzero"`
	ResetPasswordSentAt *time.Time `json:"resetPasswordSentAt,omitempty" bun:",nullzero"`
}
