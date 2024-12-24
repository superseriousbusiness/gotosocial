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

package statuses

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// StatusEditPUTHandler swagger:operation PUT /api/v1/statuses statusEdit
//
// Edit an existing status using the given form field parameters.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
//
//	---
//	tags:
//	- statuses
//
//	consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	parameters:
//	-
//		name: status
//		x-go-name: Status
//		description: |-
//			Text content of the status.
//			If media_ids is provided, this becomes optional.
//			Attaching a poll is optional while status is provided.
//		type: string
//		in: formData
//	-
//		name: media_ids
//		x-go-name: MediaIDs
//		description: |-
//			Array of Attachment ids to be attached as media.
//			If provided, status becomes optional, and poll cannot be used.
//
//			If the status is being submitted as a form, the key is 'media_ids[]',
//			but if it's json or xml, the key is 'media_ids'.
//		type: array
//		items:
//			type: string
//		in: formData
//	-
//		name: poll[options][]
//		x-go-name: PollOptions
//		description: |-
//			Array of possible poll answers.
//			If provided, media_ids cannot be used, and poll[expires_in] must be provided.
//		type: array
//		items:
//			type: string
//		in: formData
//	-
//		name: poll[expires_in]
//		x-go-name: PollExpiresIn
//		description: |-
//			Duration the poll should be open, in seconds.
//			If provided, media_ids cannot be used, and poll[options] must be provided.
//		type: integer
//		format: int64
//		in: formData
//	-
//		name: poll[multiple]
//		x-go-name: PollMultiple
//		description: Allow multiple choices on this poll.
//		type: boolean
//		default: false
//		in: formData
//	-
//		name: poll[hide_totals]
//		x-go-name: PollHideTotals
//		description: Hide vote counts until the poll ends.
//		type: boolean
//		default: true
//		in: formData
//	-
//		name: sensitive
//		x-go-name: Sensitive
//		description: Status and attached media should be marked as sensitive.
//		type: boolean
//		in: formData
//	-
//		name: spoiler_text
//		x-go-name: SpoilerText
//		description: |-
//			Text to be shown as a warning or subject before the actual content.
//			Statuses are generally collapsed behind this field.
//		type: string
//		in: formData
//	-
//		name: language
//		x-go-name: Language
//		description: ISO 639 language code for this status.
//		type: string
//		in: formData
//	-
//		name: content_type
//		x-go-name: ContentType
//		description: Content type to use when parsing this status.
//		type: string
//		enum:
//			- text/plain
//			- text/markdown
//		in: formData
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- write:statuses
//
//	responses:
//		'200':
//			description: "The latest status revision."
//			schema:
//				"$ref": "#/definitions/status"
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
func (m *Module) StatusEditPUTHandler(c *gin.Context) {
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

	form, errWithCode := parseStatusEditForm(c)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiStatus, errWithCode := m.processor.Status().Edit(
		c.Request.Context(),
		authed.Account,
		c.Param(IDKey),
		form,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiStatus)
}

func parseStatusEditForm(c *gin.Context) (*apimodel.StatusEditRequest, gtserror.WithCode) {
	form := new(apimodel.StatusEditRequest)

	switch ct := c.ContentType(); ct {
	case binding.MIMEJSON:
		// Just bind with default json binding.
		if err := c.ShouldBindWith(form, binding.JSON); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

	case binding.MIMEPOSTForm:
		// Bind with default form binding first.
		if err := c.ShouldBindWith(form, binding.FormPost); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

	case binding.MIMEMultipartPOSTForm:
		// Bind with default form binding first.
		if err := c.ShouldBindWith(form, binding.FormMultipart); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

	default:
		text := fmt.Sprintf("content-type %s not supported for this endpoint; supported content-types are %s, %s, %s",
			ct, binding.MIMEJSON, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm)
		return nil, gtserror.NewErrorNotAcceptable(errors.New(text), text)
	}

	// Normalize poll expiry time if a poll was given.
	if form.Poll != nil && form.Poll.ExpiresInI != nil {

		// If we parsed this as JSON, expires_in
		// may be either a float64 or a string.
		expiresIn, err := apiutil.ParseDuration(
			form.Poll.ExpiresInI,
			"expires_in",
		)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		form.Poll.ExpiresIn = util.PtrOrZero(expiresIn)
	}

	return form, nil

}
