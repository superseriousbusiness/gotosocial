/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
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
	if apiVersion := c.Param(APIVersionKey); apiVersion != APIv1 {
		err := errors.New("api version must be one v1 for this path")
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err, err.Error()), m.processor.InstanceGet)
		return
	}

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	attachmentID := c.Param(IDKey)
	if attachmentID == "" {
		err := errors.New("no attachment id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	form := &apimodel.AttachmentUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if err := validateUpdateMedia(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	attachment, errWithCode := m.processor.MediaUpdate(c.Request.Context(), authed, attachmentID, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, attachment)
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
