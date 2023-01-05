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

package streaming

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) AuthorizeStreamingRequest(ctx context.Context, accessToken string) (*gtsmodel.Account, gtserror.WithCode) {
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

	user, err := p.db.GetUserByID(ctx, uid)
	if err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("no user found for validated uid %s", uid)
			return nil, gtserror.NewErrorUnauthorized(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	acct, err := p.db.GetAccountByID(ctx, user.AccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("no account found for validated uid %s", uid)
			return nil, gtserror.NewErrorUnauthorized(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	return acct, nil
}
