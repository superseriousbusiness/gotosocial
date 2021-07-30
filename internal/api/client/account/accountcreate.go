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

package account

import (
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// AccountCreatePOSTHandler handles create account requests, validates them,
// and puts them in the database if they're valid.
//
// swagger:operation POST /api/v1/accounts accountCreate
//
// Create a new account using an application token.
//
// ---
// consumes:
// - application/json
// - application/xml
// - application/x-www-form-urlencoded
// - multipart/form-data
//
// produces:
// - application/json
//
// parameters:
// - name: Account Create Request
//   in: body
//   schema:
//     "$ref": "#/definitions/accountCreateRequest"
//
// security:
// - OAuth2 Application:
//   - write:accounts
//
// responses:
//   '200':
//     description: "An OAuth2 access token for the newly-created account."
//     schema:
//       "$ref": "#/definitions/oauthToken"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
//   '500':
//      description: internal error
func (m *Module) AccountCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountCreatePOSTHandler")
	authed, err := oauth.Authed(c, true, true, false, false)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	l.Trace("parsing request form")
	form := &model.AccountCreateRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	l.Tracef("validating form %+v", form)
	if err := validateCreateAccount(form, m.config.AccountsConfig); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientIP := c.ClientIP()
	l.Tracef("attempting to parse client ip address %s", clientIP)
	signUpIP := net.ParseIP(clientIP)
	if signUpIP == nil {
		l.Debugf("error validating sign up ip address %s", clientIP)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip address could not be parsed from request"})
		return
	}

	form.IP = signUpIP

	ti, err := m.processor.AccountCreate(authed, form)
	if err != nil {
		l.Errorf("internal server error while creating new account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ti)
}

// validateCreateAccount checks through all the necessary prerequisites for creating a new account,
// according to the provided account create request. If the account isn't eligible, an error will be returned.
func validateCreateAccount(form *model.AccountCreateRequest, c *config.AccountsConfig) error {
	if !c.OpenRegistration {
		return errors.New("registration is not open for this server")
	}

	if err := util.ValidateUsername(form.Username); err != nil {
		return err
	}

	if err := util.ValidateEmail(form.Email); err != nil {
		return err
	}

	if err := util.ValidateNewPassword(form.Password); err != nil {
		return err
	}

	if !form.Agreement {
		return errors.New("agreement to terms and conditions not given")
	}

	if err := util.ValidateLanguage(form.Locale); err != nil {
		return err
	}

	if err := util.ValidateSignUpReason(form.Reason, c.ReasonRequired); err != nil {
		return err
	}

	return nil
}
