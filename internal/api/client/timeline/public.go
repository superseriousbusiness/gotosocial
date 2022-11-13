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

package timeline

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// PublicTimelineGETHandler swagger:operation GET /api/v1/timelines/public publicTimeline
//
// See public statuses/posts that your instance is aware of.
//
// The statuses will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// The returned Link header can be used to generate the previous and next queries when scrolling up or down a timeline.
//
// Example:
//
// ```
// <https://example.org/api/v1/timelines/public?limit=20&max_id=01FC3GSQ8A3MMJ43BPZSGEG29M>; rel="next", <https://example.org/api/v1/timelines/public?limit=20&min_id=01FC3KJW2GYXSDDRA6RWNDM46M>; rel="prev"
// ````
//
//	---
//	tags:
//	- timelines
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only statuses *OLDER* than the given max status ID.
//			The status with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only statuses *NEWER* than the given since status ID.
//			The status with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only statuses *NEWER* than the given since status ID.
//			The status with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of statuses to return.
//		default: 20
//		in: query
//		required: false
//	-
//		name: local
//		type: boolean
//		description: Show only statuses posted by local accounts.
//		default: false
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- read:statuses
//
//	responses:
//		'200':
//			name: statuses
//			description: Array of statuses.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/status"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//		'401':
//			description: unauthorized
//		'400':
//			description: bad request
func (m *Module) PublicTimelineGETHandler(c *gin.Context) {
	var authed *oauth.Auth
	var err error

	if config.GetInstanceExposePublicTimeline() {
		// If the public timeline is allowed to be exposed, still check if we
		// can extract various authentication properties, but don't require them.
		authed, err = oauth.Authed(c, false, false, false, false)
	} else {
		authed, err = oauth.Authed(c, true, true, true, true)
	}

	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
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
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		limit = int(i)
	}

	local := false
	localString := c.Query(LocalKey)
	if localString != "" {
		i, err := strconv.ParseBool(localString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LocalKey, err)
			api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		local = i
	}

	resp, errWithCode := m.processor.PublicTimelineGet(c.Request.Context(), authed, maxID, sinceID, minID, limit, local)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Items)
}
