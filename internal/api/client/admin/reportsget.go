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

package admin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// ReportsGETHandler swagger:operation GET /api/v1/admin/reports adminReports
//
// View user moderation reports.
//
// The reports will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// The next and previous queries can be parsed from the returned Link header.
//
// Example:
//
// ```
// <https://example.org/api/v1/admin/reports?limit=20&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/admin/reports?limit=20&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
//	---
//	tags:
//	- admin
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
//		name: account_id
//		type: string
//		description: Return only reports created by the given account id.
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
//			Return only reports *OLDER* than the given max ID (for paging downwards).
//			The report with the specified ID will not be included in the response.
//		in: query
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only reports *NEWER* than the given since ID.
//			The report with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only reports immediately *NEWER* than the given min ID (for paging upwards).
//			The report with the specified ID will not be included in the response.
//		in: query
//	-
//		name: limit
//		type: integer
//		description: Number of reports to return.
//		default: 20
//		minimum: 1
//		maximum: 100
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			name: reports
//			description: Array of reports.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/adminReport"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
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

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resolved, errWithCode := apiutil.ParseResolved(c.Query(apiutil.ResolvedKey), nil)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,   // min limit
		100, // max limit
		20,  // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Admin().ReportsGet(
		c.Request.Context(),
		authed.Account,
		resolved,
		c.Query(apiutil.AccountIDKey),
		c.Query(apiutil.TargetAccountIDKey),
		page,
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
