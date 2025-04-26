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

package user

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
	"codeberg.org/gruf/go-byteutil"
	"golang.org/x/crypto/bcrypt"
)

// PasswordChange processes a password change request for the given user.
func (p *Processor) PasswordChange(ctx context.Context, user *gtsmodel.User, oldPassword string, newPassword string) gtserror.WithCode {
	// Ensure provided oldPassword is the correct current password.
	if err := bcrypt.CompareHashAndPassword(
		byteutil.S2B(user.EncryptedPassword),
		byteutil.S2B(oldPassword),
	); err != nil {
		err := gtserror.Newf("%w", err)
		return gtserror.NewErrorUnauthorized(err, "old password was incorrect")
	}

	// Ensure new password is strong enough.
	if err := validate.Password(newPassword); err != nil {
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure new password is different from old password.
	if newPassword == oldPassword {
		const help = "new password cannot be the same as previous password"
		err := gtserror.New(help)
		return gtserror.NewErrorBadRequest(err, help)
	}

	// Hash the new password.
	encryptedPassword, err := bcrypt.GenerateFromPassword(
		byteutil.S2B(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		err := gtserror.Newf("%w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Set new password on user.
	user.EncryptedPassword = string(encryptedPassword)
	if err := p.state.DB.UpdateUser(
		ctx, user,
		"encrypted_password",
	); err != nil {
		err := gtserror.Newf("db error updating user: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
