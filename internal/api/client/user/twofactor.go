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

package user

import (
	"errors"
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/gin-gonic/gin"
)

const OIDCTwoFactorHelp = "two factor authentication request cannot be processed by GoToSocial as this instance is running with OIDC enabled; you must use 2FA provided by your OIDC provider"

// TwoFactorQRCodePngGETHandler swagger:operation GET /api/v1/user/2fa/qr.png TwoFactorQRCodePngGet
//
// Return a QR code png to allow the authorized user to enable 2fa for their login.
//
// For the plaintext version of the QR code URI, call /api/v1/user/2fa/qruri instead.
//
// If 2fa is already enabled for this user, the QR code (with its secret) will not be shared again. Instead, code 409 Conflict will be returned. To get a fresh secret, first disable 2fa using POST /api/v1/user/2fa/disable, and then call this endpoint again.
//
// If the instance is running with OIDC enabled, two factor authentication cannot be turned on or off in GtS, it must be enabled or disabled using the OIDC provider. All calls to 2fa api endpoints will return 422 Unprocessable Entity while OIDC is enabled.
//
//	---
//	tags:
//	- user
//
//	produces:
//	- image/png
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			description: QR code png
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict
//		'422':
//			description: unprocessable entity
//		'500':
//			description: internal error
func (m *Module) TwoFactorQRCodePngGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, "image/png"); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if config.GetOIDCEnabled() {
		err := errors.New("instance running with OIDC")
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, OIDCTwoFactorHelp), m.processor.InstanceGetV1)
		return
	}

	content, errWithCode := m.processor.User().TwoFactorQRCodePngGet(c.Request.Context(), authed.User)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	defer func() {
		// Close content when we're done, catch errors.
		if err := content.Content.Close(); err != nil {
			log.Errorf(c.Request.Context(), "error closing readcloser: %v", err)
		}
	}()

	c.DataFromReader(
		http.StatusOK,
		content.ContentLength,
		content.ContentType,
		content.Content,
		nil,
	)
}

// TwoFactorQRCodeURIGETHandler swagger:operation GET /api/v1/user/2fa/qruri TwoFactorQRCodeURIGet
//
// Return a QR code uri to allow the authorized user to enable 2fa for their login.
//
// For a png of the QR code, call /api/v1/user/2fa/qr.png instead.
//
// If 2fa is already enabled for this user, the QR code URI (with its secret) will not be shared again. Instead, code 409 Conflict will be returned. To get a fresh secret, first disable 2fa using POST /api/v1/user/2fa/disable, and then call this endpoint again.
//
// If the instance is running with OIDC enabled, two factor authentication cannot be turned on or off in GtS, it must be enabled or disabled using the OIDC provider. All calls to 2fa api endpoints will return 422 Unprocessable Entity while OIDC is enabled.
//
//	---
//	tags:
//	- user
//
//	produces:
//	- text/plain
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			description: QR code uri
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict
//		'422':
//			description: unprocessable entity
//		'500':
//			description: internal error
func (m *Module) TwoFactorQRCodeURIGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.TextPlain); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if config.GetOIDCEnabled() {
		err := errors.New("instance running with OIDC")
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, OIDCTwoFactorHelp), m.processor.InstanceGetV1)
		return
	}

	uri, errWithCode := m.processor.User().TwoFactorQRCodeURIGet(c.Request.Context(), authed.User)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.Data(
		c,
		http.StatusOK,
		apiutil.TextPlain,
		[]byte(uri.String()),
	)
}

// TwoFactorEnablePOSTHandler swagger:operation POST /api/v1/user/2fa/enable TwoFactorEnablePost
//
// Enable 2fa for the authorized user, using the provided code from an authenticator app, and return an array of one-time recovery codes to allow bypassing 2fa.
//
// If 2fa is already enabled for this user, code 409 Conflict will be returned.
//
// If the instance is running with OIDC enabled, two factor authentication cannot be turned on or off in GtS, it must be enabled or disabled using the OIDC provider. All calls to 2fa api endpoints will return 422 Unprocessable Entity while OIDC is enabled.
//
//	---
//	tags:
//	- user
//
//	consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: code
//		type: string
//		description: |-
//			2fa code from the user's authenticator app.
//			Sample: 123456
//		in: formData
//
//	security:
//	- OAuth2 Bearer:
//		- write:accounts
//
//	responses:
//		'200':
//			description: QR code
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict
//		'422':
//			description: unprocessable entity
//		'500':
//			description: internal error
func (m *Module) TwoFactorEnablePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if config.GetOIDCEnabled() {
		err := errors.New("instance running with OIDC")
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, OIDCPasswordHelp), m.processor.InstanceGetV1)
		return
	}

	form := &struct {
		Code string `json:"code" form:"code" validation:"required"`
	}{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	recoveryCodes, errWithCode := m.processor.User().TwoFactorEnable(
		c.Request.Context(),
		authed.User,
		form.Code,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, recoveryCodes)
}

// TwoFactorDisablePOSTHandler swagger:operation POST /api/v1/user/2fa/disable TwoFactorDisablePost
//
// Disable 2fa for the authorized user. User's current password must be provided for verification purposes.
//
// If 2fa is already disabled for this user, code 409 Conflict will be returned.
//
// If the instance is running with OIDC enabled, two factor authentication cannot be turned on or off in GtS, it must be enabled or disabled using the OIDC provider. All calls to 2fa api endpoints will return 422 Unprocessable Entity while OIDC is enabled.
//
//	---
//	tags:
//	- user
//
//	consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	parameters:
//	-
//		name: password
//		type: string
//		description: User's current password, for verification.
//		in: formData
//
//	security:
//	- OAuth2 Bearer:
//		- write:accounts
//
//	responses:
//		'200':
//			description: QR code
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict
//		'422':
//			description: unprocessable entity
//		'500':
//			description: internal error
func (m *Module) TwoFactorDisablePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if config.GetOIDCEnabled() {
		err := errors.New("instance running with OIDC")
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, OIDCPasswordHelp), m.processor.InstanceGetV1)
		return
	}

	form := &struct {
		Password string `json:"password" form:"password" validation:"required"`
	}{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if errWithCode := m.processor.User().TwoFactorDisable(
		c.Request.Context(),
		authed.User,
		form.Password,
	); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.Status(http.StatusOK)
}
