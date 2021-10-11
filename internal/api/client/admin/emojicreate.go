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

package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// emojiCreateRequest swagger:operation POST /api/v1/admin/custom_emojis emojiCreate
//
// Upload and create a new instance emoji.
//
// ---
// tags:
// - admin
//
// consumes:
// - multipart/form-data
//
// produces:
// - application/json
//
// parameters:
// - name: shortcode
//   in: formData
//   description: |-
//     The code to use for the emoji, which will be used by instance denizens to select it.
//     This must be unique on the instance.
//   type: string
//   pattern: \w{2,30}
//   required: true
// - name: image
//   in: formData
//   description: A png or gif image of the emoji. Animated pngs work too!
//   type: file
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - admin
//
// responses:
//   '200':
//     description: The newly-created emoji.
//     schema:
//       "$ref": "#/definitions/emoji"
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
func (m *Module) emojiCreatePOSTHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "emojiCreatePOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})

	// make sure we're authed with an admin account
	authed, err := oauth.Authed(c, true, true, true, true) // posting a status is serious business so we want *everything*
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if !authed.User.Admin {
		l.Debugf("user %s not an admin", authed.User.ID)
		c.JSON(http.StatusForbidden, gin.H{"error": "not an admin"})
		return
	}

	// extract the media create form from the request context
	l.Tracef("parsing request form: %+v", c.Request.Form)
	form := &model.EmojiCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		l.Debugf("error parsing form %+v: %s", c.Request.Form, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not parse form: %s", err)})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateEmoji(form); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiEmoji, err := m.processor.AdminEmojiCreate(c.Request.Context(), authed, form)
	if err != nil {
		l.Debugf("error creating emoji: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiEmoji)
}

func validateCreateEmoji(form *model.EmojiCreateRequest) error {
	// check there actually is an image attached and it's not size 0
	if form.Image == nil || form.Image.Size == 0 {
		return errors.New("no emoji given")
	}

	// a very superficial check to see if the media size limit is exceeded
	if form.Image.Size > media.EmojiMaxBytes {
		return fmt.Errorf("file size limit exceeded: limit is %d bytes but emoji was %d bytes", media.EmojiMaxBytes, form.Image.Size)
	}

	return validate.EmojiShortcode(form.Shortcode)
}
