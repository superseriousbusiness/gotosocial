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

package prune

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
)

type prune struct {
	dbService db.DB
	storage   *gtsstorage.Driver
	manager   *media.Manager
	cleaner   *cleaner.Cleaner
	state     *state.State
}

func setupPrune(ctx context.Context) (*prune, error) {
	var state state.State

	state.Caches.Init()
	if err := state.Caches.Start(); err != nil {
		return nil, fmt.Errorf("error starting caches: %w", err)
	}

	// Scheduler is required for the
	// cleaner, but no other workers
	// are needed for this CLI action.
	state.Workers.StartScheduler()

	// Set state DB connection.
	// Don't need Actions for this.
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
	state.Storage = storage

	//nolint:contextcheck
	manager := media.NewManager(&state)

	//nolint:contextcheck
	cleaner := cleaner.New(&state)

	return &prune{
		dbService: dbService,
		storage:   storage,
		manager:   manager,
		cleaner:   cleaner,
		state:     &state,
	}, nil
}

func (p *prune) shutdown() error {
	errs := gtserror.NewMultiError(2)

	if err := p.dbService.Close(); err != nil {
		errs.Appendf("error stopping database: %w", err)
	}

	p.state.Workers.Scheduler.Stop()
	p.state.Caches.Stop()

	return errs.Combine()
}
