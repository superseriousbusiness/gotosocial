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

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Remote prunes old and/or unused remote media.
var Remote action.GTSAction = func(ctx context.Context) error {
	prune, err := setupPrune(ctx)
	if err != nil {
		return err
	}

	dry := config.GetAdminMediaPruneDryRun()

	pruned, err := prune.manager.PruneUnusedRemote(ctx, dry)
	if err != nil {
		return fmt.Errorf("error pruning: %w", err)
	}

	uncached, err := prune.manager.UncacheRemote(ctx, config.GetMediaRemoteCacheDays(), dry)
	if err != nil {
		return fmt.Errorf("error pruning: %w", err)
	}

	total := pruned + uncached

	if dry /* dick heyyoooooo */ {
		log.Infof(ctx, "DRY RUN: %d remote items are unused/stale and eligible to be pruned", total)
	} else {
		log.Infof(ctx, "%d unused/stale remote items were pruned", pruned)
	}

	return prune.shutdown(ctx)
}
