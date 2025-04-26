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

package postgres

import (
	"context"
	"database/sql/driver"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	pgx "github.com/jackc/pgx/v5/stdlib"
)

var (
	// global PostgreSQL driver instances.
	postgresDriver = pgx.GetDefaultDriver().(*pgx.Driver)

	// check the postgres driver types
	// conforms to our interface types.
	// (note SQLite doesn't export their
	// driver types, and gets checked in
	// tests very regularly anywho).
	_ connIface = (*pgx.Conn)(nil)
	_ stmtIface = (*pgx.Stmt)(nil)
	_ rowsIface = (*pgx.Rows)(nil)
)

// Driver is our own wrapper around the
// pgx/stdlib.Driver{} type in order to wrap further
// SQL driver types with our own err processing.
type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) {
	conn, err := postgresDriver.Open(name)
	if err != nil {
		err = processPostgresError(err)
		return nil, err
	}
	return &postgresConn{conn.(connIface)}, nil
}

func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	cc, err := postgresDriver.OpenConnector(name)
	if err != nil {
		err = processPostgresError(err)
		return nil, err
	}
	return &postgresConnector{driver: d, Connector: cc}, nil
}

type postgresConnector struct {
	driver *Driver
	driver.Connector
}

func (c *postgresConnector) Driver() driver.Driver { return c.driver }

func (c *postgresConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		err = processPostgresError(err)
		return nil, err
	}
	return &postgresConn{conn.(connIface)}, nil
}

type postgresConn struct{ connIface }

func (c *postgresConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *postgresConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	tx, err := c.connIface.BeginTx(ctx, opts)
	err = processPostgresError(err)
	if err != nil {
		return nil, err
	}
	return &postgresTx{tx}, nil
}

func (c *postgresConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *postgresConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	st, err := c.connIface.PrepareContext(ctx, query)
	err = processPostgresError(err)
	if err != nil {
		return nil, err
	}
	return &postgresStmt{st.(stmtIface)}, nil
}

func (c *postgresConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, db.ToNamedValues(args))
}

func (c *postgresConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	result, err := c.connIface.ExecContext(ctx, query, args)
	err = processPostgresError(err)
	return result, err
}

func (c *postgresConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.QueryContext(context.Background(), query, db.ToNamedValues(args))
}

func (c *postgresConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows, err := c.connIface.QueryContext(ctx, query, args)
	err = processPostgresError(err)
	if err != nil {
		return nil, err
	}
	return &postgresRows{rows.(rowsIface)}, nil
}

func (c *postgresConn) Close() error {
	err := c.connIface.Close()
	return processPostgresError(err)
}

type postgresTx struct{ driver.Tx }

func (tx *postgresTx) Commit() error {
	err := tx.Tx.Commit()
	return processPostgresError(err)
}

func (tx *postgresTx) Rollback() error {
	err := tx.Tx.Rollback()
	return processPostgresError(err)
}

type postgresStmt struct{ stmtIface }

func (stmt *postgresStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), db.ToNamedValues(args))
}

func (stmt *postgresStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	res, err := stmt.stmtIface.ExecContext(ctx, args)
	err = processPostgresError(err)
	return res, err
}

func (stmt *postgresStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), db.ToNamedValues(args))
}

func (stmt *postgresStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, err := stmt.stmtIface.QueryContext(ctx, args)
	err = processPostgresError(err)
	if err != nil {
		return nil, err
	}
	return &postgresRows{rows.(rowsIface)}, nil
}

type postgresRows struct{ rowsIface }

func (r *postgresRows) Next(dest []driver.Value) error {
	err := r.rowsIface.Next(dest)
	err = processPostgresError(err)
	return err
}

func (r *postgresRows) Close() error {
	err := r.rowsIface.Close()
	err = processPostgresError(err)
	return err
}

type connIface interface {
	driver.Conn
	driver.ConnPrepareContext
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnBeginTx
}

type stmtIface interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

type rowsIface interface {
	driver.Rows
	driver.RowsColumnTypeDatabaseTypeName
	driver.RowsColumnTypeLength
	driver.RowsColumnTypePrecisionScale
	driver.RowsColumnTypeScanType
	driver.RowsColumnTypeScanType
}
