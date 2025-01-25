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

//go:build !moderncsqlite3 && !nowasm

package sqlite

import (
	"database/sql/driver"
	"fmt"

	"github.com/ncruces/go-sqlite3"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// processSQLiteError processes an sqlite3.Error to
// handle conversion to any of our common db types.
func processSQLiteError(err error) error {
	// Attempt to cast as sqlite error.
	sqliteErr, ok := err.(*sqlite3.Error)
	if !ok {
		return err
	}

	// Handle supplied error code:
	switch sqliteErr.ExtendedCode() {
	case sqlite3.CONSTRAINT_UNIQUE,
		sqlite3.CONSTRAINT_PRIMARYKEY:
		return db.ErrAlreadyExists

	// Busy should be very rare, but on
	// busy tell the database to close the
	// connection, re-open and re-attempt
	// which should give necessary timeout.
	case sqlite3.BUSY_RECOVERY,
		sqlite3.BUSY_SNAPSHOT:
		return driver.ErrBadConn
	}

	// Wrap the returned error with the code and
	// extended code for easier debugging later.
	return fmt.Errorf("%w (code=%d extended=%d)", err,
		sqliteErr.Code(),
		sqliteErr.ExtendedCode(),
	)
}
