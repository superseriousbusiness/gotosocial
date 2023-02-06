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

package user

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

func (p *processor) ChangePassword(ctx context.Context, user *gtsmodel.User, oldPassword string, newPassword string) gtserror.WithCode {
	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(oldPassword)); err != nil {
		return gtserror.NewErrorUnauthorized(err, "old password was incorrect")
	}

	if err := validate.NewPassword(newPassword); err != nil {
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return gtserror.NewErrorInternalError(err, "error hashing password")
	}

	user.EncryptedPassword = string(newPasswordHash)

	if err := p.db.UpdateUser(ctx, user, "encrypted_password"); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
