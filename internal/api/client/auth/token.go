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
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"

	"github.com/gin-gonic/gin"
)

// TokenRequestForm models a token request
type TokenRequestForm struct {
	ClientID     string      `form:"client_id" json:"client_id" xml:"client_id"`
	ClientSecret string      `form:"client_secret" json:"client_secret" xml:"client_secret"`
	Code         string      `form:"code" json:"code" xml:"code"`
	GrantType    string      `form:"grant_type" json:"grant_type" xml:"grant_type"`
	RedirectURI  string      `form:"redirect_uri" json:"redirect_uri" xml:"redirect_uri"`
	Scope        interface{} `form:"scope" json:"scope" xml:"scope"` // scope might be an array or a string
}

func (m *Module) BindTokenRequestForm(c *gin.Context) error {
	form := &TokenRequestForm{}

	if err := c.ShouldBind(form); err != nil {
		return err
	}

	// assign values onto the request form
	if c.Request.Form == nil {
		c.Request.Form = url.Values{}
	}
	c.Request.Form.Set("client_id", form.ClientID)
	c.Request.Form.Set("client_secret", form.ClientSecret)
	c.Request.Form.Set("code", form.Code)
	c.Request.Form.Set("grant_type", form.GrantType)
	c.Request.Form.Set("redirect_uri", form.RedirectURI)

	// check if scope is a string
	if scope, ok := form.Scope.(string); ok {
		c.Request.Form.Set("scope", scope)
	}

	// check if scopeI is a slice of strings
	if scopeI, ok := form.Scope.([]interface{}); ok {
		var scope []string
		for _, sI := range scopeI {
			if scopeValue, ok := sI.(string); ok {
				scope = append(scope, scopeValue)
			}
		}
		c.Request.Form.Set("scope", strings.Join(scope, " "))
	}

	return nil
}

// TokenPOSTHandler should be served as a POST at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
func (m *Module) TokenPOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "TokenPOSTHandler")
	l.Trace("entered TokenPOSTHandler")

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	if err := m.BindTokenRequestForm(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := m.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
