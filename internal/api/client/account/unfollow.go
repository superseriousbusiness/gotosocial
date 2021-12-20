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

// AccountUnfollowPOSTHandler swagger:operation POST /api/v1/accounts/{id}/unfollow accountUnfollow
//
// Unfollow account with id.
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
//   description: The id of the account to unfollow.
//   in: path
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - write:follows
//
// responses:
//   '200':
//     name: account relationship
//     description: Your relationship to this account.
//     schema:
//       "$ref": "#/definitions/accountRelationship"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) AccountUnfollowPOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "AccountUnfollowPOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debug(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		l.Debug(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
		return
	}

	relationship, errWithCode := m.processor.AccountFollowRemove(c.Request.Context(), authed, targetAcctID)
	if errWithCode != nil {
		l.Debug(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, relationship)
}
