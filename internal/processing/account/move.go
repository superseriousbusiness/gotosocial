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

package account

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"golang.org/x/crypto/bcrypt"
)

func (p *Processor) MoveSelf(
	ctx context.Context,
	authed *oauth.Auth,
	form *apimodel.AccountMoveRequest,
) gtserror.WithCode {
	// Ensure valid MovedToURI.
	if form.MovedToURI == "" {
		err := errors.New("no moved_to_uri provided in account Move request")
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	movedToURI, err := url.Parse(form.MovedToURI)
	if err != nil {
		err := fmt.Errorf("invalid moved_to_uri provided in account Move request: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	if movedToURI.Scheme != "https" && movedToURI.Scheme != "http" {
		err := errors.New("invalid moved_to_uri provided in account Move request: uri scheme must be http or https")
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Self account Move requires password to ensure it's for real.
	if form.Password == "" {
		err := errors.New("no password provided in account Move request")
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(authed.User.EncryptedPassword),
		[]byte(form.Password),
	); err != nil {
		err := errors.New("invalid password provided in account Move request")
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	var (
		// Current account from which
		// the move is taking place.
		account = authed.Account

		// Target account to which
		// the move is taking place.
		targetAccount *gtsmodel.Account
	)

	switch {
	case account.MovedToURI == "":
		// No problemo.

	case account.MovedToURI == form.MovedToURI:
		// Trying to move again to the same
		// destination, perhaps to reprocess
		// side effects. This is OK.
		log.Info(ctx,
			"reprocessing Move side effects from %s to %s",
			account.URI, form.MovedToURI,
		)

	default:
		// Account already moved, and now
		// trying to move somewhere else.
		err := fmt.Errorf(
			"account %s is already Moved to %s, cannot also Move to %s",
			account.URI, account.MovedToURI, form.MovedToURI,
		)
		return gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Ensure we have a valid, up-to-date representation of the target account.
	targetAccount, _, err = p.federator.GetAccountByURI(ctx, account.Username, movedToURI)
	if err != nil {
		err := fmt.Errorf("error dereferencing moved_to_uri account: %w", err)
		return gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	if !targetAccount.SuspendedAt.IsZero() {
		err := fmt.Errorf(
			"target account %s is suspended from this instance; "+
				"you will not be able to Move to that account",
			targetAccount.URI,
		)
		return gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Target account MUST be aliased to this
	// account for this to be a valid Move.
	if !slices.Contains(targetAccount.AlsoKnownAsURIs, account.URI) {
		err := fmt.Errorf(
			"target account %s is not aliased to this account via alsoKnownAs; "+
				"if you just changed it, wait five minutes and try the Move again",
			targetAccount.URI,
		)
		return gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Target account cannot itself have
	// already Moved somewhere else.
	if targetAccount.MovedToURI != "" {
		err := fmt.Errorf(
			"target account %s has already Moved somewhere else (%s); "+
				"you will not be able to Move to that account",
			targetAccount.URI, targetAccount.MovedToURI,
		)
		return gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Everything seems OK, so process the Move.
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityMove,
		OriginAccount:  account,
		TargetAccount:  targetAccount,
	})

	return nil
}
