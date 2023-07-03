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

package media

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

var List action.GTSAction = func(ctx context.Context) error {
	remote := config.GetAdminMediaListRemote()
	local := config.GetAdminMediaListLocal()

	if !local && !remote {
		return fmt.Errorf("need to pass at least one of --%s, --%s", config.AdminMediaListRemoteFlag(), config.AdminMediaListLocalFlag())
	}

	var state state.State
	state.Caches.Init()
	state.Workers.Start()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	attachments, err := dbConn.GetAttachments(ctx, "", 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve attachments: %w", err)
	}

	f := bufio.NewWriter(os.Stdout)

	mediaPath := config.GetStorageLocalBasePath()

	for _, a := range attachments {
		if local && a.RemoteURL == "" {
			if _, err := f.WriteString(path.Join(mediaPath, a.File.Path) + "\n"); err != nil {
				return err
			}
		}
		if remote && a.RemoteURL != "" {
			if _, err := f.WriteString(a.RemoteURL + "\n"); err != nil {
				return err
			}
		}
	}

	// Explicitly flush instead of deferring it so the log output of
	// dbConn.Stop doesn't get interleaved in the output
	f.Flush()

	return dbConn.Stop(ctx)
}
