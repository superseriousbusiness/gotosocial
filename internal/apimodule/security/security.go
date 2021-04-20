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
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// Module implements the ClientAPIModule interface for security middleware
type Module struct {
	config *config.Config
	log    *logrus.Logger
}

// New returns a new security module
func New(config *config.Config, log *logrus.Logger) apimodule.ClientAPIModule {
	return &Module{
		config: config,
		log:    log,
	}
}

// Route attaches security middleware to the given router
func (m *Module) Route(s router.Router) error {
	s.AttachMiddleware(m.FlocBlock)
	return nil
}

// CreateTables doesn't do diddly squat at the moment, it's just for fulfilling the interface
func (m *Module) CreateTables(db db.DB) error {
	return nil
}
