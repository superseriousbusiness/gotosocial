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
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountDeletePOSTHandler swagger:operation POST /api/v1/accounts/delete accountDelete
//
// Delete your account.
//
// ---
// tags:
// - accounts
//
// consumes:
// - multipart/form-data
//
// parameters:
// - name: password
//   in: formData
//   description: Password of the account user, for confirmation.
//   type: string
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - write:accounts
//
// responses:
//   '202':
//     description: "The account deletion has been accepted and the account will be deleted."
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
func (m *Module) AccountDeletePOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "AccountDeletePOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("retrieved account %+v", authed.Account.ID)

	form := &model.AccountDeleteRequest{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if form.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no password provided in account delete request"})
		return
	}

	form.DeleteOriginID = authed.Account.ID

	if errWithCode := m.processor.AccountDeleteLocal(c.Request.Context(), authed, form); errWithCode != nil {
		l.Debugf("could not delete account: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "accepted"})
}
