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

package favourites

import (
	"fmt"
	"net/http"
	"strconv"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// FavouritesGETHandler swagger:operation GET /api/v1/favourites favouritesGet
//
// Get an array of statuses that the requesting account has favourited.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/favourites?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/favourites?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
//	---
//	tags:
//	- favourites
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: limit
//		type: integer
//		description: Number of statuses to return.
//		default: 20
//		in: query
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only favourited statuses *OLDER* than the given favourite ID.
//			The status with the corresponding fave ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only favourited statuses *NEWER* than the given favourite ID.
//			The status with the corresponding fave ID will not be included in the response.
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- read:favourites
//
//	responses:
//		'200':
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/status"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) FavouritesGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadFavourites,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
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

	limit := 20
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 32)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}
		limit = int(i)
	}

	resp, errWithCode := m.processor.Timeline().FavedTimelineGet(c.Request.Context(), authed, maxID, minID, limit)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}
