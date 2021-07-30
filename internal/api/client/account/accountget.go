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

package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountGETHandler serves account information in response to a GET request.
//
// swagger:operation GET /api/v1/accounts/{id} accountGET
//
// ---
// produces:
// - application/json
//
// parameters:
// - name: id
//   in: path
//   required: true
//   type: string
//
// security:
// - oauth2:
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
func (m *Module) AccountGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
		return
	}

	acctInfo, err := m.processor.AccountGet(authed, targetAcctID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, acctInfo)
}
