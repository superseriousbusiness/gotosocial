/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) GetUser(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	// get the account the request is referring to
	requestedAccount, err := p.db.GetLocalAccountByUsername(ctx, requestedUsername)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	var requestedPerson vocab.ActivityStreamsPerson
	if util.IsPublicKeyPath(requestURL) {
		// if it's a public key path, we don't need to authenticate but we'll only serve the bare minimum user profile needed for the public key
		requestedPerson, err = p.tc.AccountToASMinimal(ctx, requestedAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else if util.IsUserPath(requestURL) {
		// if it's a user path, we want to fully authenticate the request before we serve any data, and then we can serve a more complete profile
		requestingAccountURI, authenticated, err := p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
		if err != nil || !authenticated {
			return nil, gtserror.NewErrorNotAuthorized(errors.New("not authorized"), "not authorized")
		}

		// if we're not already handshaking/dereferencing a remote account, dereference it now
		if !p.federator.Handshaking(ctx, requestedUsername, requestingAccountURI) {
			requestingAccount, _, err := p.federator.GetRemoteAccount(ctx, requestedUsername, requestingAccountURI, false)
			if err != nil {
				return nil, gtserror.NewErrorNotAuthorized(err)
			}

			blocked, err := p.db.IsBlocked(ctx, requestedAccount.ID, requestingAccount.ID, true)
			if err != nil {
				return nil, gtserror.NewErrorInternalError(err)
			}

			if blocked {
				return nil, gtserror.NewErrorNotAuthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
			}
		}

		requestedPerson, err = p.tc.AccountToAS(ctx, requestedAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else {
		return nil, gtserror.NewErrorBadRequest(fmt.Errorf("path was not public key path or user path"))
	}

	data, err := streams.Serialize(requestedPerson)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
