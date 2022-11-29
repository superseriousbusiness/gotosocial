/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (p *processor) GetStatus(ctx context.Context, requestedUsername string, requestedStatusID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	// get the account the request is referring to
	requestedAccount, err := p.db.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccountURI, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	requestingAccount, err := p.federator.GetAccount(transport.WithFastfail(ctx), dereferencing.GetAccountParams{
		RequestingUsername: requestedUsername,
		RemoteAccountID:    requestingAccountURI,
	})
	if err != nil {
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	// authorize the request:
	// 1. check if a block exists between the requester and the requestee
	blocked, err := p.db.IsBlocked(ctx, requestedAccount.ID, requestingAccount.ID, true)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	// get the status out of the database here
	s, err := p.db.GetStatusByID(ctx, requestedStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting status with id %s and account id %s: %s", requestedStatusID, requestedAccount.ID, err))
	}

	if s.AccountID != requestedAccount.ID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s does not belong to account with id %s", s.ID, requestedAccount.ID))
	}

	visible, err := p.filter.StatusVisible(ctx, s, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s not visible to user with id %s", s.ID, requestingAccount.ID))
	}

	// requester is authorized to view the status, so convert it to AP representation and serialize it
	asStatus, err := p.tc.StatusToAS(ctx, s)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := streams.Serialize(asStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
