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
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

type commonAuth struct {
	handshakingURI *url.URL          // Set to requestingAcct's URI if we're currently handshaking them.
	requestingAcct *gtsmodel.Account // Remote account making request to this instance.
	receivingAcct  *gtsmodel.Account // Local account receiving the request.
}

func (p *Processor) authenticate(ctx context.Context, requestedUser string) (*commonAuth, gtserror.WithCode) {
	// First get the requested (receiving) LOCAL account with username from database.
	receiver, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUser, "")
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = gtserror.Newf("db error getting account %s: %w", requestedUser, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Account just not found in the db.
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Ensure request signed, and use signature URI to
	// get requesting account, dereferencing if necessary.
	pubKeyAuth, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if pubKeyAuth.Handshaking {
		// We're still handshaking so we
		// don't know the requester yet.
		return &commonAuth{
			handshakingURI: pubKeyAuth.OwnerURI,
			receivingAcct:  receiver,
		}, nil
	}

	// Get requester from auth.
	requester := pubKeyAuth.Owner

	// Ensure block does not exist between receiver and requester.
	blocked, err := p.state.DB.IsEitherBlocked(ctx, receiver.ID, requester.ID)
	if err != nil {
		err := gtserror.Newf("error checking block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		const text = "block exists between accounts"
		return nil, gtserror.NewErrorForbidden(errors.New(text))
	}

	return &commonAuth{
		requestingAcct: requester,
		receivingAcct:  receiver,
	}, nil
}

// validateIntReqRequest is a shortcut function
// for returning an accepted interaction request
// targeting `requestedUser`.
func (p *Processor) validateIntReqRequest(
	ctx context.Context,
	requestedUser string,
	intReqID string,
) (*gtsmodel.InteractionRequest, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if auth.handshakingURI != nil {
		// We're currently handshaking, which means we don't know
		// this account yet. This should be a very rare race condition.
		err := gtserror.Newf("network race handshaking %s", auth.handshakingURI)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Fetch interaction request with the given ID.
	req, err := p.state.DB.GetInteractionRequestByID(ctx, intReqID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request %s: %w", intReqID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure that this is an existing
	// and *accepted* interaction request.
	if req == nil || !req.IsAccepted() {
		const text = "interaction request not found"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// Ensure interaction request was accepted
	// by the account in the request path.
	if req.TargetAccountID != auth.receivingAcct.ID {
		text := fmt.Sprintf(
			"account %s is not targeted by interaction request %s and therefore can't accept it",
			requestedUser, intReqID,
		)
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// All fine.
	return req, nil
}
