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

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// UserGet handles the getting of a fedi/activitypub representation of a user/account,
// performing authentication before returning a JSON serializable interface to the caller.
func (p *Processor) UserGet(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	// (Try to) get the requested local account from the db.
	requestedAccount, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUsername, "")
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
		// TODO: https://github.com/superseriousbusiness/gotosocial/issues/1186
		minimalPerson, err := p.converter.AccountToASMinimal(ctx, requestedAccount)
		if err != nil {
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
	person, err := p.converter.AccountToAS(ctx, requestedAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

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
	if p.federator.Handshaking(requestedUsername, pubKeyAuth.OwnerURI) {
		return data(person)
	}

	// We're not currently handshaking with the requestingAccountURI,
	// so fetch its details and ensure it's up to date + not blocked.
	requestingAccount, _, err := p.federator.GetAccountByURI(
		// On a hot path so fail quickly.
		gtscontext.SetFastFail(ctx),
		requestedUsername,
		pubKeyAuth.OwnerURI,
	)
	if err != nil {
		err := gtserror.Newf("error getting account %s: %w", pubKeyAuth.OwnerURI, err)
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	blocked, err := p.state.DB.IsBlocked(ctx, requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		err := gtserror.Newf("error checking block from account %s to account %s: %w", requestedAccount.ID, requestingAccount.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		err := fmt.Errorf("account %s blocks account %s", requestedAccount.ID, requestingAccount.ID)
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	return data(person)
}

func data(requestedPerson vocab.ActivityStreamsPerson) (interface{}, gtserror.WithCode) {
	data, err := ap.Serialize(requestedPerson)
	if err != nil {
		err := gtserror.Newf("error serializing person: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
