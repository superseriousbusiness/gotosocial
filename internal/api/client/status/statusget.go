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

// StatusGETHandler swagger:operation GET /api/v1/statuses/{id} statusGet
//
// View status with the given ID.
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
//   - read:statuses
//
// responses:
//   '200':
//     description: "The requested created status."
//     schema:
//       "$ref": "#/definitions/status"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
//   '500':
//      description: internal error
func (m *Module) StatusGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "statusGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		l.Errorf("error authing status faved by request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "not authed"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	targetStatusID := c.Param(IDKey)
	if targetStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id provided"})
		return
	}

	apiStatus, err := m.processor.StatusGet(c.Request.Context(), authed, targetStatusID)
	if err != nil {
		l.Debugf("error processing status get: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	c.JSON(http.StatusOK, apiStatus)
}
