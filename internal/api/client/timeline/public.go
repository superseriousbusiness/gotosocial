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

package timeline

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// PublicTimelineGETHandler handles PUBLIC timeline requests.
// This includes requests to local, which are actually just public
// requests with a filter.
func (m *Module) PublicTimelineGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "PublicTimelineGETHandler")

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("error authing: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
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

	minID := ""
	minIDString := c.Query(MinIDKey)
	if minIDString != "" {
		minID = minIDString
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

	local := false
	localString := c.Query(LocalKey)
	if localString != "" {
		i, err := strconv.ParseBool(localString)
		if err != nil {
			l.Debugf("error parsing local string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse local query param"})
			return
		}
		local = i
	}

	resp, errWithCode := m.processor.PublicTimelineGet(authed, maxID, sinceID, minID, limit, local)
	if errWithCode != nil {
		l.Debugf("error from processor PublicTimelineGet: %s", errWithCode)
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Statuses)
}
