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

package security

import (
	"net/http"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const robotsPath = "/robots.txt"

// Module implements the ClientAPIModule interface for security middleware
type Module struct {
	db     db.DB
	server oauth.Server
}

// New returns a new security module
func New(db db.DB, server oauth.Server) api.ClientModule {
	return &Module{
		db:     db,
		server: server,
	}
}

// Route attaches security middleware to the given router
func (m *Module) Route(s router.Router) error {
	// only enable rate limit middleware if configured
	// advanced-rate-limit-requests is greater than 0
	if rateLimitRequests := config.GetAdvancedRateLimitRequests(); rateLimitRequests > 0 {
		s.AttachMiddleware(m.RateLimit(RateLimitOptions{
			Period: 5 * time.Minute,
			Limit:  int64(rateLimitRequests),
		}))
	}
	s.AttachMiddleware(m.SignatureCheck)
	s.AttachMiddleware(m.FlocBlock)
	s.AttachMiddleware(m.ExtraHeaders)
	s.AttachMiddleware(m.UserAgentBlock)
	s.AttachMiddleware(m.TokenCheck)
	s.AttachHandler(http.MethodGet, robotsPath, m.RobotsGETHandler)
	return nil
}
