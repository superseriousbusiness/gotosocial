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
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"github.com/superseriousbusiness/oauth2/v4"
)

// accountCreatePOSTHandler handles create account requests, validates them,
// and puts them in the database if they're valid.
// It should be served as a POST at /api/v1/accounts
func (m *accountModule) accountCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountCreatePOSTHandler")
	authed, err := oauth.MustAuth(c, true, true, false, false)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	l.Trace("parsing request form")
	form := &mastotypes.AccountCreateRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	l.Tracef("validating form %+v", form)
	if err := validateCreateAccount(form, m.config.AccountsConfig, m.db); err != nil {
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

	ti, err := m.accountCreate(form, signUpIP, authed.Token, authed.Application)
	if err != nil {
		l.Errorf("internal server error while creating new account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ti)
}

// accountCreate does the dirty work of making an account and user in the database.
// It then returns a token to the caller, for use with the new account, as per the
// spec here: https://docs.joinmastodon.org/methods/accounts/
func (m *accountModule) accountCreate(form *mastotypes.AccountCreateRequest, signUpIP net.IP, token oauth2.TokenInfo, app *model.Application) (*mastotypes.Token, error) {
	l := m.log.WithField("func", "accountCreate")

	// don't store a reason if we don't require one
	reason := form.Reason
	if !m.config.AccountsConfig.ReasonRequired {
		reason = ""
	}

	l.Trace("creating new username and account")
	user, err := m.db.NewSignup(form.Username, reason, m.config.AccountsConfig.RequireApproval, form.Email, form.Password, signUpIP, form.Locale, app.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new signup in the database: %s", err)
	}

	l.Tracef("generating a token for user %s with account %s and application %s", user.ID, user.AccountID, app.ID)
	ti, err := m.oauthServer.GenerateUserAccessToken(token, app.ClientSecret, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new access token for user %s: %s", user.ID, err)
	}

	return &mastotypes.Token{
		AccessToken: ti.GetCode(),
		TokenType:   "Bearer",
		Scope:       ti.GetScope(),
		CreatedAt:   ti.GetCodeCreateAt().Unix(),
	}, nil
}

// validateCreateAccount checks through all the necessary prerequisites for creating a new account,
// according to the provided account create request. If the account isn't eligible, an error will be returned.
func validateCreateAccount(form *mastotypes.AccountCreateRequest, c *config.AccountsConfig, database db.DB) error {
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

	if err := database.IsEmailAvailable(form.Email); err != nil {
		return err
	}

	if err := database.IsUsernameAvailable(form.Username); err != nil {
		return err
	}

	return nil
}
