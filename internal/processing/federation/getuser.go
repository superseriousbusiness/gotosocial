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

package federation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) GetUser(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	// Get the instance-local account the request is referring to.
	requestedAccount, err := p.db.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	var requestedPerson vocab.ActivityStreamsPerson

	if uris.IsPublicKeyPath(requestURL) {
		// if it's a public key path, we don't need to authenticate but we'll only serve the bare minimum user profile needed for the public key
		requestedPerson, err = p.tc.AccountToASMinimal(ctx, requestedAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else {
		// if it's any other path, we want to fully authenticate the request before we serve any data, and then we can serve a more complete profile
		requestingAccountURI, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
		if errWithCode != nil {
			return nil, errWithCode
		}

		// if we're not already handshaking/dereferencing a remote account, dereference it now
		if !p.federator.Handshaking(requestedUsername, requestingAccountURI) {
			requestingAccount, err := p.federator.GetAccountByURI(
				transport.WithFastfail(ctx), requestedUsername, requestingAccountURI, false,
			)
			if err != nil {
				return nil, gtserror.NewErrorUnauthorized(err)
			}

			blocked, err := p.db.IsBlocked(ctx, requestedAccount.ID, requestingAccount.ID, true)
			if err != nil {
				return nil, gtserror.NewErrorInternalError(err)
			}

			if blocked {
				return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
			}
		}

		requestedPerson, err = p.tc.AccountToAS(ctx, requestedAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	data, err := streams.Serialize(requestedPerson)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
