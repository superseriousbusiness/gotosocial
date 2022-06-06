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

package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// StatusBoostedByGETHandler swagger:operation GET /api/v1/statuses/{id}/reblogged_by statusBoostedBy
//
// View accounts that have reblogged/boosted the target status.
//
// ---
// tags:
// - statuses
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   type: string
//   description: Target status ID.
//   in: path
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - read:accounts
//
// responses:
//   '200':
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/account"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '404':
//      description: not found
func (m *Module) StatusBoostedByGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "StatusBoostedByGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.Authed(c, true, true, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Errorf("error authing status boosted by request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "not authed"})
		return
	}

	targetStatusID := c.Param(IDKey)
	if targetStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id provided"})
		return
	}

	apiAccounts, errWithCode := m.processor.StatusBoostedBy(c.Request.Context(), authed, targetStatusID)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, apiAccounts)
}
