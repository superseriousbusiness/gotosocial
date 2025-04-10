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

package processing

import (
	"context"
	"net/http"

	"codeberg.org/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *Processor) OAuthHandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) gtserror.WithCode {
	// todo: some kind of metrics stuff here
	return p.oauthServer.HandleAuthorizeRequest(w, r)
}

func (p *Processor) OAuthHandleTokenRequest(r *http.Request) (map[string]interface{}, gtserror.WithCode) {
	// todo: some kind of metrics stuff here
	return p.oauthServer.HandleTokenRequest(r)
}

func (p *Processor) OAuthValidateBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	// todo: some kind of metrics stuff here
	return p.oauthServer.ValidationBearerToken(r)
}

func (p *Processor) OAuthRevokeAccessToken(
	ctx context.Context,
	clientID string,
	clientSecret string,
	accessToken string,
) gtserror.WithCode {
	return p.oauthServer.RevokeAccessToken(
		ctx,
		clientID,
		clientSecret,
		accessToken,
	)
}
