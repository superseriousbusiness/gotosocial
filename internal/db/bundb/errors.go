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
	"database/sql/driver"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// errBusy is a sentinel error indicating
// busy database (e.g. retry needed).
var errBusy = errors.New("busy")

// processPostgresError processes an error, replacing any postgres specific errors with our own error type
func processPostgresError(err error) error {
	// Catch nil errs.
	if err == nil {
		return nil
	}

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
	// Catch nil errs.
	if err == nil {
		return nil
	}

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
	case sqlite3.SQLITE_BUSY,
		sqlite3.SQLITE_BUSY_SNAPSHOT,
		sqlite3.SQLITE_BUSY_RECOVERY:
		return errBusy
	case sqlite3.SQLITE_BUSY_TIMEOUT:
		return db.ErrBusyTimeout

	// WORKAROUND:
	// text copied from matrix dev chat:
	//
	// okay i've found a workaround for now. so between
	// v1.29.0 and v1.29.2 (modernc.org/sqlite) is that
	// slightly tweaked interruptOnDone() behaviour, which
	// causes interrupt to (imo, correctly) get called when
	// a context is cancelled to cancel the running query. the
	// issue is that every single query after that point seems
	// to still then return interrupted. so as you thought,
	// maybe that query count isn't being decremented. i don't
	// think it's our code, but i haven't ruled it out yet.
	//
	// the workaround for now is adding to our sqlite error
	// processor to replace an SQLITE_INTERRUPTED code with
	// driver.ErrBadConn, which hints to the golang sql package
	// that the conn needs to be closed and a new one opened
	//
	case sqlite3.SQLITE_INTERRUPT:
		return driver.ErrBadConn
	}

	return err
}
