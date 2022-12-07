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

package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

type Provider struct {
	db db.DB
}

// New returns a new middleware Provider.
//
// To attach global middlewares to an engine using this provider, call the UseGlobals function.
func New(db db.DB, oauthServer oauth.Server) *Provider {
	return &Provider{
		db: db,
	}
}

// UseGlobals attaches all global gin middlewares to the given engine.
//
// Global middlewares are those which should be used for every single
// request, and/or stateful middlewares whose state must be consistent
// across requests (like session cookie middleware).
//
// The provided context is only used to select a session from the database,
// which happens once when this function is called, not per request.
func (p *Provider) UseGlobals(ctx context.Context, e *gin.Engine) error {
	session, err := p.db.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("UseGlobals: error getting session from db: %w", err)
	}

	sessionName, err := SessionName()
	if err != nil {
		return fmt.Errorf("UseGlobals: error deriving session name: %w", err)
	}

	// instantiate middlewares that require configuration
	sessionMiddleware := p.Session(sessionName, session.Auth, session.Crypt)
	corsMiddleware := p.Cors(CorsConfig())
	gzipMiddleware := p.Gzip()
	rateLimitMiddleware := p.RateLimit()

	e.Use(
		p.Logger,
		corsMiddleware,
		gzipMiddleware,
		sessionMiddleware,
		rateLimitMiddleware,
		p.UserAgentBlock,
		p.SignatureCheck,
		p.TokenCheck,
		p.ExtraHeaders,
	)

	return nil
}
