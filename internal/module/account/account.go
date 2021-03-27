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
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/module"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"github.com/superseriousbusiness/oauth2/v4"
)

const (
	basePath       = "/api/v1/accounts"
	basePathWithID = basePath + "/:id"
	verifyPath     = basePath + "/verify_credentials"
)

type accountModule struct {
	config      *config.Config
	db          db.DB
	oauthServer oauth.Server
	log         *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, oauthServer oauth.Server, log *logrus.Logger) module.ClientAPIModule {
	return &accountModule{
		config:      config,
		db:          db,
		oauthServer: oauthServer,
		log:         log,
	}
}

// Route attaches all routes from this module to the given router
func (m *accountModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, basePath, m.accountCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, verifyPath, m.accountVerifyGETHandler)
	return nil
}

// accountCreatePOSTHandler handles create account requests, validates them,
// and puts them in the database if they're valid.
// It should be served as a POST at /api/v1/accounts
func (m *accountModule) accountCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountCreatePOSTHandler")
	authed, err := oauth.MustAuthed(c, true, true, false, false)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	l.Trace("parsing request form")
	form := &mastotypes.AccountCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// accountVerifyGETHandler serves a user's account details to them IF they reached this
// handler while in possession of a valid token, according to the oauth middleware.
// It should be served as a GET at /api/v1/accounts/verify_credentials
func (m *accountModule) accountVerifyGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "AccountVerifyGETHandler")

	l.Trace("getting account details from session")
	i, ok := c.Get(oauth.SessionAuthorizedAccount)
	if !ok {
		l.Trace("no account in session, returning error to client")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The access token is invalid"})
		return
	}

	l.Trace("attempting to convert account interface into account struct...")
	acct, ok := i.(*model.Account)
	if !ok {
		l.Tracef("could not convert %+v into account struct, returning error to client", i)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The access token is invalid"})
		return
	}

	l.Tracef("retrieved account %+v, converting to mastosensitive...", acct)
	acctSensitive, err := m.db.AccountToMastoSensitive(acct)
	if err != nil {
		l.Tracef("could not convert account into mastosensitive account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("conversion successful, returning OK and mastosensitive account %+v", acctSensitive)
	c.JSON(http.StatusOK, acctSensitive)
}

/*
	HELPER FUNCTIONS
*/

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
