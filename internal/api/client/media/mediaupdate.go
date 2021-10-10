/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
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
// ---
// tags:
// - media
//
// consumes:
// - application/json
// - application/xml
// - application/x-www-form-urlencoded
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   description: id of the attachment to update
//   type: string
//   in: path
//   required: true
// - name: description
//   in: formData
//   description: |-
//     Image or media description to use as alt-text on the attachment.
//     This is very useful for users of screenreaders.
//     May or may not be required, depending on your instance settings.
//   type: string
//   allowEmptyValue: true
// - name: focus
//   in: formData
//   description: |-
//     Focus of the media file.
//     If present, it should be in the form of two comma-separated floats between -1 and 1.
//     For example: `-0.5,0.25`.
//   type: string
//   allowEmptyValue: true
//
// security:
// - OAuth2 Bearer:
//   - write:media
//
// responses:
//   '200':
//     description: The newly-updated media attachment.
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
func (m *Module) MediaPUTHandler(c *gin.Context) {
	l := logrus.WithField("func", "MediaGETHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	attachmentID := c.Param(IDKey)
	if attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no attachment ID given in request"})
		return
	}

	// extract the media update form from the request context
	l.Tracef("parsing request form: %s", c.Request.Form)
	var form model.AttachmentUpdateRequest
	if err := c.ShouldBind(&form); err != nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateUpdateMedia(&form, m.config.MediaConfig); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attachment, errWithCode := m.processor.MediaUpdate(c.Request.Context(), authed, attachmentID, &form)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, attachment)
}

func validateUpdateMedia(form *model.AttachmentUpdateRequest, config *config.MediaConfig) error {

	if form.Description != nil {
		if len(*form.Description) < config.MinDescriptionChars || len(*form.Description) > config.MaxDescriptionChars {
			return fmt.Errorf("image description length must be between %d and %d characters (inclusive), but provided image description was %d chars", config.MinDescriptionChars, config.MaxDescriptionChars, len(*form.Description))
		}
	}

	if form.Focus == nil && form.Description == nil {
		return errors.New("focus and description were both nil, there's nothing to update")
	}

	return nil
}
