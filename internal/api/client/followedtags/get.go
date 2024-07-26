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

package followedtags

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// FollowedTagsGETHandler swagger:operation GET /api/v1/followed_tags getFollowedTags
//
// Get an array of all hashtags that you currently follow.
//
//	---
//	tags:
//	- tags
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read:follows
//
//	parameters:
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only followed tags *OLDER* than the given max ID.
//			The followed tag with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal followed tag, NOT a tag name.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only followed tags *NEWER* than the given since ID.
//			The followed tag with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal followed tag, NOT a tag name.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only followed tags *IMMEDIATELY NEWER* than the given min ID.
//			The followed tag with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal followed tag, NOT a tag name.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of followed tags to return.
//		default: 100
//		minimum: 1
//		maximum: 200
//		in: query
//		required: false
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
//					"$ref": "#/definitions/tag"
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
func (m *Module) FollowedTagsGETHandler(c *gin.Context) {
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
		1,   // min limit
		200, // max limit
		100, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Tags().Followed(
		c.Request.Context(),
		authed.Account.ID,
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
