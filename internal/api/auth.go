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

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/auth"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

type Auth struct {
	routerSession *gtsmodel.RouterSession
	sessionName   string

	auth *auth.Module
}

// Route attaches 'auth' and 'oauth' groups to the given router.
func (a *Auth) Route(r *router.Router, m ...gin.HandlerFunc) {
	// create groupings for the 'auth' and 'oauth' prefixes
	authGroup := r.AttachGroup("auth")
	oauthGroup := r.AttachGroup("oauth")

	// instantiate + attach shared, non-global middlewares to both of these groups
	var (
		ccMiddleware = middleware.CacheControl(middleware.CacheControlConfig{
			Directives: []string{"private", "max-age=120"},
			Vary:       []string{"Accept", "Accept-Encoding"},
		})
		sessionMiddleware = middleware.Session(a.sessionName, a.routerSession.Auth, a.routerSession.Crypt)
	)
	authGroup.Use(m...)
	oauthGroup.Use(m...)
	authGroup.Use(ccMiddleware, sessionMiddleware)
	oauthGroup.Use(ccMiddleware, sessionMiddleware)

	a.auth.RouteAuth(authGroup.Handle)
	a.auth.RouteOAuth(oauthGroup.Handle)
}

func NewAuth(
	state *state.State,
	p *processing.Processor,
	idp oidc.IDP,
	routerSession *gtsmodel.RouterSession,
	sessionName string,
) *Auth {
	return &Auth{
		routerSession: routerSession,
		sessionName:   sessionName,
		auth:          auth.New(state, p, idp),
	}
}
