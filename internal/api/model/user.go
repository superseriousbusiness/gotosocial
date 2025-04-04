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

package model

// User models fields relevant to one user.
//
// swagger:model user
type User struct {
	// Database ID of this user.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	ID string `json:"id"`
	// Time this user was created. (ISO 8601 Datetime)
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`
	// Confirmed email address of this user, if set.
	// example: someone@example.org
	Email string `json:"email,omitempty"`
	// Unconfirmed email address of this user, if set.
	// example: someone.else@somewhere.else.example.org
	UnconfirmedEmail string `json:"unconfirmed_email,omitempty"`
	// Reason for sign-up, if provided.
	// example: Please! Pretty please!
	Reason string `json:"reason,omitempty"`
	// Time at which this user was last emailed, if at all. (ISO 8601 Datetime)
	// example: 2021-07-30T09:20:25+00:00
	LastEmailedAt string `json:"last_emailed_at,omitempty"`
	// Time at which the email given in the `email` field was confirmed, if at all. (ISO 8601 Datetime)
	// example: 2021-07-30T09:20:25+00:00
	ConfirmedAt string `json:"confirmed_at,omitempty"`
	// Time when the last "please confirm your email address" email was sent, if at all. (ISO 8601 Datetime)
	// example: 2021-07-30T09:20:25+00:00
	ConfirmationSentAt string `json:"confirmation_sent_at,omitempty"`
	// User is a moderator.
	// example: false
	Moderator bool `json:"moderator"`
	// User is an admin.
	// example: false
	Admin bool `json:"admin"`
	// User's account is disabled.
	// example: false
	Disabled bool `json:"disabled"`
	// User was approved by an admin.
	// example: true
	Approved bool `json:"approved"`
	// Time when the last "please reset your password" email was sent, if at all. (ISO 8601 Datetime)
	// example: 2021-07-30T09:20:25+00:00
	ResetPasswordSentAt string `json:"reset_password_sent_at,omitempty"`
	// Time at which 2fa was enabled for this user. (ISO 8601 Datetime)
	// example: 2021-07-30T09:20:25+00:00
	TwoFactorEnabledAt string `json:"two_factor_enabled_at,omitempty"`
}

// PasswordChangeRequest models user password change parameters.
//
// swagger:parameters userPasswordChange
type PasswordChangeRequest struct {
	// User's previous password.
	//
	// in: formData
	// required: true
	OldPassword string `form:"old_password" json:"old_password" xml:"old_password" validation:"required"`
	// Desired new password.
	// If the password does not have high enough entropy, it will be rejected.
	// See https://github.com/wagslane/go-password-validator
	//
	// in: formData
	// required: true
	NewPassword string `form:"new_password" json:"new_password" xml:"new_password" validation:"required"`
}

// EmailChangeRequest models user email change parameters.
//
// swagger:parameters userEmailChange
type EmailChangeRequest struct {
	// User's current password, for verification.
	//
	// in: formData
	// required: true
	Password string `form:"password" json:"password" xml:"password" validation:"required"`
	// Desired new email address.
	//
	// in: formData
	// required: true
	NewEmail string `form:"new_email" json:"new_email" xml:"new_email" validation:"required"`
}
