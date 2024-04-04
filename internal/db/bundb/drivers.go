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
	"database/sql/driver"
	"time"
	_ "unsafe" // linkname shenanigans

	pgx "github.com/jackc/pgx/v5/stdlib"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"modernc.org/sqlite"
)

var (
	// global SQL driver instances.
	postgresDriver = pgx.GetDefaultDriver()
	sqliteDriver   = getSQLiteDriver()

	// check the postgres connection
	// conforms to our conn{} interface.
	// (note SQLite doesn't export their
	// conn type, and gets checked in
	// tests very regularly anywho).
	_ conn = (*pgx.Conn)(nil)
)

//go:linkname getSQLiteDriver modernc.org/sqlite.newDriver
func getSQLiteDriver() *sqlite.Driver

func init() {
	sql.Register("pgx-gts", &PostgreSQLDriver{})
	sql.Register("sqlite-gts", &SQLiteDriver{})
}

// PostgreSQLDriver is our own wrapper around the
// pgx/stdlib.Driver{} type in order to wrap further
// SQL driver types with our own err processing.
type PostgreSQLDriver struct{}

func (d *PostgreSQLDriver) Open(name string) (driver.Conn, error) {
	c, err := postgresDriver.Open(name)
	if err != nil {
		return nil, err
	}
	return &PostgreSQLConn{conn: c.(conn)}, nil
}

type PostgreSQLConn struct{ conn }

func (c *PostgreSQLConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *PostgreSQLConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	tx, err := c.conn.BeginTx(ctx, opts)
	err = processPostgresError(err)
	if err != nil {
		return nil, err
	}
	return &PostgreSQLTx{tx}, nil
}

func (c *PostgreSQLConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *PostgreSQLConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	st, err := c.conn.PrepareContext(ctx, query)
	err = processPostgresError(err)
	if err != nil {
		return nil, err
	}
	return &PostgreSQLStmt{stmt: st.(stmt)}, nil
}

func (c *PostgreSQLConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, toNamedValues(args))
}

func (c *PostgreSQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	result, err := c.conn.ExecContext(ctx, query, args)
	err = processPostgresError(err)
	return result, err
}

func (c *PostgreSQLConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.QueryContext(context.Background(), query, toNamedValues(args))
}

func (c *PostgreSQLConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows, err := c.conn.QueryContext(ctx, query, args)
	err = processPostgresError(err)
	return rows, err
}

func (c *PostgreSQLConn) Close() error {
	return c.conn.Close()
}

type PostgreSQLTx struct{ driver.Tx }

func (tx *PostgreSQLTx) Commit() error {
	err := tx.Tx.Commit()
	return processPostgresError(err)
}

func (tx *PostgreSQLTx) Rollback() error {
	err := tx.Tx.Rollback()
	return processPostgresError(err)
}

type PostgreSQLStmt struct{ stmt }

func (stmt *PostgreSQLStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), toNamedValues(args))
}

func (stmt *PostgreSQLStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	res, err := stmt.stmt.ExecContext(ctx, args)
	err = processPostgresError(err)
	return res, err
}

func (stmt *PostgreSQLStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), toNamedValues(args))
}

func (stmt *PostgreSQLStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, err := stmt.stmt.QueryContext(ctx, args)
	err = processPostgresError(err)
	return rows, err
}

// SQLiteDriver is our own wrapper around the
// sqlite.Driver{} type in order to wrap further
// SQL driver types with our own functionality,
// e.g. hooks, retries and err processing.
type SQLiteDriver struct{}

func (d *SQLiteDriver) Open(name string) (driver.Conn, error) {
	c, err := sqliteDriver.Open(name)
	if err != nil {
		return nil, err
	}
	return &SQLiteConn{conn: c.(conn)}, nil
}

type SQLiteConn struct{ conn }

func (c *SQLiteConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *SQLiteConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	err = retryOnBusy(ctx, func() error {
		tx, err = c.conn.BeginTx(ctx, opts)
		err = processSQLiteError(err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SQLiteTx{Context: ctx, Tx: tx}, nil
}

func (c *SQLiteConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *SQLiteConn) PrepareContext(ctx context.Context, query string) (st driver.Stmt, err error) {
	err = retryOnBusy(ctx, func() error {
		st, err = c.conn.PrepareContext(ctx, query)
		err = processSQLiteError(err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SQLiteStmt{st.(stmt)}, nil
}

func (c *SQLiteConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, toNamedValues(args))
}

func (c *SQLiteConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	err = retryOnBusy(ctx, func() error {
		result, err = c.conn.ExecContext(ctx, query, args)
		err = processSQLiteError(err)
		return err
	})
	return
}

func (c *SQLiteConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.QueryContext(context.Background(), query, toNamedValues(args))
}

func (c *SQLiteConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	err = retryOnBusy(ctx, func() error {
		rows, err = c.conn.QueryContext(ctx, query, args)
		defer rows.Close()
		err = processSQLiteError(err)
		return err
	})
	return
}

func (c *SQLiteConn) Close() error {
	// see: https://www.sqlite.org/pragma.html#pragma_optimize
	const onClose = "PRAGMA analysis_limit=1000; PRAGMA optimize;"
	_, _ = c.conn.ExecContext(context.Background(), onClose, nil)
	return c.conn.Close()
}

type SQLiteTx struct {
	context.Context
	driver.Tx
}

func (tx *SQLiteTx) Commit() (err error) {
	err = retryOnBusy(tx.Context, func() error {
		err = tx.Tx.Commit()
		err = processSQLiteError(err)
		return err
	})
	return
}

func (tx *SQLiteTx) Rollback() (err error) {
	err = retryOnBusy(tx.Context, func() error {
		err = tx.Tx.Rollback()
		err = processSQLiteError(err)
		return err
	})
	return
}

type SQLiteStmt struct{ stmt }

func (stmt *SQLiteStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), toNamedValues(args))
}

func (stmt *SQLiteStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	err = retryOnBusy(ctx, func() error {
		res, err = stmt.stmt.ExecContext(ctx, args)
		err = processSQLiteError(err)
		return err
	})
	return
}

func (stmt *SQLiteStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), toNamedValues(args))
}

func (stmt *SQLiteStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	err = retryOnBusy(ctx, func() error {
		rows, err = stmt.stmt.QueryContext(ctx, args)
		defer rows.Close()
		err = processSQLiteError(err)
		return err
	})
	return
}

type conn interface {
	driver.Conn
	driver.ConnPrepareContext
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnBeginTx
}

type stmt interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

// retryOnBusy will retry given function on returned 'errBusy'.
func retryOnBusy(ctx context.Context, fn func() error) error {
	if err := fn(); err != errBusy {
		return err
	}
	return retryOnBusySlow(ctx, fn)
}

// retryOnBusySlow is the outlined form of retryOnBusy, to allow the fast path (i.e. only
// 1 attempt) to be inlined, leaving the slow retry loop to be a separate function call.
func retryOnBusySlow(ctx context.Context, fn func() error) error {
	var backoff time.Duration

	for i := 0; ; i++ {
		// backoff according to a multiplier of 2ms * 2^2n,
		// up to a maximum possible backoff time of 5 minutes.
		//
		// this works out as the following:
		// 4ms
		// 16ms
		// 64ms
		// 256ms
		// 1.024s
		// 4.096s
		// 16.384s
		// 1m5.536s
		// 4m22.144s
		backoff = 2 * time.Millisecond * (1 << (2*i + 1))
		if backoff >= 5*time.Minute {
			break
		}

		select {
		// Context cancelled.
		case <-ctx.Done():
			return ctx.Err()

		// Backoff for some time.
		case <-time.After(backoff):
		}

		// Perform func.
		err := fn()

		if err != errBusy {
			// May be nil, or may be
			// some other error, either
			// way return here.
			return err
		}
	}

	return gtserror.Newf("%w (waited > %s)", db.ErrBusyTimeout, backoff)
}

// toNamedValues converts older driver.Value types to driver.NamedValue types.
func toNamedValues(args []driver.Value) []driver.NamedValue {
	if args == nil {
		return nil
	}
	args2 := make([]driver.NamedValue, len(args))
	for i := range args {
		args2[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   args[i],
		}
	}
	return args2
}
