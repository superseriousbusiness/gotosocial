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

package prune

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
)

type prune struct {
	dbService db.DB
	storage   *gtsstorage.Driver
	manager   media.Manager
	state     *state.State
}

func setupPrune(ctx context.Context) (*prune, error) {
	var state state.State
	state.Caches.Init()
	state.Workers.Start()

	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbservice: %w", err)
	}
	state.DB = dbService

	//nolint:contextcheck
	storage, err := gtsstorage.AutoConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating storage backend: %w", err)
	}

	state.DB = dbService
	state.Storage = storage

	//nolint:contextcheck
	manager := media.NewManager(&state)

	return &prune{
		dbService: dbService,
		storage:   storage,
		manager:   manager,
		state:     &state,
	}, nil
}

func (p *prune) shutdown(ctx context.Context) error {
	if err := p.storage.Close(); err != nil {
		return fmt.Errorf("error closing storage backend: %w", err)
	}

	if err := p.dbService.Stop(ctx); err != nil {
		return fmt.Errorf("error closing dbservice: %w", err)
	}

	p.state.Caches.Stop()
	p.state.Workers.Stop()

	return nil
}
