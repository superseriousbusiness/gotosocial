// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bookmarks

import (
	"fmt"
	"net/http"
	"strconv"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

const (
	// LimitKey is for setting the return amount limit for eg., requesting an account's statuses
	LimitKey = "limit"

	// MaxIDKey is for specifying the maximum ID of the bookmark to retrieve.
	MaxIDKey = "max_id"
	// MinIDKey is for specifying the minimum ID of the bookmark to retrieve.
	MinIDKey = "min_id"
)

// BookmarksGETHandler swagger:operation GET /api/v1/bookmarks bookmarksGet
//
// Get an array of statuses bookmarked in the instance
//
//	---
//	tags:
//	- bookmarks
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read:bookmarks
//
//	parameters:
//	-
//		name: limit
//		type: integer
//		description: Number of statuses to return.
//		default: 30
//		in: query
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only bookmarked statuses *OLDER* than the given bookmark ID.
//			The status with the corresponding bookmark ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only bookmarked statuses *NEWER* than the given bookmark ID.
//			The status with the corresponding bookmark ID will not be included in the response.
//		in: query
//
//	responses:
//		'200':
//			description: Array of bookmarked statuses
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
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) BookmarksGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadBookmarks,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	limit := 30
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 64)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}
		limit = int(i)
	}

	maxID := ""
	maxIDString := c.Query(MaxIDKey)
	if maxIDString != "" {
		maxID = maxIDString
	}

	minID := ""
	minIDString := c.Query(MinIDKey)
	if minIDString != "" {
		minID = minIDString
	}

	resp, errWithCode := m.processor.Account().BookmarksGet(c.Request.Context(), authed.Account, limit, maxID, minID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}
