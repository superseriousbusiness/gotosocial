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

var oneWeek = 168 * time.Hour

// EmailConfirm processes an email confirmation request, usually initiated as a result of clicking on a link
// in a 'confirm your email address' type email.
func (p *Processor) EmailConfirm(ctx context.Context, token string) (*gtsmodel.User, gtserror.WithCode) {
	if token == "" {
		return nil, gtserror.NewErrorNotFound(errors.New("no token provided"))
	}

	user, err := p.state.DB.GetUserByConfirmationToken(ctx, token)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if user.Account == nil {
		a, err := p.state.DB.GetAccountByID(ctx, user.AccountID)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(err)
		}
		user.Account = a
	}

	if !user.Account.SuspendedAt.IsZero() {
		return nil, gtserror.NewErrorForbidden(fmt.Errorf("ConfirmEmail: account %s is suspended", user.AccountID))
	}

	if user.UnconfirmedEmail == "" || user.UnconfirmedEmail == user.Email {
		// no pending email confirmations so just return OK
		return user, nil
	}

	if user.ConfirmationSentAt.Before(time.Now().Add(-oneWeek)) {
		return nil, gtserror.NewErrorForbidden(errors.New("ConfirmEmail: confirmation token expired"))
	}

	// mark the user's email address as confirmed + remove the unconfirmed address and the token
	updatingColumns := []string{"email", "unconfirmed_email", "confirmed_at", "confirmation_token", "updated_at"}
	user.Email = user.UnconfirmedEmail
	user.UnconfirmedEmail = ""
	user.ConfirmedAt = time.Now()
	user.ConfirmationToken = ""
	user.UpdatedAt = time.Now()

	if err := p.state.DB.UpdateByID(ctx, user, user.ID, updatingColumns...); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return user, nil
}
