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

package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// PasswordChangePOSTHandler swagger:operation POST /api/v1/user/password_change userPasswordChange
//
// Change the password of authenticated user.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
// ---
// tags:
// - user
//
// consumes:
// - application/json
// - application/xml
// - application/x-www-form-urlencoded
//
// produces:
// - application/json
//
// security:
// - OAuth2 Bearer:
//   - write:user
//
// responses:
//   '200':
//     description: Change successful
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
//   '500':
//      description: "internal error"
func (m *Module) PasswordChangePOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "PasswordChangePOSTHandler")

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("error authing: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// First check this user/account is active.
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled, not yet approved, or suspended"})
		return
	}

	form := &model.PasswordChangeRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil || form.NewPassword == "" || form.OldPassword == "" {
		if err != nil {
			l.Debugf("could not parse form from request: %s", err)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	if errWithCode := m.processor.UserChangePassword(c.Request.Context(), authed, form); errWithCode != nil {
		l.Debugf("error changing user password: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.Status(http.StatusOK)
}
