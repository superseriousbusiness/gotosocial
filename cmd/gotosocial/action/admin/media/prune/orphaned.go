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

package prune

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
)

// Orphaned prunes orphaned media from storage.
var Orphaned action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	storage, err := gtsstorage.AutoConfig()
	if err != nil {
		return fmt.Errorf("error creating storage backend: %w", err)
	}

	manager, err := media.NewManager(dbService, storage)
	if err != nil {
		return fmt.Errorf("error instantiating mediamanager: %s", err)
	}

	dry := config.GetAdminMediaPruneDryRun()

	pruned, err := manager.PruneOrphaned(ctx, dry)
	if err != nil {
		return fmt.Errorf("error pruning: %s", err)
	}

	if dry /* dick heyyoooooo */ {
		log.Infof("DRY RUN: %d stored items are orphaned and eligible to be pruned", pruned)
	} else {
		log.Infof("%d stored items were orphaned and pruned", pruned)
	}

	if err := storage.Close(); err != nil {
		return fmt.Errorf("error closing storage backend: %w", err)
	}

	if err := dbService.Stop(ctx); err != nil {
		return fmt.Errorf("error closing dbservice: %s", err)
	}

	return nil
}
