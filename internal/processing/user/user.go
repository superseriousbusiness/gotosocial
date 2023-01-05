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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Processor wraps a bunch of functions for processing user-level actions.
type Processor interface {
	// ChangePassword changes the specified user's password from old => new,
	// or returns an error if the new password is too weak, or the old password is incorrect.
	ChangePassword(ctx context.Context, user *gtsmodel.User, oldPassword string, newPassword string) gtserror.WithCode
	// SendConfirmEmail sends a 'confirm-your-email-address' type email to a user.
	SendConfirmEmail(ctx context.Context, user *gtsmodel.User, username string) error
	// ConfirmEmail confirms an email address using the given token.
	ConfirmEmail(ctx context.Context, token string) (*gtsmodel.User, gtserror.WithCode)
}

type processor struct {
	emailSender email.Sender
	db          db.DB
}

// New returns a new user processor
func New(db db.DB, emailSender email.Sender) Processor {
	return &processor{
		emailSender: emailSender,
		db:          db,
	}
}
