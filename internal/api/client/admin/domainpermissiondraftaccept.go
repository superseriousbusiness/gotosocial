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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainPermissionDraftAcceptPOSTHandler swagger:operation POST /api/v1/admin/domain_permission_drafts/{id}/accept domainPermissionDraftAccept
//
// Accept a domain permission draft, turning it into an enforced domain permission.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
//	- application/json
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		required: true
//		in: path
//		description: ID of the domain permission draft.
//		type: string
//	-
//		name: overwrite
//		in: formData
//		description: >-
//			If a domain permission already exists with the same
//			domain and permission type as the draft, overwrite
//			the existing permission with fields from the draft.
//		type: boolean
//		default: false
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: The newly created domain permission.
//			schema:
//				"$ref": "#/definitions/domainPermission"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict
//		'500':
//			description: internal server error
func (m *Module) DomainPermissionDraftAcceptPOSTHandler(c *gin.Context) {
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

	id, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	type AcceptForm struct {
		Overwrite bool `json:"overwrite" form:"overwrite"`
	}

	form := new(AcceptForm)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	domainPerm, _, errWithCode := m.processor.Admin().DomainPermissionDraftAccept(
		c.Request.Context(),
		authed.Account,
		id,
		form.Overwrite,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, domainPerm)
}
