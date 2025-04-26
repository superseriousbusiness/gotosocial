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

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// All performs all media clean actions
var All action.GTSAction = func(ctx context.Context) error {
	// Setup pruning utilities.
	prune, err := setupPrune(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure pruner gets shutdown on exit.
		if err := prune.shutdown(); err != nil {
			log.Error(ctx, err)
		}
	}()

	if config.GetAdminMediaPruneDryRun() {
		log.Info(ctx, "prune DRY RUN")
		ctx = gtscontext.SetDryRun(ctx)
	}

	days := config.GetMediaRemoteCacheDays()

	// Perform the actual pruning with logging.
	prune.cleaner.Media().All(ctx, days)
	prune.cleaner.Emoji().All(ctx, days)

	// Perform a cleanup of storage (for removed local dirs).
	if err := prune.storage.Storage.Clean(ctx); err != nil {
		log.Error(ctx, "error cleaning storage: %v", err)
	}

	return nil
}
