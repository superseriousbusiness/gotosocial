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

package model

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
