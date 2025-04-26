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

package auth

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	oautherr "code.superseriousbusiness.org/oauth2/v4/errors"
	"github.com/gin-gonic/gin"
)

// TokenRevokePOSTHandler swagger:operation POST /oauth/revoke oauthTokenRevoke
//
// Revoke an access token to make it no longer valid for use.
//
//	---
//	tags:
//	- oauth
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: client_id
//		in: formData
//		description: The client ID, obtained during app registration.
//		type: string
//		required: true
//	-
//		name: client_secret
//		in: formData
//		description: The client secret, obtained during app registration.
//		type: string
//		required: true
//	-
//		name: token
//		in: formData
//		description: The previously obtained token, to be invalidated.
//		type: string
//		required: true
//
//	responses:
//		'200':
//			description: >-
//				OK - If you own the provided token, the API call will provide OK and an empty response `{}`.
//				This operation is idempotent, so calling this API multiple times will still return OK.
//		'400':
//			description: bad request
//		'403':
//			description: >-
//				forbidden - If you provide a token you do not own, the API call will return a 403 error.
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) TokenRevokePOSTHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Don't set `binding:"required"` on these
	// fields as we want to validate them ourself.
	form := &struct {
		ClientID     string `form:"client_id"`
		ClientSecret string `form:"client_secret"`
		Token        string `form:"token"`
	}{}
	if err := c.ShouldBind(form); err != nil {
		errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if form.Token == "" {
		errWithCode := gtserror.NewErrorBadRequest(
			oautherr.ErrInvalidRequest,
			"token not set",
		)
		apiutil.OAuthErrorHandler(c, errWithCode)
		return
	}

	if form.ClientID == "" {
		errWithCode := gtserror.NewErrorBadRequest(
			oautherr.ErrInvalidRequest,
			"client_id not set",
		)
		apiutil.OAuthErrorHandler(c, errWithCode)
		return
	}

	if form.ClientSecret == "" {
		errWithCode := gtserror.NewErrorBadRequest(
			oautherr.ErrInvalidRequest,
			"client_secret not set",
		)
		apiutil.OAuthErrorHandler(c, errWithCode)
		return
	}

	errWithCode := m.processor.OAuthRevokeAccessToken(
		c.Request.Context(),
		form.ClientID,
		form.ClientSecret,
		form.Token,
	)
	if errWithCode != nil {
		apiutil.OAuthErrorHandler(c, errWithCode)
		return
	}

	apiutil.JSON(c, http.StatusOK, struct{}{})
}
