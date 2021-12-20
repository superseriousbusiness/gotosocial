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

package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// NotificationsGETHandler serves a list of notifications to the caller, with the desired query parameters
func (m *Module) NotificationsGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "NotificationsGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.Authed(c, true, true, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Errorf("error authing status faved by request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "not authed"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	limit := 20
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 64)
		if err != nil {
			l.Debugf("error parsing limit string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse limit query param"})
			return
		}
		limit = int(i)
	}

	maxID := ""
	maxIDString := c.Query(MaxIDKey)
	if maxIDString != "" {
		maxID = maxIDString
	}

	sinceID := ""
	sinceIDString := c.Query(SinceIDKey)
	if sinceIDString != "" {
		sinceID = sinceIDString
	}

	notifs, errWithCode := m.processor.NotificationsGet(c.Request.Context(), authed, limit, maxID, sinceID)
	if errWithCode != nil {
		l.Debugf("error processing notifications get: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, notifs)
}
