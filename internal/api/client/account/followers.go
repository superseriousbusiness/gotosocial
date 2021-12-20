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

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountFollowersGETHandler swagger:operation GET /api/v1/accounts/{id}/followers accountFollowers
//
// See followers of account with given id.
//
// ---
// tags:
// - accounts
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   type: string
//   description: Account ID.
//   in: path
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - read:accounts
//
// responses:
//   '200':
//     name: accounts
//     description: Array of accounts that follow this account.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/account"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) AccountFollowersGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
		return
	}

	followers, errWithCode := m.processor.AccountFollowersGet(c.Request.Context(), authed, targetAcctID)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, followers)
}
