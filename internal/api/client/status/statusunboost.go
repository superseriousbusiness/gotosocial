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

// StatusUnboostPOSTHandler handles unboost requests against a given status ID
func (m *Module) StatusUnboostPOSTHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "StatusUnboostPOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.Authed(c, true, false, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Debug("not authed so can't unboost status")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	targetStatusID := c.Param(IDKey)
	if targetStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id provided"})
		return
	}

	mastoStatus, errWithCode := m.processor.StatusUnboost(authed, targetStatusID)
	if errWithCode != nil {
		l.Debugf("error processing status unboost: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, mastoStatus)
}
