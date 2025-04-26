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

package lists

import (
	"errors"
	"net/http"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
	"github.com/gin-gonic/gin"
)

// ListUpdatePUTHandler swagger:operation PUT /api/v1/lists/{id} listUpdate
//
// Update an existing list.
//
//	---
//	tags:
//	- lists
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: ID of the list
//		in: path
//		required: true
//	-
//		name: title
//		type: string
//		description: |-
//			Title of this list.
//			Sample: Cool People
//		in: formData
//	-
//		name: replies_policy
//		type: string
//		description: |-
//		  RepliesPolicy for this list.
//		  followed = Show replies to any followed user
//		  list = Show replies to members of the list
//		  none = Show replies to no one
//		  Sample: list
//		enum:
//			- followed
//			- list
//			- none
//		in: formData
//	-
//		name: exclusive
//		in: formData
//		description: Hide posts from members of this list from your home timeline.
//		type: boolean
//
//	security:
//	- OAuth2 Bearer:
//		- write:lists
//
//	responses:
//		'200':
//			description: "The newly updated list."
//			schema:
//				"$ref": "#/definitions/list"
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
func (m *Module) ListUpdatePUTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteLists,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
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

	targetListID := c.Param(IDKey)
	if targetListID == "" {
		err := errors.New("no list id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.ListUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if form.Title != nil {
		if err := validate.ListTitle(*form.Title); err != nil {
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}
	}

	var repliesPolicy *gtsmodel.RepliesPolicy
	if form.RepliesPolicy != nil {
		rp := gtsmodel.RepliesPolicy(strings.ToLower(*form.RepliesPolicy))

		if err := validate.ListRepliesPolicy(rp); err != nil {
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}

		repliesPolicy = &rp
	}

	if form.Title == nil && repliesPolicy == nil && form.Exclusive == nil {
		err := errors.New("neither title nor replies_policy nor exclusive was set; nothing to update")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiList, errWithCode := m.processor.List().Update(
		c.Request.Context(),
		authed.Account,
		targetListID,
		form.Title,
		repliesPolicy,
		form.Exclusive,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiList)
}
