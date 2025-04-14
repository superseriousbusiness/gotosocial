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

// MediaCreatePOSTHandler swagger:operation POST /api/{api_version}/media mediaCreate
//
// Upload a new media attachment.
//
//	---
//	tags:
//	- media
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: api_version
//		type: string
//		in: path
//		description: Version of the API to use. Must be either `v1` or `v2`.
//		required: true
//	-
//		name: description
//		in: formData
//		description: >-
//			Image or media description to use as alt-text on the attachment.
//			This is very useful for users of screenreaders!
//			May or may not be required, depending on your instance settings.
//		type: string
//	-
//		name: focus
//		in: formData
//		description: >-
//			Focus of the media file.
//			If present, it should be in the form of two comma-separated floats between -1 and 1.
//			For example: `-0.5,0.25`.
//		type: string
//		default: "0,0"
//	-
//		name: file
//		in: formData
//		description: The media attachment to upload.
//		type: file
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- write:media
//
//	responses:
//		'200':
//			description: The newly-created media attachment.
//			schema:
//				"$ref": "#/definitions/attachment"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'422':
//			description: unprocessable
//		'500':
//			description: internal server error
func (m *Module) MediaCreatePOSTHandler(c *gin.Context) {
	_, errWithCode := apiutil.ParseAPIVersion(
		c.Param(apiutil.APIVersionKey),
		apiutil.APIv1,
		apiutil.APIv2,
	)
	if errWithCode != nil {
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

	form := &apimodel.AttachmentRequest{}
	if err := c.ShouldBind(&form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateCreateMedia(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiAttachment, errWithCode := m.processor.Media().Create(c.Request.Context(), authed.Account, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// The v2 mastodon endpoint always returns TextURL,
	// but in the case that a 202 Accepted (i.e. processing
	// still in progress) is returned then the URL will be
	// nil. Since we only ever return a 200 OK, we behave
	// exactly the same as the v1 endpoint.
	//
	// https://docs.joinmastodon.org/methods/media/#v2

	apiutil.JSON(c, http.StatusOK, apiAttachment)
}

func validateCreateMedia(form *apimodel.AttachmentRequest) error {
	// check there actually is a file attached and it's not size 0
	if form.File == nil {
		return errors.New("no attachment given")
	}

	minDescriptionChars := config.GetMediaDescriptionMinChars()
	maxDescriptionChars := config.GetMediaDescriptionMaxChars()

	if length := len([]rune(form.Description)); length > maxDescriptionChars {
		return fmt.Errorf("image description length must be between %d and %d characters (inclusive), but provided image description was %d chars", minDescriptionChars, maxDescriptionChars, length)
	}

	return nil
}
