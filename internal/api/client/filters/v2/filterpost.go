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

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// FilterPOSTHandler swagger:operation POST /api/v2/filters filterV2Post
//
// Create a single filter.
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
//		type: array
//		items:
//			type:
//				string
//			enum:
//				- home
//				- notifications
//				- public
//				- thread
//				- account
//		collectionFormat: multi
//		minItems: 1
//		uniqueItems: true
//	-
//		name: expires_in
//		in: formData
//		description: |-
//			Number of seconds from now that the filter should expire. If omitted, filter never expires.
//
//			Sample: 86400
//		type: number
//	-
//		name: filter_action
//		in: formData
//		description: |-
//			The action to be taken when a status matches this filter.
//
//			Sample: warn
//		type: string
//		enum:
//			- warn
//			- hide
//		default: warn
//	-
//		name: keywords_attributes[][keyword]
//		in: formData
//		type: array
//		items:
//			type: string
//		description: Keywords to be added (if not using id param) or updated (if using id param).
//		collectionFormat: multi
//	-
//		name: keywords_attributes[][whole_word]
//		in: formData
//		type: array
//		items:
//			type: boolean
//		description: Should each keyword consider word boundaries?
//		collectionFormat: multi
//	-
//		name: statuses_attributes[][status_id]
//		in: formData
//		type: array
//		items:
//			type: string
//		description: Statuses to be added to the filter.
//		collectionFormat: multi
//
//	security:
//	- OAuth2 Bearer:
//		- write:filters
//
//	responses:
//		'200':
//			name: filter
//			description: New filter.
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
func (m *Module) FilterPOSTHandler(c *gin.Context) {
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

	form := &apimodel.FilterCreateRequestV2{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateNormalizeCreateFilter(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiFilter, errWithCode := m.processor.FiltersV2().Create(c.Request.Context(), authed.Account, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiFilter)
}

func validateNormalizeCreateFilter(form *apimodel.FilterCreateRequestV2) error {
	if err := validate.FilterTitle(form.Title); err != nil {
		return err
	}
	action := util.PtrOrValue(form.FilterAction, apimodel.FilterActionWarn)
	if err := validate.FilterAction(action); err != nil {
		return err
	}
	if err := validate.FilterContexts(form.Context); err != nil {
		return err
	}

	// Parse form variant of normal filter keyword creation structs.
	if len(form.KeywordsAttributesKeyword) > 0 {
		form.Keywords = make([]apimodel.FilterKeywordCreateUpdateRequest, 0, len(form.KeywordsAttributesKeyword))
		for i, keyword := range form.KeywordsAttributesKeyword {
			formKeyword := apimodel.FilterKeywordCreateUpdateRequest{
				Keyword: keyword,
			}
			if i < len(form.KeywordsAttributesWholeWord) {
				formKeyword.WholeWord = &form.KeywordsAttributesWholeWord[i]
			}
			form.Keywords = append(form.Keywords, formKeyword)
		}
	}

	// Parse form variant of normal filter status creation structs.
	if len(form.StatusesAttributesStatusID) > 0 {
		form.Statuses = make([]apimodel.FilterStatusCreateRequest, 0, len(form.StatusesAttributesStatusID))
		for _, statusID := range form.StatusesAttributesStatusID {
			form.Statuses = append(form.Statuses, apimodel.FilterStatusCreateRequest{
				StatusID: statusID,
			})
		}
	}

	// Apply defaults for missing fields.
	form.FilterAction = util.Ptr(action)

	// If `expires_in` was provided
	// as JSON, then normalize it.
	if form.ExpiresInI.IsSpecified() {
		var err error
		form.ExpiresIn, err = apiutil.ParseNullableDuration(
			form.ExpiresInI,
			"expires_in",
		)
		if err != nil {
			return err
		}
	}

	// Normalize and validate new keywords and statuses.
	for i, formKeyword := range form.Keywords {
		if err := validate.FilterKeyword(formKeyword.Keyword); err != nil {
			return err
		}
		form.Keywords[i].WholeWord = util.Ptr(util.PtrOrValue(formKeyword.WholeWord, false))
	}
	for _, formStatus := range form.Statuses {
		if err := validate.ULID(formStatus.StatusID, "status_id"); err != nil {
			return err
		}
	}

	return nil
}
