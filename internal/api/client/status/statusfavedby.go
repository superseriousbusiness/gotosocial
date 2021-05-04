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

package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// StatusFavedByGETHandler is for serving a list of accounts that have faved a given status
func (m *Module) StatusFavedByGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "statusGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.Authed(c, false, false, false, false) // we don't really need an app here but we want everything else
	if err != nil {
		l.Errorf("error authing status faved by request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "not authed"})
		return
	}

	targetStatusID := c.Param(IDKey)
	if targetStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id provided"})
		return
	}

	mastoAccounts, err := m.processor.StatusFavedBy(authed, targetStatusID)
	if err != nil {
		l.Debugf("error processing status faved by request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	c.JSON(http.StatusOK, mastoAccounts)
}
