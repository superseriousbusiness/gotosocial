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

//go:build moderncsqlite3 || nowasm

package sqlite

import (
	"database/sql/driver"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"code.superseriousbusiness.org/gotosocial/internal/db"
)

// processSQLiteError processes an sqlite3.Error to
// handle conversion to any of our common db types.
func processSQLiteError(err error) error {
	// Attempt to cast as sqlite error.
	sqliteErr, ok := err.(*sqlite.Error)
	if !ok {
		return err
	}

	// Handle supplied error code:
	switch sqliteErr.Code() {
	case sqlite3.SQLITE_CONSTRAINT_UNIQUE,
		sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
		return db.ErrAlreadyExists

	// Busy should be very rare, but
	// on busy tell the database to close
	// the connection, re-open and re-attempt
	// which should give a necessary timeout.
	case sqlite3.SQLITE_BUSY,
		sqlite3.SQLITE_BUSY_RECOVERY,
		sqlite3.SQLITE_BUSY_SNAPSHOT:
		return driver.ErrBadConn
	}

	// Wrap the returned error with the code and
	// extended code for easier debugging later.
	return fmt.Errorf("%w (code=%d)", err,
		sqliteErr.Code(),
	)
}
