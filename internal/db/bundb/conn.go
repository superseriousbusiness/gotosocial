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
	"context"
	"database/sql"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// DBConn wrapps a bun.DB conn to provide SQL-type specific additional functionality
type DBConn struct {
	errProc func(error) error // errProc is the SQL-type specific error processor
	*bun.DB                   // DB is the underlying bun.DB connection
}

// WrapDBConn wraps a bun DB connection to provide our own error processing dependent on DB dialect.
func WrapDBConn(dbConn *bun.DB) *DBConn {
	var errProc func(error) error
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

func (conn *DBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx bun.Tx, err error) {
	err = retryOnBusy(ctx, func() error {
		tx, err = conn.DB.BeginTx(ctx, opts)
		err = conn.ProcessError(err)
		return err
	})
	return
}

func (conn *DBConn) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	err = retryOnBusy(ctx, func() error {
		result, err = conn.DB.ExecContext(ctx, query, args...)
		err = conn.ProcessError(err)
		return err
	})
	return
}

func (conn *DBConn) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	err = retryOnBusy(ctx, func() error {
		rows, err = conn.DB.QueryContext(ctx, query, args...)
		err = conn.ProcessError(err)
		return err
	})
	return
}

func (conn *DBConn) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	_ = retryOnBusy(ctx, func() error {
		row = conn.DB.QueryRowContext(ctx, query, args...)
		err := conn.ProcessError(row.Err())
		return err
	})
	return
}

func (conn *DBConn) RunInTx(ctx context.Context, fn func(bun.Tx) error) error {
	// Attempt to start new transaction.
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			// Rollback (with retry-backoff).
			_ = retryOnBusy(ctx, func() error {
				err := tx.Rollback()
				return conn.errProc(err)
			})
		}
	}()

	// Perform supplied transaction
	if err := fn(tx); err != nil {
		return conn.errProc(err)
	}

	// Commit (with retry-backoff).
	err = retryOnBusy(ctx, func() error {
		err := tx.Commit()
		return conn.errProc(err)
	})
	done = true
	return err
}

func (conn *DBConn) NewValues(model interface{}) *bun.ValuesQuery {
	return bun.NewValuesQuery(conn.DB, model).Conn(conn)
}

func (conn *DBConn) NewMerge() *bun.MergeQuery {
	return bun.NewMergeQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewSelect() *bun.SelectQuery {
	return bun.NewSelectQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewInsert() *bun.InsertQuery {
	return bun.NewInsertQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewUpdate() *bun.UpdateQuery {
	return bun.NewUpdateQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewDelete() *bun.DeleteQuery {
	return bun.NewDeleteQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	return bun.NewRawQuery(conn.DB, query, args...).Conn(conn)
}

func (conn *DBConn) NewCreateTable() *bun.CreateTableQuery {
	return bun.NewCreateTableQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewDropTable() *bun.DropTableQuery {
	return bun.NewDropTableQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewCreateIndex() *bun.CreateIndexQuery {
	return bun.NewCreateIndexQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewDropIndex() *bun.DropIndexQuery {
	return bun.NewDropIndexQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewTruncateTable() *bun.TruncateTableQuery {
	return bun.NewTruncateTableQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewAddColumn() *bun.AddColumnQuery {
	return bun.NewAddColumnQuery(conn.DB).Conn(conn)
}

func (conn *DBConn) NewDropColumn() *bun.DropColumnQuery {
	return bun.NewDropColumnQuery(conn.DB).Conn(conn)
}

// ProcessError processes an error to replace any known values with our own error types,
// making it easier to catch specific situations (e.g. no rows, already exists, etc)
func (conn *DBConn) ProcessError(err error) error {
	if err == nil {
		return err
	}
	return conn.errProc(err)
}

// Exists checks the results of a SelectQuery for the existence of the data in question, masking ErrNoEntries errors
func (conn *DBConn) Exists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
	exists, err := query.Exists(ctx)
	switch err {
	case nil:
		return exists, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// NotExists is the functional opposite of conn.Exists()
func (conn *DBConn) NotExists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
	exists, err := conn.Exists(ctx, query)
	return !exists, err
}

// retryOnBusy will retry given function on returned db.ErrBusyTimeout.
func retryOnBusy(ctx context.Context, fn func() error) error {
	const (
		// max no. attempts.
		maxRetries = 10

		// base backoff duration multiplier.
		baseBackoff = 2 * time.Millisecond
	)

	for i := 0; i < maxRetries; i += 2 {
		// Perform func.
		err := fn()

		if err != errBusy {
			// May be nil, or may be
			// some other error, either
			// way return here.
			return err
		}

		// backoff according to a multiplier of 2^n.
		backoff := baseBackoff * (1 << (i + 1))

		select {
		// Context cancelled.
		case <-ctx.Done():

		// Backoff for some time.
		case <-time.After(backoff):
		}
	}

	return gtserror.Newf("%w (waited > %s)", db.ErrBusyTimeout, baseBackoff*(1<<maxRetries))
}
