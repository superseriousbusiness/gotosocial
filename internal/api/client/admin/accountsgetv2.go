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

// AccountsGETHandlerV2 swagger:operation GET /api/v2/admin/accounts adminAccountsGetV2
//
// View + page through known accounts according to given filters.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v2/admin/accounts?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v2/admin/accounts?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
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
//		name: origin
//		in: query
//		type: string
//		description: Filter for `local` or `remote` accounts.
//	-
//		name: status
//		in: query
//		type: string
//		description: Filter for `active`, `pending`, `disabled`, `silenced`, or `suspended` accounts.
//	-
//		name: permissions
//		in: query
//		type: string
//		description: Filter for accounts with staff permissions (users that can manage reports).
//	-
//		name: role_ids[]
//		in: query
//		type: array
//		items:
//			type: string
//		description: Filter for users with these roles.
//	-
//		name: invited_by
//		in: query
//		type: string
//		description: Lookup users invited by the account with this ID.
//	-
//		name: username
//		in: query
//		type: string
//		description: Search for the given username.
//	-
//		name: display_name
//		in: query
//		type: string
//		description: Search for the given display name.
//	-
//		name: by_domain
//		in: query
//		type: string
//		description: Filter by the given domain.
//	-
//		name: email
//		in: query
//		type: string
//		description: Lookup a user with this email.
//	-
//		name: ip
//		in: query
//		type: string
//		description: Lookup users with this IP address.
//	-
//		name: max_id
//		in: query
//		type: string
//		description: All results returned will be older than the item with this ID.
//	-
//		name: since_id
//		in: query
//		type: string
//		description: All results returned will be newer than the item with this ID.
//	-
//		name: min_id
//		in: query
//		type: string
//		description: Returns results immediately newer than the item with this ID.
//	-
//		name: limit
//		in: query
//		type: integer
//		description: Maximum number of results to return.
//		default: 100
//		maximum: 200
//		minimum: 1
//
//	security:
//	- OAuth2 Bearer:
//		- admin
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
//					"$ref": "#/definitions/adminAccountInfo"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
package admin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) AccountsGETV2Handler(c *gin.Context) {
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

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	limit, errWithCode := apiutil.ParseLimit(c.Query(apiutil.LimitKey), 100, 200, 1)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse out all optional params from the query.
	params := &apimodel.AdminGetAccountsRequest{
		Origin:      c.Query(apiutil.AdminOriginKey),
		Status:      c.Query(apiutil.AdminStatusKey),
		Permissions: c.Query(apiutil.AdminPermissionsKey),
		RoleIDs:     c.QueryArray(apiutil.AdminRoleIDsKey),
		InvitedBy:   c.Query(apiutil.AdminInvitedByKey),
		Username:    c.Query(apiutil.UsernameKey),
		DisplayName: c.Query(apiutil.AdminDisplayNameKey),
		ByDomain:    c.Query(apiutil.AdminByDomainKey),
		Email:       c.Query(apiutil.AdminEmailKey),
		IP:          c.Query(apiutil.AdminIPKey),
		MaxID:       apiutil.ParseMaxID(c.Query(apiutil.MaxIDKey), ""),
		SinceID:     apiutil.ParseSinceID(c.Query(apiutil.SinceIDKey), ""),
		MinID:       apiutil.ParseMinID(c.Query(apiutil.MinIDKey), ""),
		Limit:       limit,
		APIVersion:  2,
	}

	resp, errWithCode := m.processor.Admin().AccountsGet(c.Request.Context(), params)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}