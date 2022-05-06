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

package account

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountVerifyGETHandler swagger:operation GET /api/v1/accounts/verify_credentials accountVerify
//
// Verify a token by returning account details pertaining to it.
//
// ---
// tags:
// - accounts
//
// produces:
// - application/json
//
// security:
// - OAuth2 Bearer:
//   - read:accounts
//
// responses:
//   '200':
//     schema:
//       "$ref": "#/definitions/account"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) AccountVerifyGETHandler(c *gin.Context) {
	l := logrus.WithField("func", "accountVerifyGETHandler")
	authed, err := oauth.Authed(c, true, false, false, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	acctSensitive, err := m.processor.AccountGet(c.Request.Context(), authed, authed.Account.ID)
	if err != nil {
		l.Debugf("error getting account from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, acctSensitive)
}
