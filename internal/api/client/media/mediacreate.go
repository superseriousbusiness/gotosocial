/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// MediaCreatePOSTHandler swagger:operation POST /api/v1/media mediaCreate
//
// Upload a new media attachment.
//
// ---
// tags:
// - media
//
// consumes:
// - multipart/form-data
//
// produces:
// - application/json
//
// parameters:
// - name: description
//   in: formData
//   description: |-
//     Image or media description to use as alt-text on the attachment.
//     This is very useful for users of screenreaders.
//     May or may not be required, depending on your instance settings.
//   type: string
// - name: focus
//   in: formData
//   description: |-
//     Focus of the media file.
//     If present, it should be in the form of two comma-separated floats between -1 and 1.
//     For example: `-0.5,0.25`.
//   type: string
// - name: file
//   in: formData
//   description: The media attachment to upload.
//   type: file
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - write:media
//
// responses:
//   '200':
//     description: The newly-created media attachment.
//     schema:
//       "$ref": "#/definitions/attachment"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '422':
//      description: unprocessable
//   '500':
//      description: internal server error
func (m *Module) MediaCreatePOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	form := &model.AttachmentRequest{}
	if err := c.ShouldBind(&form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if err := validateCreateMedia(form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	apiAttachment, errWithCode := m.processor.MediaCreate(c.Request.Context(), authed, form)
	if err != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, apiAttachment)
}

func validateCreateMedia(form *model.AttachmentRequest) error {
	// check there actually is a file attached and it's not size 0
	if form.File == nil {
		return errors.New("no attachment given")
	}

	maxVideoSize := config.GetMediaVideoMaxSize()
	maxImageSize := config.GetMediaImageMaxSize()
	minDescriptionChars := config.GetMediaDescriptionMinChars()
	maxDescriptionChars := config.GetMediaDescriptionMaxChars()

	// a very superficial check to see if no size limits are exceeded
	// we still don't actually know which media types we're dealing with but the other handlers will go into more detail there
	maxSize := maxVideoSize
	if maxImageSize > maxSize {
		maxSize = maxImageSize
	}

	if form.File.Size > int64(maxSize) {
		return fmt.Errorf("file size limit exceeded: limit is %d bytes but attachment was %d bytes", maxSize, form.File.Size)
	}

	if len(form.Description) > maxDescriptionChars {
		return fmt.Errorf("image description length must be between %d and %d characters (inclusive), but provided image description was %d chars", minDescriptionChars, maxDescriptionChars, len(form.Description))
	}

	return nil
}
