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
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// DomainPermissionSubscriptionsGETHandler swagger:operation GET /api/v1/admin/domain_permission_subscriptions domainPermissionSubscriptionsGet
//
// View domain permission subscriptions.
//
// The subscriptions will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// The next and previous queries can be parsed from the returned Link header.
//
// Example:
//
// ```
// <https://example.org/api/v1/admin/domain_permission_subscriptions?limit=20&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/admin/domain_permission_subscriptions?limit=20&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
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
//		name: permission_type
//		type: string
//		description: Filter on "block" or "allow" type subscriptions.
//		in: query
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items *OLDER* than the given max ID (for paging downwards).
//			The item with the specified ID will not be included in the response.
//		in: query
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only items *NEWER* than the given since ID.
//			The item with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items immediately *NEWER* than the given min ID (for paging upwards).
//			The item with the specified ID will not be included in the response.
//		in: query
//	-
//		name: limit
//		type: integer
//		description: Number of items to return.
//		default: 20
//		minimum: 1
//		maximum: 100
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read
//
//	responses:
//		'200':
//			description: Domain permission subscriptions.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/domainPermissionSubscription"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
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
func (m *Module) DomainPermissionSubscriptionsGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminRead,
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

	permType := c.Query(apiutil.DomainPermissionPermTypeKey)
	switch permType {
	case "", "block", "allow":
		// No problem.

	default:
		// Invalid.
		text := fmt.Sprintf(
			"permission_type %s not recognized, valid values are empty string, block, or allow",
			permType,
		)
		errWithCode := gtserror.NewErrorBadRequest(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c, 1, 200, 20)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Admin().DomainPermissionSubscriptionsGet(
		c.Request.Context(),
		gtsmodel.ParseDomainPermissionType(permType),
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
