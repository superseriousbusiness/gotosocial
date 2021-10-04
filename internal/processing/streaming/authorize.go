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

package streaming

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) AuthorizeStreamingRequest(ctx context.Context, accessToken string) (*gtsmodel.Account, error) {
	ti, err := p.oauthServer.LoadAccessToken(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: error loading access token: %s", err)
	}

	uid := ti.GetUserID()
	if uid == "" {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: no userid in token")
	}

	// fetch user's and account for this user id
	user := &gtsmodel.User{}
	if err := p.db.GetByID(ctx, uid, user); err != nil || user == nil {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: no user found for validated uid %s", uid)
	}

	acct, err := p.db.GetAccountByID(ctx, user.AccountID)
	if err != nil || acct == nil {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: no account retrieved for user with id %s", uid)
	}

	return acct, nil
}
