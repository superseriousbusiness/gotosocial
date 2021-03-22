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
	"github.com/gotosocial/gotosocial/internal/cache"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/gotosocial/internal/router"
)

type Gotosocial interface {
	Start(context.Context) error
	Stop(context.Context) error
}

func New(db db.DB, cache cache.Cache, apiRouter router.Router, federationAPI pub.FederatingActor, config *config.Config) (Gotosocial, error) {
	return &gotosocial{
		db:            db,
		cache:         cache,
		apiRouter:     apiRouter,
		federationAPI: federationAPI,
		config:        config,
	}, nil
}

type gotosocial struct {
	db            db.DB
	cache         cache.Cache
	apiRouter     router.Router
	federationAPI pub.FederatingActor
	config        *config.Config
}

func (gts *gotosocial) Start(ctx context.Context) error {
	return nil
}

func (gts *gotosocial) Stop(ctx context.Context) error {
	return nil
}
