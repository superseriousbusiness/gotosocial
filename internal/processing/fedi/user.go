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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
)

// UserGet handles the getting of a fedi/activitypub representation of a user/account,
// performing authentication before returning a JSON serializable interface to the caller.
func (p *Processor) UserGet(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	// (Try to) get the requested local account from the db.
	receiver, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// Account just not found w/ this username.
			err := fmt.Errorf("account with username %s not found in the db", requestedUsername)
			return nil, gtserror.NewErrorNotFound(err)
		}

		// Real db error.
		err := fmt.Errorf("db error getting account with username %s: %w", requestedUsername, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if uris.IsPublicKeyPath(requestURL) {
		// If request is on a public key path, we don't need to
		// authenticate this request. However, we'll only serve
		// the bare minimum user profile needed for the pubkey.
		//
		// TODO: https://codeberg.org/superseriousbusiness/gotosocial/issues/1186
		minimalPerson, err := p.converter.AccountToASMinimal(ctx, receiver)
		if err != nil {
			err := gtserror.Newf("error converting to minimal account: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Return early with bare minimum data.
		return data(minimalPerson)
	}

	// If the request is not on a public key path, we want to
	// try to authenticate it before we serve any data, so that
	// we can serve a more complete profile.
	pubKeyAuth, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode // likely 401
	}

	// Auth passed, generate the proper AP representation.
	accountable, err := p.converter.AccountToAS(ctx, receiver)
	if err != nil {
		err := gtserror.Newf("error converting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if pubKeyAuth.Handshaking {
		// If we are currently handshaking with the remote account
		// making the request, then don't be coy: just serve the AP
		// representation of the target account.
		//
		// This handshake check ensures that we don't get stuck in
		// a loop with another GtS instance, where each instance is
		// trying repeatedly to dereference the other account that's
		// making the request before it will reveal its own account.
		//
		// Instead, we end up in an 'I'll show you mine if you show me
		// yours' situation, where we sort of agree to reveal each
		// other's profiles at the same time.
		return data(accountable)
	}

	// Get requester from auth.
	requester := pubKeyAuth.Owner

	// Check that block does not exist between receiver and requester.
	blocked, err := p.state.DB.IsBlocked(ctx, receiver.ID, requester.ID)
	if err != nil {
		err := gtserror.Newf("error checking block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		const text = "block exists between accounts"
		return nil, gtserror.NewErrorForbidden(errors.New(text))
	}

	return data(accountable)
}

func data(accountable ap.Accountable) (interface{}, gtserror.WithCode) {
	data, err := ap.Serialize(accountable)
	if err != nil {
		err := gtserror.Newf("error serializing accountable: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
