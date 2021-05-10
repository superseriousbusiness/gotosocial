package media

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

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

// MediaPUTHandler allows the owner of an attachment to update information about that attachment before it's used in a status.
func (m *Module) MediaPUTHandler(c *gin.Context) {
	l := m.log.WithField("func", "MediaGETHandler")
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

	attachment, errWithCode := m.processor.MediaUpdate(authed, attachmentID, &form)
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
