/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package reports

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// ReportsGETHandler swagger:operation GET /api/v1/reports reports
//
// See reports created by the requesting account.
//
// The reports will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// The next and previous queries can be parsed from the returned Link header.
//
// Example:
//
// ```
// <https://example.org/api/v1/reports?limit=20&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/reports?limit=20&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
//	---
//	tags:
//	- reports
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: resolved
//		type: boolean
//		description: >-
//			If set to true, only resolved reports will be returned.
//			If false, only unresolved reports will be returned.
//			If unset, reports will not be filtered on their resolved status.
//		in: query
//	-
//		name: target_account_id
//		type: string
//		description: Return only reports that target the given account id.
//		in: query
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only reports *OLDER* than the given max ID.
//			The report with the specified ID will not be included in the response.
//		in: query
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only reports *NEWER* than the given since ID.
//			The report with the specified ID will not be included in the response.
//			This parameter is functionally equivalent to min_id.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only reports *NEWER* than the given min ID.
//			The report with the specified ID will not be included in the response.
//			This parameter is functionally equivalent to since_id.
//		in: query
//	-
//		name: limit
//		type: integer
//		description: >-
//			Number of reports to return.
//			If less than 1, will be clamped to 1.
//			If more than 100, will be clamped to 100.
//		default: 20
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- read:reports
//
//	responses:
//		'200':
//			name: reports
//			description: Array of reports.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/report"
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
func (m *Module) ReportsGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	var resolved *bool
	if resolvedString := c.Query(ResolvedKey); resolvedString != "" {
		i, err := strconv.ParseBool(resolvedString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", ResolvedKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}
		resolved = &i
	}

	limit := 20
	if limitString := c.Query(LimitKey); limitString != "" {
		i, err := strconv.Atoi(limitString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}

		// normalize
		if i <= 0 {
			i = 1
		} else if i >= 100 {
			i = 100
		}
		limit = i
	}

	resp, errWithCode := m.processor.ReportsGet(c.Request.Context(), authed, resolved, c.Query(TargetAccountIDKey), c.Query(MaxIDKey), c.Query(SinceIDKey), c.Query(MinIDKey), limit)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Items)
}
