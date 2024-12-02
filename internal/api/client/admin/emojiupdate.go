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
	"strings"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// EmojiPATCHHandler swagger:operation PATCH /api/v1/admin/custom_emojis/{id} emojiUpdate
//
// Perform admin action on a local or remote emoji known to this instance.
//
// Action performed depends upon the action `type` provided.
//
// `disable`: disable a REMOTE emoji from being used/displayed on this instance. Does not work for local emojis.
//
// `copy`: copy a REMOTE emoji to this instance. When doing this action, a shortcode MUST be provided, and it must
// be unique among emojis already present on this instance. A category MAY be provided, and the copied emoji will then
// be put into the provided category.
//
// `modify`: modify a LOCAL emoji. You can provide a new image for the emoji and/or update the category.
//
// Local emojis cannot be deleted using this endpoint. To delete a local emoji, check DELETE /api/v1/admin/custom_emojis/{id} instead.
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
//		name: id
//		type: string
//		description: The id of the emoji.
//		in: path
//		required: true
//	-
//		name: type
//		in: formData
//		description: |-
//			Type of action to be taken. One of: (`disable`, `copy`, `modify`).
//			For REMOTE emojis, `copy` or `disable` are supported.
//			For LOCAL emojis, only `modify` is supported.
//		type: string
//		enum:
//			- copy
//			- disable
//			- modify
//		required: true
//	-
//		name: shortcode
//		in: formData
//		description: >-
//			The code to use for the emoji, which will be used by instance denizens to select it.
//			This must be unique on the instance. Works for the `copy` action type only.
//		type: string
//		pattern: \w{1,30}
//	-
//		name: image
//		in: formData
//		description: >-
//			A new png or gif image to use for the emoji. Animated pngs work too!
//			To ensure compatibility with other fedi implementations, emoji size limit is 50kb by default.
//			Works for LOCAL emojis only.
//		type: file
//	-
//		name: category
//		in: formData
//		description: >-
//			Category in which to place the emoji.
//			If a category with the given name doesn't exist yet, it will be created.
//		type: string
//		maximumLength: 64
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: The updated emoji.
//			schema:
//				"$ref": "#/definitions/adminEmoji"
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
//		'500':
//			description: internal server error
func (m *Module) EmojiPATCHHandler(c *gin.Context) {
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

	emojiID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.EmojiUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateUpdateEmoji(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	emoji, errWithCode := m.processor.Admin().EmojiUpdate(c.Request.Context(), emojiID, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, emoji)
}

// do a first pass on the form here
func validateUpdateEmoji(form *apimodel.EmojiUpdateRequest) error {
	// check + normalize update type so we don't need
	// to do this trimming + lowercasing again later
	switch strings.TrimSpace(strings.ToLower(string(form.Type))) {
	case string(apimodel.EmojiUpdateDisable):
		// no params required for this one, so don't bother checking
		form.Type = apimodel.EmojiUpdateDisable
	case string(apimodel.EmojiUpdateCopy):
		// need at least a valid shortcode when doing a copy
		if form.Shortcode == nil {
			return errors.New("emoji action type was 'copy' but no shortcode was provided")
		}

		if err := validate.EmojiShortcode(*form.Shortcode); err != nil {
			return err
		}

		// category optional during copy
		if form.CategoryName != nil {
			if err := validate.EmojiCategory(*form.CategoryName); err != nil {
				return err
			}
		}

		form.Type = apimodel.EmojiUpdateCopy
	case string(apimodel.EmojiUpdateModify):
		// need either image or category name for modify
		hasImage := form.Image != nil && form.Image.Size != 0
		hasCategoryName := form.CategoryName != nil
		if !hasImage && !hasCategoryName {
			return errors.New("emoji action type was 'modify' but no image or category name was provided")
		}

		if hasImage {
			maxSize := int64(config.GetMediaEmojiLocalMaxSize()) // #nosec G115 -- Already validated.
			if form.Image.Size > maxSize {
				return fmt.Errorf("emoji image too large: image is %dKB but size limit for custom emojis is %dKB", form.Image.Size/1024, maxSize/1024)
			}
		}

		if hasCategoryName {
			if err := validate.EmojiCategory(*form.CategoryName); err != nil {
				return err
			}
		}

		form.Type = apimodel.EmojiUpdateModify
	default:
		return errors.New("emoji action type must be one of 'disable', 'copy', 'modify'")
	}

	return nil
}
