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

package admin

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
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// EmojiCreatePOSTHandler swagger:operation POST /api/v1/admin/custom_emojis emojiCreate
//
// Upload and create a new instance emoji.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: shortcode
//		in: formData
//		description: >-
//			The code to use for the emoji, which will be used by instance denizens to select it.
//			This must be unique on the instance.
//		type: string
//		pattern: \w{1,30}
//		required: true
//	-
//		name: image
//		in: formData
//		description: >-
//			A png or gif image of the emoji. Animated pngs work too!
//			To ensure compatibility with other fedi implementations, emoji size limit is 50kb by default.
//		type: file
//		required: true
//	-
//		name: category
//		in: formData
//		description: >-
//			Category in which to place the new emoji.
//			If left blank, emoji will be uncategorized. If a category with the
//			given name doesn't exist yet, it will be created.
//		type: string
//		maximumLength: 64
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: The newly-created emoji.
//			schema:
//				"$ref": "#/definitions/emoji"
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
//		'409':
//			description: conflict -- shortcode for this emoji is already in use
//		'500':
//			description: internal server error
func (m *Module) EmojiCreatePOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
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

	form := &apimodel.EmojiCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateCreateEmoji(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiEmoji, errWithCode := m.processor.Admin().EmojiCreate(c.Request.Context(), authed.Account, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiEmoji)
}

func validateCreateEmoji(form *apimodel.EmojiCreateRequest) error {
	if form.Image == nil || form.Image.Size == 0 {
		return errors.New("no emoji given")
	}

	maxSize := int64(config.GetMediaEmojiLocalMaxSize()) // #nosec G115 -- Already validated.
	if form.Image.Size > maxSize {
		return fmt.Errorf("emoji image too large: image is %dKB but size limit for custom emojis is %dKB", form.Image.Size/1024, maxSize/1024)
	}

	if err := validate.EmojiShortcode(form.Shortcode); err != nil {
		return err
	}

	return validate.EmojiCategory(form.CategoryName)
}
