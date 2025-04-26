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

package tags

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"github.com/gin-gonic/gin"
)

// UnfollowTagPOSTHandler swagger:operation POST /api/v1/tags/{tag_name}/unfollow unfollowTag
//
// Unfollow a hashtag.
//
// Idempotent: if you are not following the tag, this call will still succeed.
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
//		- write:follows
//
//	parameters:
//	-
//		name: tag_name
//		type: string
//		description: Name of the tag (no leading `#`)
//		in: path
//		required: true
//
//	responses:
//		'200':
//			description: "Info about the tag."
//			schema:
//				"$ref": "#/definitions/tag"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: unauthorized
//		'500':
//			description: internal server error
func (m *Module) UnfollowTagPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteFollows,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	name, errWithCode := apiutil.ParseTagName(c.Param(apiutil.TagNameKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiTag, errWithCode := m.processor.Tags().Unfollow(c.Request.Context(), authed.Account, name)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiTag)
}
