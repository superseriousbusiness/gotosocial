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

// AccountsGETHandlerV1 swagger:operation GET /api/v1/admin/accounts adminAccountsGetV1
//
// View + page through known accounts according to given filters.
//
// Returned accounts will be ordered alphabetically (a-z) by domain + username.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/admin/accounts?limit=80&max_id=example.org%2F%40someone>; rel="next", <https://example.org/api/v1/admin/accounts?limit=80&min_id=example.org%2F%40someone_else>; rel="prev"
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
//		name: local
//		in: query
//		type: boolean
//		description: Filter for local accounts.
//		default: false
//	-
//		name: remote
//		in: query
//		type: boolean
//		description: Filter for remote accounts.
//		default: false
//	-
//		name: active
//		in: query
//		type: boolean
//		description: Filter for currently active accounts.
//		default: false
//	-
//		name: pending
//		in: query
//		type: boolean
//		description: Filter for currently pending accounts.
//		default: false
//	-
//		name: disabled
//		in: query
//		type: boolean
//		description: Filter for currently disabled accounts.
//		default: false
//	-
//		name: silenced
//		in: query
//		type: boolean
//		description: Filter for currently silenced accounts.
//		default: false
//	-
//		name: suspended
//		in: query
//		type: boolean
//		description: Filter for currently suspended accounts.
//		default: false
//	-
//		name: sensitized
//		in: query
//		type: boolean
//		description: Filter for accounts force-marked as sensitive.
//		default: false
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
//		name: staff
//		in: query
//		type: boolean
//		description: Filter for staff accounts.
//		default: false
//	-
//		name: max_id
//		in: query
//		type: string
//		description: >-
//			max_id in the form `[domain]/@[username]`.
//			All results returned will be later in the alphabet than `[domain]/@[username]`.
//			For example, if max_id = `example.org/@someone` then returned entries might
//			contain `example.org/@someone_else`, `later.example.org/@someone`, etc.
//			Local account IDs in this form use an empty string for the `[domain]` part,
//			for example local account with username `someone` would be `/@someone`.
//	-
//		name: min_id
//		in: query
//		type: string
//		description: >-
//			min_id in the form `[domain]/@[username]`.
//			All results returned will be earlier in the alphabet than `[domain]/@[username]`.
//			For example, if min_id = `example.org/@someone` then returned entries might
//			contain `example.org/@earlier_account`, `earlier.example.org/@someone`, etc.
//			Local account IDs in this form use an empty string for the `[domain]` part,
//			for example local account with username `someone` would be `/@someone`.
//	-
//		name: limit
//		in: query
//		type: integer
//		description: Maximum number of results to return.
//		default: 50
//		maximum: 200
//		minimum: 1
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:accounts
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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gin-gonic/gin"
)

func (m *Module) AccountsGETV1Handler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminReadAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
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

	page, errWithCode := paging.ParseIDPage(c, 1, 200, 50)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	/* Translate to v2 `origin` query param */

	local, errWithCode := apiutil.ParseLocal(c.Query(apiutil.LocalKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	remote, errWithCode := apiutil.ParseAdminRemote(c.Query(apiutil.AdminRemoteKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if local && remote {
		keys := []string{apiutil.LocalKey, apiutil.AdminRemoteKey}
		err := fmt.Errorf("only one of %+v can be true", keys)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	var origin string
	if local {
		origin = "local"
	} else if remote {
		origin = "remote"
	}

	/* Translate to v2 `status` query param */

	active, errWithCode := apiutil.ParseAdminActive(c.Query(apiutil.AdminActiveKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	pending, errWithCode := apiutil.ParseAdminPending(c.Query(apiutil.AdminPendingKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	disabled, errWithCode := apiutil.ParseAdminDisabled(c.Query(apiutil.AdminDisabledKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	silenced, errWithCode := apiutil.ParseAdminSilenced(c.Query(apiutil.AdminSilencedKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	suspended, errWithCode := apiutil.ParseAdminSuspended(c.Query(apiutil.AdminSuspendedKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Ensure only one `status` query param set.
	var status string
	states := map[string]bool{
		apiutil.AdminActiveKey:    active,
		apiutil.AdminPendingKey:   pending,
		apiutil.AdminDisabledKey:  disabled,
		apiutil.AdminSilencedKey:  silenced,
		apiutil.AdminSuspendedKey: suspended,
	}
	for k, v := range states {
		if !v {
			// False status,
			// so irrelevant.
			continue
		}

		if status != "" {
			// Status was already set by another
			// query param, this is an error.
			keys := []string{
				apiutil.AdminActiveKey,
				apiutil.AdminPendingKey,
				apiutil.AdminDisabledKey,
				apiutil.AdminSilencedKey,
				apiutil.AdminSuspendedKey,
			}
			err := fmt.Errorf("only one of %+v can be true", keys)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}

		// Use this
		// account status.
		status = k
	}

	/* Translate to v2 `permissions` query param */

	staff, errWithCode := apiutil.ParseAdminStaff(c.Query(apiutil.AdminStaffKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	var permissions string
	if staff {
		permissions = "staff"
	}

	// Parse out all optional params from the query.
	params := &apimodel.AdminGetAccountsRequest{
		Origin:      origin,
		Status:      status,
		Permissions: permissions,
		RoleIDs:     nil, // Can't do in V1.
		InvitedBy:   "",  // Can't do in V1.
		Username:    c.Query(apiutil.UsernameKey),
		DisplayName: c.Query(apiutil.AdminDisplayNameKey),
		ByDomain:    c.Query(apiutil.AdminByDomainKey),
		Email:       c.Query(apiutil.AdminEmailKey),
		IP:          c.Query(apiutil.AdminIPKey),
		APIVersion:  1,
	}

	resp, errWithCode := m.processor.Admin().AccountsGet(
		c.Request.Context(),
		params,
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
