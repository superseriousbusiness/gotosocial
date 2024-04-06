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
	"errors"
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// EmailGetUserForConfirmToken retrieves the user (with account) from
// the database for the given "confirm your email" token string.
func (p *Processor) EmailGetUserForConfirmToken(ctx context.Context, token string) (*gtsmodel.User, gtserror.WithCode) {
	if token == "" {
		err := errors.New("no token provided")
		return nil, gtserror.NewErrorNotFound(err)
	}

	user, err := p.state.DB.GetUserByConfirmationToken(ctx, token)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			return nil, gtserror.NewErrorInternalError(err)
		}

		// No user found for this token.
		return nil, gtserror.NewErrorNotFound(err)
	}

	if user.Account == nil {
		user.Account, err = p.state.DB.GetAccountByID(ctx, user.AccountID)
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			return nil, gtserror.NewErrorInternalError(err)
		}

		// No account found for this user,
		// or error populating account.
		return nil, gtserror.NewErrorNotFound(err)
	}

	if !user.Account.SuspendedAt.IsZero() {
		err := fmt.Errorf("account %s is suspended", user.AccountID)
		return nil, gtserror.NewErrorForbidden(err, err.Error())
	}

	return user, nil
}

// EmailConfirm processes an email confirmation request,
// usually initiated as a result of clicking on a link
// in a 'confirm your email address' type email.
func (p *Processor) EmailConfirm(ctx context.Context, token string) (*gtsmodel.User, gtserror.WithCode) {
	user, errWithCode := p.EmailGetUserForConfirmToken(ctx, token)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if user.UnconfirmedEmail == "" ||
		user.UnconfirmedEmail == user.Email {
		// Confirmed already, just return.
		return user, nil
	}

	// Ensure token not expired.
	const oneWeek = 168 * time.Hour
	if user.ConfirmationSentAt.Before(time.Now().Add(-oneWeek)) {
		err := errors.New("confirmation token expired (older than one week)")
		return nil, gtserror.NewErrorForbidden(err, err.Error())
	}

	// Mark the user's email address as confirmed,
	// and remove the unconfirmed address and the token.
	user.Email = user.UnconfirmedEmail
	user.UnconfirmedEmail = ""
	user.ConfirmedAt = time.Now()
	user.ConfirmationToken = ""

	if err := p.state.DB.UpdateUser(
		ctx,
		user,
		"email",
		"unconfirmed_email",
		"confirmed_at",
		"confirmation_token",
	); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return user, nil
}
