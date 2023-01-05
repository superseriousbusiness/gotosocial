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

package bundb

import (
	"context"
	"database/sql"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// DBConn wrapps a bun.DB conn to provide SQL-type specific additional functionality
type DBConn struct {
	errProc func(error) db.Error // errProc is the SQL-type specific error processor
	*bun.DB                      // DB is the underlying bun.DB connection
}

// WrapDBConn wraps a bun DB connection to provide our own error processing dependent on DB dialect.
func WrapDBConn(dbConn *bun.DB) *DBConn {
	var errProc func(error) db.Error
	switch dbConn.Dialect().Name() {
	case dialect.PG:
		errProc = processPostgresError
	case dialect.SQLite:
		errProc = processSQLiteError
	default:
		panic("unknown dialect name: " + dbConn.Dialect().Name().String())
	}
	return &DBConn{
		errProc: errProc,
		DB:      dbConn,
	}
}

// RunInTx wraps execution of the supplied transaction function.
func (conn *DBConn) RunInTx(ctx context.Context, fn func(bun.Tx) error) db.Error {
	return conn.ProcessError(func() error {
		// Acquire a new transaction
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		var done bool

		defer func() {
			if !done {
				_ = tx.Rollback()
			}
		}()

		// Perform supplied transaction
		if err := fn(tx); err != nil {
			return err
		}

		// Finally, commit
		err = tx.Commit() //nolint:contextcheck
		done = true
		return err
	}())
}

// ProcessError processes an error to replace any known values with our own db.Error types,
// making it easier to catch specific situations (e.g. no rows, already exists, etc)
func (conn *DBConn) ProcessError(err error) db.Error {
	switch {
	case err == nil:
		return nil
	case err == sql.ErrNoRows:
		return db.ErrNoEntries
	default:
		return conn.errProc(err)
	}
}

// Exists checks the results of a SelectQuery for the existence of the data in question, masking ErrNoEntries errors
func (conn *DBConn) Exists(ctx context.Context, query *bun.SelectQuery) (bool, db.Error) {
	exists, err := query.Exists(ctx)

	// Process error as our own and check if it exists
	switch err := conn.ProcessError(err); err {
	case nil:
		return exists, nil
	case db.ErrNoEntries:
		return false, nil
	default:
		return false, err
	}
}

// NotExists is the functional opposite of conn.Exists()
func (conn *DBConn) NotExists(ctx context.Context, query *bun.SelectQuery) (bool, db.Error) {
	exists, err := conn.Exists(ctx, query)
	return !exists, err
}
