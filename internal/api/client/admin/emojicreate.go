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

package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// EmojiCreatePOSTHandler swagger:operation POST /api/v1/admin/custom_emojis emojiCreate
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
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '404':
//      description: not found
//   '406':
//      description: not acceptable
//   '409':
//      description: conflict -- domain/shortcode combo for emoji already exists
//   '500':
//      description: internal server error
func (m *Module) EmojiCreatePOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if !authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		api.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	form := &model.EmojiCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if err := validateCreateEmoji(form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	apiEmoji, errWithCode := m.processor.AdminEmojiCreate(c.Request.Context(), authed, form)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, apiEmoji)
}

func validateCreateEmoji(form *model.EmojiCreateRequest) error {
	if form.Image == nil || form.Image.Size == 0 {
		return errors.New("no emoji given")
	}

	return validate.EmojiShortcode(form.Shortcode)
}
