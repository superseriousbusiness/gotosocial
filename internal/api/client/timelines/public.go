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

package timelines

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gin-gonic/gin"
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
	var (
		authed      *apiutil.Auth
		errWithCode gtserror.WithCode
	)
	if config.GetInstanceExposePublicTimeline() {
		// If the public timeline is allowed to be exposed, still check if we
		// can extract various authentication properties, but don't require them.
		authed, errWithCode = apiutil.TokenAuth(c,
			false, false, false, false,
		)
	} else {
		authed, errWithCode = apiutil.TokenAuth(c,
			true, true, true, true,
			apiutil.ScopeReadStatuses,
		)
	}

	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account != nil && authed.Account.IsMoving() {
		// For moving/moved accounts, just return
		// empty to avoid breaking client apps.
		apiutil.Data(c, http.StatusOK, apiutil.AppJSON, apiutil.EmptyJSONArray)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		40, // max limit
		20, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	local, errWithCode := apiutil.ParseLocal(c.Query(apiutil.LocalKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Timeline().PublicTimelineGet(
		c.Request.Context(),
		authed.Account,
		page,
		local,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}
