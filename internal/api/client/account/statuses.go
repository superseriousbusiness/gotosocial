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

package account

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountStatusesGETHandler serves the statuses of the requested account, if they're visible to the requester.
//
// Several different filters might be passed into this function in the query:
//
// 	limit -- show only limit number of statuses
// 	exclude_replies -- exclude statuses that are a reply to another status
// 	max_id -- the maximum ID of the status to show
// 	pinned -- show only pinned statuses
// 	media_only -- show only statuses that have media attachments
func (m *Module) AccountStatusesGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "AccountStatusesGETHandler")

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		l.Debugf("error authing: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		l.Debug("no account id specified in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
		return
	}

	limit := 30
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

	excludeReplies := false
	excludeRepliesString := c.Query(ExcludeRepliesKey)
	if excludeRepliesString != "" {
		i, err := strconv.ParseBool(excludeRepliesString)
		if err != nil {
			l.Debugf("error parsing replies string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse exclude replies query param"})
			return
		}
		excludeReplies = i
	}

	maxID := ""
	maxIDString := c.Query(MaxIDKey)
	if maxIDString != "" {
		maxID = maxIDString
	}

	pinnedOnly := false
	pinnedString := c.Query(PinnedKey)
	if pinnedString != "" {
		i, err := strconv.ParseBool(pinnedString)
		if err != nil {
			l.Debugf("error parsing pinned string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse pinned query param"})
			return
		}
		pinnedOnly = i
	}

	mediaOnly := false
	mediaOnlyString := c.Query(MediaOnlyKey)
	if mediaOnlyString != "" {
		i, err := strconv.ParseBool(mediaOnlyString)
		if err != nil {
			l.Debugf("error parsing media only string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse media only query param"})
			return
		}
		mediaOnly = i
	}

	statuses, errWithCode := m.processor.AccountStatusesGet(authed, targetAcctID, limit, excludeReplies, maxID, pinnedOnly, mediaOnly)
	if errWithCode != nil {
		l.Debugf("error from processor account statuses get: %s", errWithCode)
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}
