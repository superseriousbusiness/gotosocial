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

package notifications

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// NotificationsGETHandler swagger:operation GET /api/v1/notifications notifications
//
// Get notifications for currently authorized user.
//
// The notifications will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/notifications?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/notifications?limit=80&since_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
//	---
//	tags:
//	- notifications
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only notifications *OLDER* than the given max notification ID.
//			The notification with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only notifications *newer* than the given since notification ID.
//			The notification with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only notifications *immediately newer* than the given since notification ID.
//			The notification with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of notifications to return.
//		default: 20
//		in: query
//		required: false
//	-
//		name: types[]
//		type: array
//		items:
//			type: string
//			enum:
//				- follow
//				- follow_request
//				- mention
//				- reblog
//				- favourite
//				- poll
//				- status
//				- admin.sign_up
//		description: Types of notifications to include. If not provided, all notification types will be included.
//		in: query
//		required: false
//	-
//		name: exclude_types[]
//		type: array
//		items:
//			type: string
//			enum:
//				- follow
//				- follow_request
//				- mention
//				- reblog
//				- favourite
//				- poll
//				- status
//				- admin.sign_up
//		description: Types of notifications to exclude.
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- read:notifications
//
//	responses:
//		'200':
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//			name: notifications
//			description: Array of notifications.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/notification"
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
func (m *Module) NotificationsGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadNotifications,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		20, // no limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	ctx := c.Request.Context()
	resp, errWithCode := m.processor.Timeline().NotificationsGet(
		ctx,
		authed,
		page,
		parseNotificationTypes(ctx, c.QueryArray(TypesKey)),        // Include types.
		parseNotificationTypes(ctx, c.QueryArray(ExcludeTypesKey)), // Exclude types.
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

// parseNotificationTypes converts the given slice of string values
// to gtsmodel notification types, logging + skipping unknown types.
func parseNotificationTypes(
	ctx context.Context,
	values []string,
) []gtsmodel.NotificationType {
	if len(values) == 0 {
		return nil
	}

	ntypes := make([]gtsmodel.NotificationType, 0, len(values))
	for _, value := range values {
		ntype := gtsmodel.ParseNotificationType(value)
		if ntype == gtsmodel.NotificationUnknown {
			// Type we don't know about (yet), log and ignore it.
			log.Warnf(ctx, "ignoring unknown type %s", value)
			continue
		}

		ntypes = append(ntypes, ntype)
	}

	return ntypes
}
