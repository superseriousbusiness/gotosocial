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

package gotosocial

import (
	"context"

	"github.com/go-fed/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// Gotosocial is the 'main' function of the gotosocial server, and the place where everything hangs together.
// The logic of stopping and starting the entire server is contained here.
type Gotosocial interface {
	Start(context.Context) error
}

// New returns a new gotosocial server, initialized with the given configuration.
// An error will be returned the caller if something goes wrong during initialization
// eg., no db or storage connection, port for router already in use, etc.
func New(db db.DB, cache cache.Cache, apiRouter router.Router, federationAPI pub.FederatingActor, config *config.Config) (Gotosocial, error) {
	return &gotosocial{
		db:            db,
		cache:         cache,
		apiRouter:     apiRouter,
		federationAPI: federationAPI,
		config:        config,
	}, nil
}

// gotosocial fulfils the gotosocial interface.
type gotosocial struct {
	db            db.DB
	cache         cache.Cache
	apiRouter     router.Router
	federationAPI pub.FederatingActor
	config        *config.Config
}

// Start starts up the gotosocial server. It is a blocking call, so only call it when
// you're absolutely sure you want to start up the server. If something goes wrong
// while starting the server, then an error will be returned. You can treat this function a
// lot like you would treat http.ListenAndServe()
func (gts *gotosocial) Start(ctx context.Context) error {
	return nil
}
