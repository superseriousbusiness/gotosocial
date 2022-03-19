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

package admin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountActionPOSTHandler swagger:operation POST /api/v1/admin/accounts/{id}/action adminAccountAction
//
// Perform an admin action on an account.
//
// ---
// tags:
// - admin
//
// consumes:
// - multipart/form-data
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   required: true
//   in: path
//   description: ID of the account.
//   type: string
// - name: type
//   in: formData
//   description: |-
//     Type of action to be taken. One of: disable, silence, suspend.
//   type: string
//   required: true
// - name: text
//   in: formData
//   description: Optional text describing why this action was taken.
//   type: string
//
// security:
// - OAuth2 Bearer:
//   - admin
//
// responses:
//   '200':
//     description: OK
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
func (m *Module) AccountActionPOSTHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "AccountActionPOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})

	// make sure we're authed...
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// with an admin account
	if !authed.User.Admin {
		l.Debugf("user %s not an admin", authed.User.ID)
		c.JSON(http.StatusForbidden, gin.H{"error": "not an admin"})
		return
	}

	// extract the form from the request context
	l.Tracef("parsing request form: %+v", c.Request.Form)
	form := &model.AdminAccountActionRequest{}
	if err := c.ShouldBind(form); err != nil {
		l.Debugf("error parsing form %+v: %s", c.Request.Form, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not parse form: %s", err)})
		return
	}

	if form.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no type specified"})
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
		return
	}
	form.TargetAccountID = targetAcctID

	if errWithCode := m.processor.AdminAccountAction(c.Request.Context(), authed, form); errWithCode != nil {
		l.Debugf("error performing account action: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
