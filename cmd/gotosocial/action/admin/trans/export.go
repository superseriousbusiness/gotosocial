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

package trans

import (
	"context"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/trans"
)

// Export exports info from the database into a file
var Export action.GTSAction = func(ctx context.Context) error {
	var state state.State

	// Only set state DB connection.
	// Don't need Actions or Workers for this.
	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}
	state.DB = dbConn

	exporter := trans.NewExporter(dbConn)

	path := config.GetAdminTransPath()
	if path == "" {
		return errors.New("no path set")
	}

	if err := exporter.ExportMinimal(ctx, path); err != nil {
		return err
	}

	return dbConn.Close()
}
