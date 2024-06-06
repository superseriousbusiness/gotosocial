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
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// FilterPUTHandler swagger:operation PUT /api/v2/filters/{id} filterV2Put
//
// Update a single filter with the given ID.
// Note that this is actually closer to a PATCH operation:
// only provided fields will be updated, and omitted fields will remain set to previous values.
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
//		description: ID of the filter.
//	-
//		name: title
//		in: formData
//		required: true
//		description: |-
//			The name of the filter.
//
//			Sample: illuminati nonsense
//		type: string
//		minLength: 1
//		maxLength: 200
//	-
//		name: context[]
//		in: formData
//		required: true
//		description: |-
//			The contexts in which the filter should be applied.
//
//			Sample: home, public
//		enum:
//			- home
//			- notifications
//			- public
//			- thread
//			- account
//		type: array
//		items:
//			type:
//				string
//		collectionFormat: multi
//		minItems: 1
//		uniqueItems: true
//	-
//		name: expires_in
//		in: formData
//		description: |-
//			Number of seconds from now that the filter should expire.
//
//			Sample: 86400
//		type: number
//
//	security:
//	- OAuth2 Bearer:
//		- write:filters
//
//	responses:
//		'200':
//			name: filter
//			description: Updated filter.
//			schema:
//				"$ref": "#/definitions/filterV2"
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
//			description: conflict (duplicate title, keyword, or status)
//		'422':
//			description: unprocessable content
//		'500':
//			description: internal server error
func (m *Module) FilterPUTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
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

	form := &apimodel.FilterUpdateRequestV2{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateNormalizeUpdateFilter(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiFilter, errWithCode := m.processor.FiltersV2().Update(c.Request.Context(), authed.Account, id, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiFilter)
}

func validateNormalizeUpdateFilter(form *apimodel.FilterUpdateRequestV2) error {
	if form.Title != nil {
		if err := validate.FilterTitle(*form.Title); err != nil {
			return err
		}
	}
	if form.FilterAction != nil {
		if err := validate.FilterAction(*form.FilterAction); err != nil {
			return err
		}
	}
	if form.Context != nil {
		if err := validate.FilterContexts(*form.Context); err != nil {
			return err
		}
	}

	// Normalize filter expiry if necessary.
	// If we parsed this as JSON, expires_in
	// may be either a float64 or a string.
	if ei := form.ExpiresInI; ei != nil {
		switch e := ei.(type) {
		case float64:
			form.ExpiresIn = util.Ptr(int(e))

		case string:
			expiresIn, err := strconv.Atoi(e)
			if err != nil {
				return fmt.Errorf("could not parse expires_in value %s as integer: %w", e, err)
			}

			form.ExpiresIn = &expiresIn

		default:
			return fmt.Errorf("could not parse expires_in type %T as integer", ei)
		}
	}

	return nil
}
