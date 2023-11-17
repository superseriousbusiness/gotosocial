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

	// Ensure request signed, and use signature URI to
	// get requesting account, dereferencing if necessary.
	pubKeyAuth, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUser)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	if pubKeyAuth.Handshaking {
		// This should happen very rarely, we are in the middle of handshaking.
		err := gtserror.Newf("network race handshaking %s", pubKeyAuth.OwnerURI)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	// Get requester from auth.
	requester := pubKeyAuth.Owner

	// Check that block does not exist between receiver and requester.
	blocked, err := p.state.DB.IsBlocked(ctx, receiver.ID, requester.ID)
	if err != nil {
		err := gtserror.Newf("error checking block: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		const text = "block exists between accounts"
		return nil, nil, gtserror.NewErrorForbidden(errors.New(text))
	}

	return requester, receiver, nil
}
