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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

var oneWeek = 168 * time.Hour

func (p *processor) SendConfirmEmail(ctx context.Context, user *gtsmodel.User, username string) error {
	if user.UnconfirmedEmail == "" || user.UnconfirmedEmail == user.Email {
		// user has already confirmed this email address, so there's nothing to do
		return nil
	}

	// We need a token and a link for the user to click on.
	// We'll use a uuid as our token since it's basically impossible to guess.
	// From the uuid package we use (which uses crypto/rand under the hood):
	//      Randomly generated UUIDs have 122 random bits.  One's annual risk of being
	//      hit by a meteorite is estimated to be one chance in 17 billion, that
	//      means the probability is about 0.00000000006 (6 × 10−11),
	//      equivalent to the odds of creating a few tens of trillions of UUIDs in a
	//      year and having one duplicate.
	confirmationToken := uuid.NewString()
	confirmationLink := uris.GenerateURIForEmailConfirm(confirmationToken)

	// pull our instance entry from the database so we can greet the user nicely in the email
	instance := &gtsmodel.Instance{}
	host := config.GetHost()
	if err := p.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: host}}, instance); err != nil {
		return fmt.Errorf("SendConfirmEmail: error getting instance: %s", err)
	}

	// assemble the email contents and send the email
	confirmData := email.ConfirmData{
		Username:     username,
		InstanceURL:  instance.URI,
		InstanceName: instance.Title,
		ConfirmLink:  confirmationLink,
	}
	if err := p.emailSender.SendConfirmEmail(user.UnconfirmedEmail, confirmData); err != nil {
		return fmt.Errorf("SendConfirmEmail: error sending to email address %s belonging to user %s: %s", user.UnconfirmedEmail, username, err)
	}

	// email sent, now we need to update the user entry with the token we just sent them
	updatingColumns := []string{"confirmation_sent_at", "confirmation_token", "last_emailed_at", "updated_at"}
	user.ConfirmationSentAt = time.Now()
	user.ConfirmationToken = confirmationToken
	user.LastEmailedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := p.db.UpdateByID(ctx, user, user.ID, updatingColumns...); err != nil {
		return fmt.Errorf("SendConfirmEmail: error updating user entry after email sent: %s", err)
	}

	return nil
}

func (p *processor) ConfirmEmail(ctx context.Context, token string) (*gtsmodel.User, gtserror.WithCode) {
	if token == "" {
		return nil, gtserror.NewErrorNotFound(errors.New("no token provided"))
	}

	user, err := p.db.GetUserByConfirmationToken(ctx, token)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if user.Account == nil {
		a, err := p.db.GetAccountByID(ctx, user.AccountID)
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

	if err := p.db.UpdateByID(ctx, user, user.ID, updatingColumns...); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return user, nil
}
