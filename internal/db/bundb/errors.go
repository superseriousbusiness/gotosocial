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

package bundb

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// errBusy is a sentinel error indicating
// busy database (e.g. retry needed).
var errBusy = errors.New("busy")

// processPostgresError processes an error, replacing any postgres specific errors with our own error type
func processPostgresError(err error) error {
	// Attempt to cast as postgres
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return err
	}

	// Handle supplied error code:
	// (https://www.postgresql.org/docs/10/errcodes-appendix.html)
	switch pgErr.Code { //nolint
	case "23505" /* unique_violation */ :
		return db.ErrAlreadyExists
	}

	return err
}

// processSQLiteError processes an error, replacing any sqlite specific errors with our own error type
func processSQLiteError(err error) error {
	// Attempt to cast as sqlite
	sqliteErr, ok := err.(*sqlite.Error)
	if !ok {
		return err
	}

	// Handle supplied error code:
	switch sqliteErr.Code() {
	case sqlite3.SQLITE_CONSTRAINT_UNIQUE,
		sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
		return db.ErrAlreadyExists
	case sqlite3.SQLITE_BUSY:
		return errBusy
	case sqlite3.SQLITE_BUSY_TIMEOUT:
		return db.ErrBusyTimeout
	}

	return err
}
