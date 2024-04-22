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

	pgx "github.com/jackc/pgx/v5/stdlib"

	// our sqlite driver wrapper.
	"github.com/superseriousbusiness/gotosocial/internal/db/sqlite"
)

var (
	// global wrapped gts driver instances.
	gtsPostgresDriver = &PostgreSQLDriver{}

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
	sql.Register("sqlite-gts", &sqlite.Driver{})
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
