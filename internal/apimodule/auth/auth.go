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

package auth

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// AuthSignInPath is the API path for users to sign in through
	AuthSignInPath     = "/auth/sign_in"
	// OauthTokenPath is the API path to use for granting token requests to users with valid credentials
	OauthTokenPath     = "/oauth/token"
	// OauthAuthorizePath is the API path for authorization requests (eg., authorize this app to act on my behalf as a user)
	OauthAuthorizePath = "/oauth/authorize"
)

// Module implements the ClientAPIModule interface for
type Module struct {
	server oauth.Server
	db     db.DB
	log    *logrus.Logger
}

// New returns a new auth module
func New(srv oauth.Server, db db.DB, log *logrus.Logger) apimodule.ClientAPIModule {
	return &Module{
		server: srv,
		db:     db,
		log:    log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, AuthSignInPath, m.SignInGETHandler)
	s.AttachHandler(http.MethodPost, AuthSignInPath, m.SignInPOSTHandler)

	s.AttachHandler(http.MethodPost, OauthTokenPath, m.TokenPOSTHandler)

	s.AttachHandler(http.MethodGet, OauthAuthorizePath, m.AuthorizeGETHandler)
	s.AttachHandler(http.MethodPost, OauthAuthorizePath, m.AuthorizePOSTHandler)

	s.AttachMiddleware(m.OauthTokenMiddleware)
	return nil
}

// CreateTables creates the necessary tables for this module in the given database
func (m *Module) CreateTables(db db.DB) error {
	models := []interface{}{
		&oauth.Client{},
		&oauth.Token{},
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Application{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}
