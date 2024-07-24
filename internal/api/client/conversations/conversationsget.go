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

package conversations

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// ConversationsGETHandler swagger:operation GET /api/v1/conversations conversationsGet
//
// Get an array of (direct message) conversations that requesting account is involved in.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/conversations?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/conversations?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
//	---
//	tags:
//	- conversations
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only conversations with last statuses *OLDER* than the given max ID.
//			The conversation with the specified ID will not be included in the response.
//			NOTE: The ID is a status ID. Use the Link header for pagination.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only conversations with last statuses *NEWER* than the given since ID.
//			The conversation with the specified ID will not be included in the response.
//			NOTE: The ID is a status ID. Use the Link header for pagination.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only conversations with last statuses *IMMEDIATELY NEWER* than the given min ID.
//			The conversation with the specified ID will not be included in the response.
//			NOTE: The ID is a status ID. Use the Link header for pagination.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of conversations to return.
//		default: 40
//		minimum: 1
//		maximum: 80
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- read:statuses
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
//					"$ref": "#/definitions/conversation"
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
func (m *Module) ConversationsGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Conversations().GetAll(
		c.Request.Context(),
		authed.Account,
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
