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
	"github.com/superseriousbusiness/gotosocial/internal/db/sqlite"
)

var (
	// global wrapped gts driver instances.
	gtsPostgresDriver = &PostgreSQLDriver{}
	gtsSQLiteDriver   = &SQLiteDriver{}

	// global PostgreSQL driver instances.
	postgresDriver = pgx.GetDefaultDriver()

	// global SQLite3 driver instance.
	sqliteDriver = &sqlite.Driver{}

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
	tx, err = c.ConnIface.BeginTx(ctx, opts)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &SQLiteTx{tx}, nil
}

func (c *SQLiteConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *SQLiteConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	stmt, err = c.ConnIface.PrepareContext(ctx, query)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &SQLiteStmt{StmtIface: stmt.(sqlite.StmtIface)}, nil
}

func (c *SQLiteConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, toNamedValues(args))
}

func (c *SQLiteConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	res, err = c.ConnIface.ExecContext(ctx, query, args)
	err = processSQLiteError(err)
	return
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

	// Finally, close.
	err = raw.Close()
	return
}

type SQLiteTx struct{ driver.Tx }

func (tx *SQLiteTx) Commit() (err error) {
	err = tx.Tx.Commit()
	err = processSQLiteError(err)
	return
}

func (tx *SQLiteTx) Rollback() (err error) {
	err = tx.Tx.Rollback()
	err = processSQLiteError(err)
	return
}

type SQLiteStmt struct{ sqlite.StmtIface }

func (stmt *SQLiteStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), toNamedValues(args))
}

func (stmt *SQLiteStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	res, err = stmt.StmtIface.ExecContext(ctx, args)
	err = processSQLiteError(err)
	return
}

func (stmt *SQLiteStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), toNamedValues(args))
}

func (stmt *SQLiteStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	rows, err = stmt.StmtIface.QueryContext(ctx, args)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &SQLiteRows{
		Ctx:       ctx,
		RowsIface: rows.(sqlite.RowsIface),
	}, nil
}

func (stmt *SQLiteStmt) Close() (err error) {
	err = stmt.StmtIface.Close()
	err = processSQLiteError(err)
	return
}

type SQLiteRows struct {
	Ctx context.Context
	sqlite.RowsIface
}

func (r *SQLiteRows) Next(dest []driver.Value) (err error) {
	err = r.RowsIface.Next(dest)
	err = processSQLiteError(err)
	return
}

func (r *SQLiteRows) Close() (err error) {
	// use background ctx as these rows MUST be closed.
	err = r.RowsIface.Close()
	err = processSQLiteError(err)
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

type rows interface {
	driver.Rows
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
