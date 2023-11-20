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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *Processor) authenticate(ctx context.Context, requestedUser string) (
	*gtsmodel.Account, // requester: i.e. user making the request
	*gtsmodel.Account, // receiver: i.e. the receiving inbox user
	gtserror.WithCode,
) {
	// First get the requested (receiving) LOCAL account with username from database.
	receiver, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUser, "")
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = gtserror.Newf("db error getting account %s: %w", requestedUser, err)
			return nil, nil, gtserror.NewErrorInternalError(err)
		}

		// Account just not found in the db.
		return nil, nil, gtserror.NewErrorNotFound(err)
	}

	var requester *gtsmodel.Account

	// Ensure request signed, and use signature URI to
	// get requesting account, dereferencing if necessary.
	pubKeyAuth, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUser)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	if requester = pubKeyAuth.Owner; requester == nil {
		requester, _, err = p.federator.GetAccountByURI(
			gtscontext.SetFastFail(ctx),
			requestedUser,
			pubKeyAuth.OwnerURI,
		)
		if err != nil {
			err = gtserror.Newf("error getting account %s: %w", pubKeyAuth.OwnerURI, err)
			return nil, nil, gtserror.NewErrorUnauthorized(err)
		}
	}

	if !requester.SuspendedAt.IsZero() {
		// Account was marked as suspended by a
		// local admin action. Stop request early.
		const text = "requesting account is suspended"
		return nil, nil, gtserror.NewErrorForbidden(errors.New(text))
	}

	// Ensure no block exists between requester + requested.
	blocked, err := p.state.DB.IsEitherBlocked(ctx, receiver.ID, requester.ID)
	if err != nil {
		err = gtserror.Newf("db error getting checking block: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		err = gtserror.Newf("block exists between accounts %s and %s", requester.ID, receiver.ID)
		return nil, nil, gtserror.NewErrorForbidden(err)
	}

	return requester, receiver, nil
}
