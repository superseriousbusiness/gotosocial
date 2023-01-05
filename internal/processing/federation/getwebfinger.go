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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

const (
	webfingerProfilePage            = "http://webfinger.net/rel/profile-page"
	webFingerProfilePageContentType = "text/html"
	webfingerSelf                   = "self"
	webFingerSelfContentType        = "application/activity+json"
	webfingerAccount                = "acct"
)

func (p *processor) GetWebfingerAccount(ctx context.Context, requestedUsername string) (*apimodel.WellKnownResponse, gtserror.WithCode) {
	// get the account the request is referring to
	requestedAccount, err := p.db.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	accountDomain := config.GetAccountDomain()
	if accountDomain == "" {
		accountDomain = config.GetHost()
	}

	// return the webfinger representation
	return &apimodel.WellKnownResponse{
		Subject: fmt.Sprintf("%s:%s@%s", webfingerAccount, requestedAccount.Username, accountDomain),
		Aliases: []string{
			requestedAccount.URI,
			requestedAccount.URL,
		},
		Links: []apimodel.Link{
			{
				Rel:  webfingerProfilePage,
				Type: webFingerProfilePageContentType,
				Href: requestedAccount.URL,
			},
			{
				Rel:  webfingerSelf,
				Type: webFingerSelfContentType,
				Href: requestedAccount.URI,
			},
		},
	}, nil
}
