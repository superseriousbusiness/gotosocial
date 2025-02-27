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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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
//		name: keywords_attributes[][keyword]
//		in: formData
//		type: array
//		items:
//			type: string
//		description: Keywords to be added to the created filter.
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
//		description: Statuses to be added to the newly created filter.
//		collectionFormat: multi
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
//			Number of seconds from now that the filter should expire.
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

	// Parse form variant of normal filter keyword update structs.
	// All filter keyword update struct fields are optional.
	numFormKeywords := max(
		len(form.KeywordsAttributesID),
		len(form.KeywordsAttributesKeyword),
		len(form.KeywordsAttributesWholeWord),
		len(form.KeywordsAttributesDestroy),
	)
	if numFormKeywords > 0 {
		form.Keywords = make([]apimodel.FilterKeywordCreateUpdateDeleteRequest, 0, numFormKeywords)
		for i := 0; i < numFormKeywords; i++ {
			formKeyword := apimodel.FilterKeywordCreateUpdateDeleteRequest{}
			if i < len(form.KeywordsAttributesID) && form.KeywordsAttributesID[i] != "" {
				formKeyword.ID = &form.KeywordsAttributesID[i]
			}
			if i < len(form.KeywordsAttributesKeyword) && form.KeywordsAttributesKeyword[i] != "" {
				formKeyword.Keyword = &form.KeywordsAttributesKeyword[i]
			}
			if i < len(form.KeywordsAttributesWholeWord) {
				formKeyword.WholeWord = &form.KeywordsAttributesWholeWord[i]
			}
			if i < len(form.KeywordsAttributesDestroy) {
				formKeyword.Destroy = &form.KeywordsAttributesDestroy[i]
			}
			form.Keywords = append(form.Keywords, formKeyword)
		}
	}

	// Parse form variant of normal filter status update structs.
	// All filter status update struct fields are optional.
	numFormStatuses := max(
		len(form.StatusesAttributesID),
		len(form.StatusesAttributesStatusID),
		len(form.StatusesAttributesDestroy),
	)
	if numFormStatuses > 0 {
		form.Statuses = make([]apimodel.FilterStatusCreateDeleteRequest, 0, numFormStatuses)
		for i := 0; i < numFormStatuses; i++ {
			formStatus := apimodel.FilterStatusCreateDeleteRequest{}
			if i < len(form.StatusesAttributesID) && form.StatusesAttributesID[i] != "" {
				formStatus.ID = &form.StatusesAttributesID[i]
			}
			if i < len(form.StatusesAttributesStatusID) && form.StatusesAttributesStatusID[i] != "" {
				formStatus.StatusID = &form.StatusesAttributesStatusID[i]
			}
			if i < len(form.StatusesAttributesDestroy) {
				formStatus.Destroy = &form.StatusesAttributesDestroy[i]
			}
			form.Statuses = append(form.Statuses, formStatus)
		}
	}

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

	// Normalize and validate updates.
	for i, formKeyword := range form.Keywords {
		if formKeyword.Keyword != nil {
			if err := validate.FilterKeyword(*formKeyword.Keyword); err != nil {
				return err
			}
		}

		destroy := util.PtrOrValue(formKeyword.Destroy, false)
		form.Keywords[i].Destroy = &destroy

		if destroy && formKeyword.ID == nil {
			return errors.New("can't delete a filter keyword without an ID")
		} else if formKeyword.ID == nil && formKeyword.Keyword == nil {
			return errors.New("can't create a filter keyword without a keyword")
		}
	}
	for i, formStatus := range form.Statuses {
		if formStatus.StatusID != nil {
			if err := validate.ULID(*formStatus.StatusID, "status_id"); err != nil {
				return err
			}
		}

		destroy := util.PtrOrValue(formStatus.Destroy, false)
		form.Statuses[i].Destroy = &destroy

		switch {
		case destroy && formStatus.ID == nil:
			return errors.New("can't delete a filter status without an ID")
		case formStatus.ID != nil:
			return errors.New("filter status IDs here can only be used to delete them")
		case formStatus.StatusID == nil:
			return errors.New("can't create a filter status without a status ID")
		}
	}

	return nil
}
