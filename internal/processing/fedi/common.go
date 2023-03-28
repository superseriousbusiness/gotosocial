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
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (p *Processor) authenticate(ctx context.Context, requestedUsername string) (requestedAccount, requestingAccount *gtsmodel.Account, errWithCode gtserror.WithCode) {
	requestedAccount, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		errWithCode = gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
		return
	}

	var requestingAccountURI *url.URL
	requestingAccountURI, errWithCode = p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return
	}

	if requestingAccount, err = p.federator.GetAccountByURI(transport.WithFastfail(ctx), requestedUsername, requestingAccountURI, false); err != nil {
		errWithCode = gtserror.NewErrorUnauthorized(err)
		return
	}

	blocked, err := p.state.DB.IsEitherBlocked(ctx, requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		errWithCode = gtserror.NewErrorInternalError(err)
		return
	}

	if blocked {
		errWithCode = gtserror.NewErrorUnauthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	return
}
