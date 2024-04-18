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
	"github.com/ncruces/go-sqlite3"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/sqlite"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

var (
	// global wrapped gts driver instances.
	gtsPostgresDriver = &PostgreSQLDriver{}
	gtsSQLiteDriver   = &SQLiteDriver{}

	// global PostgreSQL driver instances.
	postgresDriver = pgx.GetDefaultDriver()

	// global SQLite3 driver instance.
	sqliteDriver = &sqlite.Driver{
		Init: func(c *sqlite3.Conn) error {
			// unset an busy handler.
			return c.BusyHandler(nil)
		},
	}

	// check the postgres connection
	// conforms to our conn{} interface.
	// (note SQLite doesn't export their
	// conn type, and gets checked in
	// tests very regularly anywho).
	_ conn = (*pgx.Conn)(nil)
)

func init() {
	sql.Register("pgx-gts", gtsPostgresDriver)
	sql.Register("sqlite-gts", gtsSQLiteDriver)
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
	cc, err := d.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return cc.Connect(context.Background())
}

func (d *SQLiteDriver) OpenConnector(name string) (driver.Connector, error) {
	cc, err := sqliteDriver.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &SQLiteConnector{cc}, nil
}

type SQLiteConnector struct{ driver.Connector }

func (c *SQLiteConnector) Driver() driver.Driver { return gtsSQLiteDriver }

func (c *SQLiteConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &SQLiteConn{conn.(sqlite.ConnIface)}, nil
}

type SQLiteConn struct{ sqlite.ConnIface }

func (c *SQLiteConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *SQLiteConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	err = retryOnBusy(ctx, func() error {
		tx, err = c.ConnIface.BeginTx(ctx, opts)
		err = processSQLiteError(err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SQLiteTx{tx}, nil
}

func (c *SQLiteConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *SQLiteConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	err = retryOnBusy(ctx, func() error {
		stmt, err = c.ConnIface.PrepareContext(ctx, query)
		err = processSQLiteError(err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SQLiteStmt{StmtIface: stmt.(sqlite.StmtIface)}, nil
}

func (c *SQLiteConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, toNamedValues(args))
}

func (c *SQLiteConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	st, err := c.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	stmt := st.(*SQLiteStmt)
	res, err = stmt.ExecContext(ctx, args)
	return
}

func (c *SQLiteConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.QueryContext(context.Background(), query, toNamedValues(args))
}

func (c *SQLiteConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	st, err := c.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	stmt := st.(*SQLiteStmt)
	err = retryOnBusy(ctx, func() error {
		rows, err = stmt.StmtIface.QueryContext(ctx, args)
		err = processSQLiteError(err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SQLiteTmpStmtRows{
		SQLiteRows: SQLiteRows{
			Ctx:       ctx,
			RowsIface: rows.(sqlite.RowsIface),
		},
		SQLiteStmt: stmt,
	}, nil
}

func (c *SQLiteConn) Close() (err error) {
	ctx := context.Background()

	// Get acces the underlying raw sqlite3 conn.
	raw := c.ConnIface.(sqlite3.DriverConn).Raw()

	// Set a timeout context to limit execution time.
	ctx, cncl := context.WithTimeout(ctx, 5*time.Second)
	old := raw.SetInterrupt(ctx)

	// see: https://www.sqlite.org/pragma.html#pragma_optimize
	const onClose = "PRAGMA analysis_limit=1000; PRAGMA optimize;"
	_ = raw.Exec(onClose)

	// Unset timeout context.
	_ = raw.SetInterrupt(old)
	cncl()

	// Finally, release + close.
	_ = raw.ReleaseMemory()
	err = raw.Close()
	return
}

type SQLiteTx struct{ driver.Tx }

func (tx *SQLiteTx) Commit() (err error) {
	// use background ctx as this commit MUST happen.
	return retryOnBusy(context.Background(), func() error {
		err = tx.Tx.Commit()
		err = processSQLiteError(err)
		return err
	})
}

func (tx *SQLiteTx) Rollback() (err error) {
	// use background ctx as this rollback MUST happen.
	return retryOnBusy(context.Background(), func() error {
		err = tx.Tx.Rollback()
		err = processSQLiteError(err)
		return err
	})
}

type SQLiteStmt struct{ sqlite.StmtIface }

func (stmt *SQLiteStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), toNamedValues(args))
}

func (stmt *SQLiteStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	err = retryOnBusy(ctx, func() error {
		res, err = stmt.StmtIface.ExecContext(ctx, args)
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
		rows, err = stmt.StmtIface.QueryContext(ctx, args)
		err = processSQLiteError(err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SQLiteRows{
		Ctx:       ctx,
		RowsIface: rows.(sqlite.RowsIface),
	}, nil
}

func (stmt *SQLiteStmt) Close() (err error) {
	// use background ctx as this stmt MUST be closed.
	err = retryOnBusy(context.Background(), func() error {
		err = stmt.StmtIface.Close()
		err = processSQLiteError(err)
		return err
	})
	return
}

type SQLiteRows struct {
	Ctx context.Context
	sqlite.RowsIface
}

func (r *SQLiteRows) Next(dest []driver.Value) (err error) {
	err = retryOnBusy(r.Ctx, func() error {
		err = r.RowsIface.Next(dest)
		err = processSQLiteError(err)
		return err
	})
	return err
}

func (r *SQLiteRows) Close() (err error) {
	// use background ctx as these rows MUST be closed.
	err = retryOnBusy(context.Background(), func() error {
		err = r.RowsIface.Close()
		err = processSQLiteError(err)
		return err
	})
	return
}

type SQLiteTmpStmtRows struct {
	SQLiteRows
	*SQLiteStmt
}

func (r *SQLiteTmpStmtRows) Close() (err error) {
	err = r.SQLiteRows.Close()
	_ = r.SQLiteStmt.Close()
	return err
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

type rows interface {
	driver.Rows
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
