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

package migration

import (
	"context"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/state"
)

// Run will initialize the database, running any available migrations.
var Run action.GTSAction = func(ctx context.Context) error {
	var state state.State

	defer func() {
		if state.DB != nil {
			// Lastly, if database service was started,
			// ensure it gets closed now all else stopped.
			if err := state.DB.Close(); err != nil {
				log.Errorf(ctx, "error stopping database: %v", err)
			}
		}

		// Finally reached end of shutdown.
		log.Info(ctx, "done! exiting...")
	}()

	// Initialize caches
	state.Caches.Init()
	if err := state.Caches.Start(); err != nil {
		return fmt.Errorf("error starting caches: %w", err)
	}

	log.Info(ctx, "starting db service...")

	// Open connection to the database now caches started.
	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set DB on state.
	state.DB = dbService

	return nil
}
