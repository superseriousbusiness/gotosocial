/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package gotosocial

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// Server is the 'main' function of the gotosocial server, and the place where everything hangs together.
// The logic of stopping and starting the entire server is contained here.
type Server interface {
	// Start starts up the gotosocial server. If something goes wrong
	// while starting the server, then an error will be returned.
	Start(context.Context) error
	// Stop closes down the gotosocial server, first closing the router
	// then the database. If something goes wrong while stopping, an
	// error will be returned.
	Stop(context.Context) error
}

// NewServer returns a new gotosocial server, initialized with the given configuration.
// An error will be returned the caller if something goes wrong during initialization
// eg., no db or storage connection, port for router already in use, etc.
func NewServer(db db.DB, apiRouter router.Router, federator federation.Federator, mediaManager media.Manager) (Server, error) {
	return &gotosocial{
		db:           db,
		apiRouter:    apiRouter,
		federator:    federator,
		mediaManager: mediaManager,
	}, nil
}

// gotosocial fulfils the gotosocial interface.
type gotosocial struct {
	db           db.DB
	apiRouter    router.Router
	federator    federation.Federator
	mediaManager media.Manager
}

// Start starts up the gotosocial server. If something goes wrong
// while starting the server, then an error will be returned.
func (gts *gotosocial) Start(ctx context.Context) error {
	gts.apiRouter.Start()
	return nil
}

// Stop closes down the gotosocial server, first closing the router,
// then the media manager, then the database.
// If something goes wrong while stopping, an error will be returned.
func (gts *gotosocial) Stop(ctx context.Context) error {
	if err := gts.apiRouter.Stop(ctx); err != nil {
		return err
	}
	if err := gts.db.Stop(ctx); err != nil {
		return err
	}
	return nil
}
