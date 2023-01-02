/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/auth"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Auth struct {
	routerSession *gtsmodel.RouterSession
	sessionName   string

	auth *auth.Module
}

// Route attaches 'auth' and 'oauth' groups to the given router.
func (a *Auth) Route(r router.Router, m ...gin.HandlerFunc) {
	// create groupings for the 'auth' and 'oauth' prefixes
	authGroup := r.AttachGroup("auth")
	oauthGroup := r.AttachGroup("oauth")

	// instantiate + attach shared, non-global middlewares to both of these groups
	var (
		cacheControlMiddleware = middleware.CacheControl("private", "max-age=120")
		sessionMiddleware      = middleware.Session(a.sessionName, a.routerSession.Auth, a.routerSession.Crypt)
	)
	authGroup.Use(m...)
	oauthGroup.Use(m...)
	authGroup.Use(cacheControlMiddleware, sessionMiddleware)
	oauthGroup.Use(cacheControlMiddleware, sessionMiddleware)

	a.auth.RouteAuth(authGroup.Handle)
	a.auth.RouteOauth(oauthGroup.Handle)
}

func NewAuth(db db.DB, p processing.Processor, idp oidc.IDP, routerSession *gtsmodel.RouterSession, sessionName string) *Auth {
	return &Auth{
		routerSession: routerSession,
		sessionName:   sessionName,
		auth:          auth.New(db, p, idp),
	}
}
