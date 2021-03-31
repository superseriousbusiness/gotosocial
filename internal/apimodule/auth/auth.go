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

// Package auth is a module that provides oauth functionality to a router.
// It adds the following paths:
//    /auth/sign_in
//    /oauth/token
//    /oauth/authorize
// It also includes the oauthTokenMiddleware, which can be attached to a router to authenticate every request by Bearer token.
package auth

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	authSignInPath     = "/auth/sign_in"
	oauthTokenPath     = "/oauth/token"
	oauthAuthorizePath = "/oauth/authorize"
)

type authModule struct {
	server oauth.Server
	db     db.DB
	log    *logrus.Logger
}

// New returns a new auth module
func New(srv oauth.Server, db db.DB, log *logrus.Logger) apimodule.ClientAPIModule {
	return &authModule{
		server: srv,
		db:     db,
		log:    log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *authModule) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, authSignInPath, m.signInGETHandler)
	s.AttachHandler(http.MethodPost, authSignInPath, m.signInPOSTHandler)

	s.AttachHandler(http.MethodPost, oauthTokenPath, m.tokenPOSTHandler)

	s.AttachHandler(http.MethodGet, oauthAuthorizePath, m.authorizeGETHandler)
	s.AttachHandler(http.MethodPost, oauthAuthorizePath, m.authorizePOSTHandler)

	s.AttachMiddleware(m.oauthTokenMiddleware)
	return nil
}
