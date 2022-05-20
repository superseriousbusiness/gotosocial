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
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// MediaCleanupPOSTHandler swagger:operation POST /api/v1/admin/media_cleanup mediaCleanup
//
// Clean up remote media older than the specified number of days.
// Also cleans up unused headers + avatars from the media cache.
//
// ---
// tags:
// - admin
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
//   - admin
//
// responses:
//   '200':
//     description: |-
//      Echos the number of days requested. The cleanup is performed asynchronously after the request completes.
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
func (m *Module) MediaCleanupPOSTHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "MediaCleanupPOSTHandler",
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
	form := &model.MediaCleanupRequest{}
	if err := c.ShouldBind(form); err != nil {
		l.Debugf("error parsing form %+v: %s", c.Request.Form, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not parse form: %s", err)})
		return
	}

	var remoteCacheDays int
	if form.RemoteCacheDays == nil {
		remoteCacheDays = viper.GetInt(config.Keys.MediaRemoteCacheDays)
	} else {
		remoteCacheDays = *form.RemoteCacheDays
	}
	if remoteCacheDays < 0 {
		remoteCacheDays = 0
	}

	if errWithCode := m.processor.AdminMediaPrune(c.Request.Context(), remoteCacheDays); errWithCode != nil {
		l.Debugf("error starting prune of remote media: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, remoteCacheDays)
}
