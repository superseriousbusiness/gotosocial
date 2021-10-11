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

package security

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const robotsPath = "/robots.txt"

// Module implements the ClientAPIModule interface for security middleware
type Module struct {
	config *config.Config
	db     db.DB
}

// New returns a new security module
func New(config *config.Config, db db.DB) api.ClientModule {
	return &Module{
		config: config,
		db:     db,
	}
}

// Route attaches security middleware to the given router
func (m *Module) Route(s router.Router) error {
	s.AttachMiddleware(m.SignatureCheck)
	s.AttachMiddleware(m.FlocBlock)
	s.AttachMiddleware(m.ExtraHeaders)
	s.AttachMiddleware(m.UserAgentBlock)
	s.AttachHandler(http.MethodGet, robotsPath, m.RobotsGETHandler)
	return nil
}
