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

package users

import (
	"errors"
	"net/http"
	"strings"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// FeaturedCollectionGETHandler swagger:operation GET /users/{username}/collections/featured s2sFeaturedCollectionGet
//
// Get the featured collection (pinned posts) for a user.
//
// The response will contain an ordered collection of Note URIs in the `items` property.
//
// It is up to the caller to dereference the provided Note URIs (or not, if they already have them cached).
//
// HTTP signature is required on the request.
//
//	---
//	tags:
//	- s2s/federation
//
//	produces:
//	- application/activity+json
//
//	parameters:
//	-
//		name: username
//		type: string
//		description: Account name of the user
//		in: path
//		required: true
//
//	responses:
//		'200':
//			in: body
//			schema:
//				"$ref": "#/definitions/swaggerFeaturedCollection"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
func (m *Module) FeaturedCollectionGETHandler(c *gin.Context) {
	// usernames on our instance are always lowercase
	requestedUsername := strings.ToLower(c.Param(UsernameKey))
	if requestedUsername == "" {
		err := errors.New("no username specified in request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	contentType, err := apiutil.NegotiateAccept(c, apiutil.ActivityPubOrHTMLHeaders...)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if contentType == string(apiutil.TextHTML) {
		// This isn't an ActivityPub request;
		// redirect to the user's profile.
		c.Redirect(http.StatusSeeOther, "/@"+requestedUsername)
		return
	}

	resp, errWithCode := m.processor.Fedi().FeaturedCollectionGet(c.Request.Context(), requestedUsername)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSONType(c, http.StatusOK, contentType, resp)
}
