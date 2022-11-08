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

package instance

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

// InstanceUpdatePATCHHandler swagger:operation PATCH /api/v1/instance instanceUpdate
//
// Update your instance information and/or upload a new avatar/header for the instance.
//
// This requires admin permissions on the instance.
//
//	---
//	tags:
//	- instance
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: title
//		in: formData
//		description: Title to use for the instance.
//		type: string
//		maximum: 40
//		allowEmptyValue: true
//	-
//		name: contact_username
//		in: formData
//		description: >-
//			Username of the contact account.
//			This must be the username of an instance admin.
//		type: string
//		allowEmptyValue: true
//	-
//		name: contact_email
//		in: formData
//		description: Email address to use as the instance contact.
//		type: string
//		allowEmptyValue: true
//	-
//		name: short_description
//		in: formData
//		description: Short description of the instance.
//		type: string
//		maximum: 500
//		allowEmptyValue: true
//	-
//		name: description
//		in: formData
//		description: Longer description of the instance.
//		type: string
//		maximum: 5000
//		allowEmptyValue: true
//	-
//		name: terms
//		in: formData
//		description: Terms and conditions of the instance.
//		type: string
//		maximum: 5000
//		allowEmptyValue: true
//	-
//		name: thumbnail
//		in: formData
//		description: Thumbnail image to use for the instance.
//		type: file
//	-
//		name: thumbnail_description
//		in: formData
//		description: Image description of the submitted instance thumbnail.
//		type: string
//	-
//		name: header
//		in: formData
//		description: Header image to use for the instance.
//		type: file
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: "The newly updated instance."
//			schema:
//				"$ref": "#/definitions/instance"
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
func (m *Module) InstanceUpdatePATCHHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if !*authed.User.Admin {
		err := errors.New("user is not an admin so cannot update instance settings")
		api.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGet)
		return
	}

	form := &model.InstanceSettingsUpdateRequest{}
	if err := c.ShouldBind(&form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if err := validateInstanceUpdate(form); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	i, errWithCode := m.processor.InstancePatch(c.Request.Context(), form)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, i)
}

func validateInstanceUpdate(form *model.InstanceSettingsUpdateRequest) error {
	if form.Title == nil &&
		form.ContactUsername == nil &&
		form.ContactEmail == nil &&
		form.ShortDescription == nil &&
		form.Description == nil &&
		form.Terms == nil &&
		form.Avatar == nil &&
		form.AvatarDescription == nil &&
		form.Header == nil {
		return errors.New("empty form submitted")
	}

	maxImageSize := config.GetMediaImageMaxSize()
	maxDescriptionChars := config.GetMediaDescriptionMaxChars()

	// validate avatar if present
	if form.Avatar != nil {
		if size := form.Avatar.Size; size > int64(maxImageSize) {
			return fmt.Errorf("file size limit exceeded: limit is %d bytes but desired instance avatar was %d bytes", maxImageSize, size)
		}

		if form.AvatarDescription != nil {
			if length := len([]rune(*form.AvatarDescription)); length > maxDescriptionChars {
				return fmt.Errorf("avatar description length must be less than %d characters (inclusive), but provided avatar description was %d chars", maxDescriptionChars, length)
			}
		}
	}

	return nil
}
