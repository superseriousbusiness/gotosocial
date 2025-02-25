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

package interactionrequests

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// InteractionRequestsGETHandler swagger:operation GET /api/v1/interaction_requests getInteractionRequests
//
// Get an array of interactions requested on your statuses by other accounts, and pending your approval.
//
// ```
// <https://example.org/api/v1/interaction_requests?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/interaction_requests?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
//	---
//	tags:
//	- interaction_requests
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: status_id
//		type: string
//		description: >-
//			If set, then only interactions targeting the given status_id will be included in the results.
//		in: query
//		required: false
//	-
//		name: favourites
//		type: boolean
//		description: >-
//			If true or not set, pending favourites will be included in the results.
//			At least one of favourites, replies, and reblogs must be true.
//		in: query
//		required: false
//		default: true
//	-
//		name: replies
//		type: boolean
//		description: >-
//			If true or not set, pending replies will be included in the results.
//			At least one of favourites, replies, and reblogs must be true.
//		in: query
//		required: false
//		default: true
//	-
//		name: reblogs
//		type: boolean
//		description: >-
//			If true or not set, pending reblogs will be included in the results.
//			At least one of favourites, replies, and reblogs must be true.
//		in: query
//		required: false
//		default: true
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only interaction requests *OLDER* than the given max ID.
//			The interaction with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only interaction requests *NEWER* than the given since ID.
//			The interaction with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only interaction requests *IMMEDIATELY NEWER* than the given min ID.
//			The interaction with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of interaction requests to return.
//		default: 40
//		minimum: 1
//		maximum: 80
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
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/interactionRequest"
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
func (m *Module) InteractionRequestsGETHandler(c *gin.Context) {
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

	includeLikes, errWithCode := apiutil.ParseInteractionFavourites(
		c.Query(apiutil.InteractionFavouritesKey), true,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	includeReplies, errWithCode := apiutil.ParseInteractionReplies(
		c.Query(apiutil.InteractionRepliesKey), true,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	includeBoosts, errWithCode := apiutil.ParseInteractionReblogs(
		c.Query(apiutil.InteractionReblogsKey), true,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !includeLikes && !includeReplies && !includeBoosts {
		const text = "at least one of favourites, replies, or boosts must be true"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
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

	resp, errWithCode := m.processor.InteractionRequests().GetPage(
		c.Request.Context(),
		authed.Account,
		c.Query(apiutil.InteractionStatusIDKey),
		includeLikes,
		includeReplies,
		includeBoosts,
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
