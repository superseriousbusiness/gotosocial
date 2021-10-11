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

package search

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// SearchGETHandler swagger:operation GET /api/v1/search searchGet
//
// Search for statuses, accounts, or hashtags, on this instance or elsewhere.
//
// If statuses are in the result, they will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// ---
// tags:
// - search
//
// security:
// - OAuth2 Bearer:
//   - read:search
//
// responses:
//   '200':
//     name: search results
//     description: Results of the search.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/searchResult"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
func (m *Module) SearchGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "SearchGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.Authed(c, true, true, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Errorf("error authing search request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "not authed"})
		return
	}

	accountID := c.Query(AccountIDKey)
	maxID := c.Query(MaxIDKey)
	minID := c.Query(MinIDKey)
	searchType := c.Query(TypeKey)

	excludeUnreviewed := false
	excludeUnreviewedString := c.Query(ExcludeUnreviewedKey)
	if excludeUnreviewedString != "" {
		var err error
		excludeUnreviewed, err = strconv.ParseBool(excludeUnreviewedString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("couldn't parse param %s: %s", excludeUnreviewedString, err)})
			return
		}
	}

	query := c.Query(QueryKey)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q was empty"})
		return
	}

	resolve := false
	resolveString := c.Query(ResolveKey)
	if resolveString != "" {
		var err error
		resolve, err = strconv.ParseBool(resolveString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("couldn't parse param %s: %s", resolveString, err)})
			return
		}
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
	if limit > 40 {
		limit = 40
	}
	if limit < 1 {
		limit = 1
	}

	offset := 0
	offsetString := c.Query(OffsetKey)
	if offsetString != "" {
		i, err := strconv.ParseInt(offsetString, 10, 64)
		if err != nil {
			l.Debugf("error parsing offset string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse offset query param"})
			return
		}
		offset = int(i)
	}
	if limit > 40 {
		limit = 40
	}
	if limit < 1 {
		limit = 1
	}

	following := false
	followingString := c.Query(FollowingKey)
	if followingString != "" {
		var err error
		following, err = strconv.ParseBool(followingString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("couldn't parse param %s: %s", followingString, err)})
			return
		}
	}

	searchQuery := &model.SearchQuery{
		AccountID:         accountID,
		MaxID:             maxID,
		MinID:             minID,
		Type:              searchType,
		ExcludeUnreviewed: excludeUnreviewed,
		Query:             query,
		Resolve:           resolve,
		Limit:             limit,
		Offset:            offset,
		Following:         following,
	}

	results, errWithCode := m.processor.SearchGet(c.Request.Context(), authed, searchQuery)
	if errWithCode != nil {
		l.Debugf("error searching: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, results)
}
