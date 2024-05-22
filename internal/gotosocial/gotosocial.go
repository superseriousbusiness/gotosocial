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

package gotosocial

import (
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// Server represents a long-running
// GoToSocial server instance.
type Server struct {
	db        db.DB
	apiRouter *router.Router
	cleaner   *cleaner.Cleaner
}

// NewServer returns a new
// GoToSocial server instance.
func NewServer(
	db db.DB,
	apiRouter *router.Router,
	cleaner *cleaner.Cleaner,
) *Server {
	return &Server{
		db:        db,
		apiRouter: apiRouter,
		cleaner:   cleaner,
	}
}

// Start starts up the GoToSocial server by starting the router,
// then the cleaner. If something goes wrong while starting the
// server, then an error will be returned.
func (s *Server) Start() error {
	s.apiRouter.Start()
	return s.cleaner.ScheduleJobs()
}

// Stop closes down the GoToSocial server, first closing the cleaner,
// then the router, then the database. If something goes wrong while
// stopping, an error will be returned.
func (s *Server) Stop() error {
	if err := s.apiRouter.Stop(); err != nil {
		return err
	}
	return s.db.Close()
}
