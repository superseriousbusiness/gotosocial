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

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
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
//   '403':
//      description: forbidden
//   '422':
//      description: unprocessable
func (m *Module) MediaCreatePOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true) // posting new media is serious business so we want *everything*
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	// extract the media create form from the request context
	l.Tracef("parsing request form: %s", c.Request.Form)
	form := &model.AttachmentRequest{}
	if err := c.ShouldBind(&form); err != nil {
		l.Debugf("error parsing form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("could not parse form: %s", err)})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateMedia(form); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	l.Debug("calling processor media create func")
	apiAttachment, err := m.processor.MediaCreate(c.Request.Context(), authed, form)
	if err != nil {
		l.Debugf("error creating attachment: %s", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiAttachment)
}

func validateCreateMedia(form *model.AttachmentRequest) error {
	// check there actually is a file attached and it's not size 0
	if form.File == nil {
		return errors.New("no attachment given")
	}

	keys := config.Keys
	maxVideoSize := viper.GetInt(keys.MediaVideoMaxSize)
	maxImageSize := viper.GetInt(keys.MediaImageMaxSize)
	minDescriptionChars := viper.GetInt(keys.MediaDescriptionMinChars)
	maxDescriptionChars := viper.GetInt(keys.MediaDescriptionMaxChars)

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
