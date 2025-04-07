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

	"codeberg.org/gruf/go-byteutil"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

// EmailChange processes an email address change request for the given user.
func (p *Processor) EmailChange(
	ctx context.Context,
	user *gtsmodel.User,
	password string,
	newEmail string,
) (*apimodel.User, gtserror.WithCode) {
	// Ensure provided password is correct.
	if err := bcrypt.CompareHashAndPassword(
		byteutil.S2B(user.EncryptedPassword),
		byteutil.S2B(password),
	); err != nil {
		err := gtserror.Newf("%w", err)
		return nil, gtserror.NewErrorUnauthorized(err, "password was incorrect")
	}

	// Ensure new email address is valid.
	if err := validate.Email(newEmail); err != nil {
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure new email address is different
	// from current email address.
	if newEmail == user.Email {
		const help = "new email address cannot be the same as current email address"
		err := gtserror.New(help)
		return nil, gtserror.NewErrorBadRequest(err, help)
	}

	if newEmail == user.UnconfirmedEmail {
		const help = "you already have an email change request pending for given email address"
		err := gtserror.New(help)
		return nil, gtserror.NewErrorBadRequest(err, help)
	}

	// Ensure this address isn't already used by another account.
	emailAvailable, err := p.state.DB.IsEmailAvailable(ctx, newEmail)
	if err != nil {
		err := gtserror.Newf("db error checking email availability: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !emailAvailable {
		const help = "new email address is already in use on this instance"
		err := gtserror.New(help)
		return nil, gtserror.NewErrorConflict(err, help)
	}

	// Set new email address on user.
	user.UnconfirmedEmail = newEmail
	if err := p.state.DB.UpdateUser(
		ctx, user,
		"unconfirmed_email",
	); err != nil {
		err := gtserror.Newf("db error updating user: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure user populated (we need account).
	if err := p.state.DB.PopulateUser(ctx, user); err != nil {
		err := gtserror.Newf("db error populating user: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Add email sending job to the queue.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		// Use ap.ObjectProfile here to
		// distinguish this message (user model)
		// from ap.ActorPerson (account model).
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       user,
		Origin:         user.Account,
		Target:         user.Account,
	})

	return p.converter.UserToAPIUser(ctx, user), nil
}

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
		if err != nil {
			// We need the account for a local user.
			return nil, gtserror.NewErrorInternalError(err)
		}
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
