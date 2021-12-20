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

package app

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

const (
	// permitted length for most fields
	formFieldLen = 64
	// redirect can be a bit bigger because we probably need to encode data in the redirect uri
	formRedirectLen = 512
)

// AppsPOSTHandler swagger:operation POST /api/v1/apps appCreate
//
// Register a new application on this instance.
//
// The registered application can be used to obtain an application token.
// This can then be used to register a new account, or (through user auth) obtain an access token.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
// ---
// tags:
// - apps
//
// consumes:
// - application/json
// - application/xml
// - application/x-www-form-urlencoded
//
// produces:
// - application/json
//
// responses:
//   '200':
//     description: "The newly-created application."
//     schema:
//       "$ref": "#/definitions/application"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '422':
//      description: unprocessable
//   '500':
//      description: internal error
func (m *Module) AppsPOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "AppsPOSTHandler")
	l.Trace("entering AppsPOSTHandler")

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	form := &model.ApplicationCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// check lengths of fields before proceeding so the user can't spam huge entries into the database
	if len(form.ClientName) > formFieldLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("client_name must be less than %d bytes", formFieldLen)})
		return
	}
	if len(form.Website) > formFieldLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("website must be less than %d bytes", formFieldLen)})
		return
	}
	if len(form.RedirectURIs) > formRedirectLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("redirect_uris must be less than %d bytes", formRedirectLen)})
		return
	}
	if len(form.Scopes) > formFieldLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scopes must be less than %d bytes", formFieldLen)})
		return
	}

	apiApp, err := m.processor.AppCreate(c.Request.Context(), authed, form)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiApp)
}
