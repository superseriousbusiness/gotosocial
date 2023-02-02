/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package accounts

import (
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// AccountCreatePOSTHandler swagger:operation POST /api/v1/accounts accountCreate
//
// Create a new account using an application token.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- accounts
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Application:
//		- write:accounts
//
//	responses:
//		'200':
//			description: "An OAuth2 access token for the newly-created account."
//			schema:
//				"$ref": "#/definitions/oauthToken"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) AccountCreatePOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, false, false)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.AccountCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateCreateAccount(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	clientIP := c.ClientIP()
	signUpIP := net.ParseIP(clientIP)
	if signUpIP == nil {
		err := errors.New("ip address could not be parsed from request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}
	form.IP = signUpIP

	ti, errWithCode := m.processor.AccountCreate(c.Request.Context(), authed, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, ti)
}

// validateCreateAccount checks through all the necessary prerequisites for creating a new account,
// according to the provided account create request. If the account isn't eligible, an error will be returned.
func validateCreateAccount(form *apimodel.AccountCreateRequest) error {
	if form == nil {
		return errors.New("form was nil")
	}

	if !config.GetAccountsRegistrationOpen() {
		return errors.New("registration is not open for this server")
	}

	if err := validate.Username(form.Username); err != nil {
		return err
	}

	if err := validate.Email(form.Email); err != nil {
		return err
	}

	if err := validate.NewPassword(form.Password); err != nil {
		return err
	}

	if !form.Agreement {
		return errors.New("agreement to terms and conditions not given")
	}

	if err := validate.Language(form.Locale); err != nil {
		return err
	}

	if err := validate.SignUpReason(form.Reason, config.GetAccountsReasonRequired()); err != nil {
		return err
	}

	return nil
}
