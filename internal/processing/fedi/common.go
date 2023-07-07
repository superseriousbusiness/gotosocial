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

package fedi

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *Processor) authenticate(ctx context.Context, requestedUsername string) (
	*gtsmodel.Account, // requestedAccount
	*gtsmodel.Account, // requestingAccount
	gtserror.WithCode,
) {
	// Get LOCAL account with the requested username.
	requestedAccount, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = gtserror.Newf("db error getting account %s: %w", requestedUsername, err)
			return nil, nil, gtserror.NewErrorInternalError(err)
		}

		// Account just not found in the db.
		return nil, nil, gtserror.NewErrorNotFound(err)
	}

	// Ensure request signed, and use signature URI to
	// get requesting account, dereferencing if necessary.
	requestingAccountURI, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	requestingAccount, _, err := p.federator.GetAccountByURI(
		gtscontext.SetFastFail(ctx),
		requestedUsername,
		requestingAccountURI,
	)
	if err != nil {
		err = gtserror.Newf("error getting account %s: %w", requestingAccountURI, err)
		return nil, nil, gtserror.NewErrorUnauthorized(err)
	}

	// Ensure no block exists between requester + requested.
	blocked, err := p.state.DB.IsEitherBlocked(ctx, requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		err = gtserror.Newf("db error getting checking block: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		err = fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID)
		return nil, nil, gtserror.NewErrorUnauthorized(err)
	}

	return requestedAccount, requestingAccount, nil
}
