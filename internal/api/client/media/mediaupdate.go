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

package media

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// MediaPUTHandler swagger:operation PUT /api/v1/media/{id} mediaUpdate
//
// Update a media attachment.
//
// You must own the media attachment, and the attachment must not yet be attached to a status.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- media
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
//		description: id of the attachment to update
//		type: string
//		in: path
//		required: true
//	-
//		name: description
//		in: formData
//		description: >-
//			Image or media description to use as alt-text on the attachment.
//			This is very useful for users of screenreaders!
//			May or may not be required, depending on your instance settings.
//		type: string
//		allowEmptyValue: true
//	-
//		name: focus
//		in: formData
//		description: >-
//			Focus of the media file.
//			If present, it should be in the form of two comma-separated floats between -1 and 1.
//			For example: `-0.5,0.25`.
//		type: string
//		allowEmptyValue: true
//		default: "0,0"
//
//	security:
//	- OAuth2 Bearer:
//		- write:media
//
//	responses:
//		'200':
//			description: The newly-updated media attachment.
//			schema:
//				"$ref": "#/definitions/attachment"
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
func (m *Module) MediaPUTHandler(c *gin.Context) {
	if _, errWithCode := apiutil.ParseAPIVersion(
		c.Param(apiutil.APIVersionKey),
		[]string{apiutil.APIv1, apiutil.APIv2}...,
	); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteMedia,
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

	attachmentID := c.Param(IDKey)
	if attachmentID == "" {
		err := errors.New("no attachment id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.AttachmentUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateUpdateMedia(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	attachment, errWithCode := m.processor.Media().Update(c.Request.Context(), authed.Account, attachmentID, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, attachment)
}

func validateUpdateMedia(form *apimodel.AttachmentUpdateRequest) error {
	minDescriptionChars := config.GetMediaDescriptionMinChars()
	maxDescriptionChars := config.GetMediaDescriptionMaxChars()

	if form.Description != nil {
		if length := len([]rune(*form.Description)); length < minDescriptionChars || length > maxDescriptionChars {
			return fmt.Errorf("image description length must be between %d and %d characters (inclusive), but provided image description was %d chars", minDescriptionChars, maxDescriptionChars, length)
		}
	}

	if form.Focus == nil && form.Description == nil {
		return errors.New("focus and description were both nil, there's nothing to update")
	}

	return nil
}
