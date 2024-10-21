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

package instance

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
//		maxLength: 40
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
//		maxLength: 500
//		allowEmptyValue: true
//	-
//		name: description
//		in: formData
//		description: Longer description of the instance.
//		type: string
//		maxLength: 5000
//		allowEmptyValue: true
//	-
//		name: terms
//		in: formData
//		description: Terms and conditions of the instance.
//		type: string
//		maxLength: 5000
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
//				"$ref": "#/definitions/instanceV1"
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
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := errors.New("user is not an admin so cannot update instance settings")
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	form := &apimodel.InstanceSettingsUpdateRequest{}
	if err := c.ShouldBind(&form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateInstanceUpdate(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	i, errWithCode := m.processor.InstancePatch(c.Request.Context(), form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, i)
}

func validateInstanceUpdate(form *apimodel.InstanceSettingsUpdateRequest) error {
	if form.Title == nil &&
		form.ContactUsername == nil &&
		form.ContactEmail == nil &&
		form.ShortDescription == nil &&
		form.Description == nil &&
		form.CustomCSS == nil &&
		form.Terms == nil &&
		form.Avatar == nil &&
		form.AvatarDescription == nil &&
		form.Header == nil {
		return errors.New("empty form submitted")
	}

	if form.AvatarDescription != nil {
		maxDescriptionChars := config.GetMediaDescriptionMaxChars()
		if length := len([]rune(*form.AvatarDescription)); length > maxDescriptionChars {
			return fmt.Errorf("avatar description length must be less than %d characters (inclusive), but provided avatar description was %d chars", maxDescriptionChars, length)
		}
	}

	return nil
}
