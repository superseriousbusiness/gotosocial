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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/gotosocial/internal/db/model"
	"github.com/gotosocial/gotosocial/internal/module"
	"github.com/gotosocial/gotosocial/internal/module/oauth"
	"github.com/gotosocial/gotosocial/internal/router"
	"github.com/sirupsen/logrus"
)

const (
	basePath       = "/api/v1/accounts"
	basePathWithID = basePath + "/:id"
	verifyPath     = basePath + "/verify_credentials"
)

type accountModule struct {
	config *config.Config
	db     db.DB
	log    *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, log *logrus.Logger) module.ClientAPIModule {
	return &accountModule{
		config: config,
		db:     db,
		log: log,
	}
}

// Route attaches all routes from this module to the given router
func (m *accountModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, basePath, m.AccountCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, verifyPath, m.AccountVerifyGETHandler)
	return nil
}

func (m *accountModule) AccountCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "AccountCreatePOSTHandler")
	l.Trace("checking if registration is open")
	if !m.config.AccountsConfig.OpenRegistration {
		l.Trace("account registration is closed, returning error to client")
	}
}

// AccountVerifyGETHandler serves a user's account details to them IF they reached this
// handler while in possession of a valid token, according to the oauth middleware.
func (m *accountModule) AccountVerifyGETHandler(c *gin.Context) {
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
