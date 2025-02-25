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

package stream

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Authorize returns an oauth2 token info in response to an access token query from the streaming API
func (p *Processor) Authorize(ctx context.Context, accessToken string) (*gtsmodel.Account, gtserror.WithCode) {
	ti, err := p.oauthServer.LoadAccessToken(ctx, accessToken)
	if err != nil {
		err := fmt.Errorf("could not load access token: %s", err)
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	uid := ti.GetUserID()
	if uid == "" {
		err := fmt.Errorf("no userid in token")
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	user, err := p.state.DB.GetUserByID(ctx, uid)
	if err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("no user found for validated uid %s", uid)
			return nil, gtserror.NewErrorUnauthorized(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	acct, err := p.state.DB.GetAccountByID(ctx, user.AccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("no account found for validated uid %s", uid)
			return nil, gtserror.NewErrorUnauthorized(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure read scope.
	//
	// TODO: make this more granular
	// depending on stream type.
	hasScopes := strings.Split(ti.GetScope(), " ")
	scopeOK := slices.ContainsFunc(
		hasScopes,
		func(hasScope string) bool {
			return apiutil.Scope(hasScope).Permits(apiutil.ScopeRead)
		},
	)

	if !scopeOK {
		const errText = "token has insufficient scope permission"
		return nil, gtserror.NewErrorForbidden(errors.New(errText), errText)
	}

	return acct, nil
}
