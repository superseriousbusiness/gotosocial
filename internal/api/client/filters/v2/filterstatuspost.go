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

package v2

import (
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
	"github.com/gin-gonic/gin"
)

// FilterStatusPOSTHandler swagger:operation POST /api/v2/filters/{id}/statuses filterStatusPost
//
// Add a filter status to an existing filter.
//
//	---
//	tags:
//	- filters
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
//		in: path
//		type: string
//		required: true
//		description: ID of the filter to add the filtered status to.
//	-
//		name: status_id
//		in: formData
//		required: true
//		description: |-
//			The ID of the status to filter.
//
//			Sample: 01HXA2NE0K8T1C70K90E74GYD0
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- write:filters
//
//	responses:
//		'200':
//			name: filterStatus
//			description: New filter status.
//			schema:
//				"$ref": "#/definitions/filterStatus"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden to moved accounts
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict (duplicate status)
//		'422':
//			description: unprocessable content
//		'500':
//			description: internal server error
func (m *Module) FilterStatusPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteFilters,
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

	filterID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.FilterStatusCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateCreateFilterStatus(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiFilter, errWithCode := m.processor.FiltersV2().StatusCreate(c.Request.Context(), authed.Account, filterID, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiFilter)
}

func validateCreateFilterStatus(form *apimodel.FilterStatusCreateRequest) error {
	return validate.ULID(form.StatusID, "status_id")
}
