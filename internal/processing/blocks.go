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

package processing

import (
	"context"
	"fmt"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) BlocksGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, limit int) (*apimodel.BlocksResponse, gtserror.WithCode) {
	accounts, nextMaxID, prevMinID, err := p.db.GetAccountBlocks(ctx, authed.Account.ID, maxID, sinceID, limit)
	if err != nil {
		if err == db.ErrNoEntries {
			// there are just no entries
			return &apimodel.BlocksResponse{
				Accounts: []*apimodel.Account{},
			}, nil
		}
		// there's an actual error
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiAccounts := []*apimodel.Account{}
	for _, a := range accounts {
		apiAccount, err := p.tc.AccountToAPIAccountBlocked(ctx, a)
		if err != nil {
			continue
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return p.packageBlocksResponse(apiAccounts, "/api/v1/blocks", nextMaxID, prevMinID, limit)
}

func (p *processor) packageBlocksResponse(accounts []*apimodel.Account, path string, nextMaxID string, prevMinID string, limit int) (*apimodel.BlocksResponse, gtserror.WithCode) {
	resp := &apimodel.BlocksResponse{
		Accounts: []*apimodel.Account{},
	}
	resp.Accounts = accounts

	// prepare the next and previous links
	if len(accounts) != 0 {
		protocol := config.GetProtocol()
		host := config.GetHost()

		nextLink := &url.URL{
			Scheme:   protocol,
			Host:     host,
			Path:     path,
			RawQuery: fmt.Sprintf("limit=%d&max_id=%s", limit, nextMaxID),
		}
		next := fmt.Sprintf("<%s>; rel=\"next\"", nextLink.String())

		prevLink := &url.URL{
			Scheme:   protocol,
			Host:     host,
			Path:     path,
			RawQuery: fmt.Sprintf("limit=%d&min_id=%s", limit, prevMinID),
		}
		prev := fmt.Sprintf("<%s>; rel=\"prev\"", prevLink.String())
		resp.LinkHeader = fmt.Sprintf("%s, %s", next, prev)
	}

	return resp, nil
}
