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

package auth

import (
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

type tokenBody struct {
	ClientID     *string `form:"client_id" json:"client_id" xml:"client_id"`
	ClientSecret *string `form:"client_secret" json:"client_secret" xml:"client_secret"`
	Code         *string `form:"code" json:"code" xml:"code"`
	GrantType    *string `form:"grant_type" json:"grant_type" xml:"grant_type"`
	RedirectURI  *string `form:"redirect_uri" json:"redirect_uri" xml:"redirect_uri"`
	Scope        *string `form:"scope" json:"scope" xml:"scope"`
}

// TokenPOSTHandler should be served as a POST at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
func (m *Module) TokenPOSTHandler(c *gin.Context) {
	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	form := &tokenBody{}
	if err := c.ShouldBind(form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, helpfulAdvice), m.processor.InstanceGet)
		return
	}

	c.Request.Form = url.Values{}
	if form.ClientID != nil {
		c.Request.Form.Set("client_id", *form.ClientID)
	}
	if form.ClientSecret != nil {
		c.Request.Form.Set("client_secret", *form.ClientSecret)
	}
	if form.Code != nil {
		c.Request.Form.Set("code", *form.Code)
	}
	if form.GrantType != nil {
		c.Request.Form.Set("grant_type", *form.GrantType)
	}
	if form.RedirectURI != nil {
		c.Request.Form.Set("redirect_uri", *form.RedirectURI)
	}
	if form.Scope != nil {
		c.Request.Form.Set("scope", *form.Scope)
	}

	// pass the writer and request into the oauth server handler, which will
	// take care of writing the oauth token into the response etc
	if err := m.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err, helpfulAdvice), m.processor.InstanceGet)
	}
}
