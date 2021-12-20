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

package media

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// MediaGETHandler swagger:operation GET /api/v1/media/{id} mediaGet
//
// Get a media attachment that you own.
//
// ---
// tags:
// - media
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   description: id of the attachment
//   type: string
//   in: path
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - read:media
//
// responses:
//   '200':
//     description: The requested media attachment.
//     schema:
//       "$ref": "#/definitions/attachment"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '422':
//      description: unprocessable
func (m *Module) MediaGETHandler(c *gin.Context) {
	l := logrus.WithField("func", "MediaGETHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	attachmentID := c.Param(IDKey)
	if attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no attachment ID given in request"})
		return
	}

	attachment, errWithCode := m.processor.MediaGet(c.Request.Context(), authed, attachmentID)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, attachment)
}
