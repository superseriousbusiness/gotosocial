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

//go:build !moderncsqlite3

package sqlite

import (
	"context"
	"database/sql/driver"

	"github.com/superseriousbusiness/gotosocial/internal/db"

	sqlite3driver "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"     // embed wasm binary
	_ "github.com/ncruces/go-sqlite3/vfs/memdb" // include memdb vfs
)

// Driver is our own wrapper around the
// driver.SQLite{} type in order to wrap
// further SQL types with our own
// functionality, e.g. err processing.
type Driver struct{ sqlite3driver.SQLite }

func (d *Driver) Open(name string) (driver.Conn, error) {
	conn, err := d.SQLite.Open(name)
	if err != nil {
		err = processSQLiteError(err)
		return nil, err
	}
	return &sqliteConn{conn.(connIface)}, nil
}

func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	cc, err := d.SQLite.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &sqliteConnector{driver: d, Connector: cc}, nil
}

type sqliteConnector struct {
	driver *Driver
	driver.Connector
}

func (c *sqliteConnector) Driver() driver.Driver { return c.driver }

func (c *sqliteConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &sqliteConn{conn.(connIface)}, nil
}

type sqliteConn struct{ connIface }

func (c *sqliteConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *sqliteConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	tx, err = c.connIface.BeginTx(ctx, opts)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &sqliteTx{tx}, nil
}

func (c *sqliteConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *sqliteConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	stmt, err = c.connIface.PrepareContext(ctx, query)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &sqliteStmt{stmtIface: stmt.(stmtIface)}, nil
}

func (c *sqliteConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.ExecContext(context.Background(), query, db.ToNamedValues(args))
}

func (c *sqliteConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	res, err = c.connIface.ExecContext(ctx, query, args)
	err = processSQLiteError(err)
	return
}

func (c *sqliteConn) Close() (err error) {
	// Get acces the underlying raw sqlite3 conn.
	raw := c.connIface.(sqlite3driver.Conn).Raw()

	// see: https://www.sqlite.org/pragma.html#pragma_optimize
	const onClose = "PRAGMA analysis_limit=1000; PRAGMA optimize;"
	_ = raw.Exec(onClose)

	// Finally, close.
	err = raw.Close()
	return
}

type sqliteTx struct{ driver.Tx }

func (tx *sqliteTx) Commit() (err error) {
	err = tx.Tx.Commit()
	err = processSQLiteError(err)
	return
}

func (tx *sqliteTx) Rollback() (err error) {
	err = tx.Tx.Rollback()
	err = processSQLiteError(err)
	return
}

type sqliteStmt struct{ stmtIface }

func (stmt *sqliteStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), db.ToNamedValues(args))
}

func (stmt *sqliteStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	res, err = stmt.stmtIface.ExecContext(ctx, args)
	err = processSQLiteError(err)
	return
}

func (stmt *sqliteStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), db.ToNamedValues(args))
}

func (stmt *sqliteStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	rows, err = stmt.stmtIface.QueryContext(ctx, args)
	err = processSQLiteError(err)
	if err != nil {
		return nil, err
	}
	return &sqliteRows{rows.(rowsIface)}, nil
}

func (stmt *sqliteStmt) Close() (err error) {
	err = stmt.stmtIface.Close()
	err = processSQLiteError(err)
	return
}

type sqliteRows struct{ rowsIface }

func (r *sqliteRows) Next(dest []driver.Value) (err error) {
	err = r.rowsIface.Next(dest)
	err = processSQLiteError(err)
	return
}

func (r *sqliteRows) Close() (err error) {
	err = r.rowsIface.Close()
	err = processSQLiteError(err)
	return
}

// connIface is the driver.Conn interface
// types (and the like) that go-sqlite3/driver.conn
// conforms to. Useful so you don't need
// to repeatedly perform checks yourself.
type connIface interface {
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.ExecerContext
}

// StmtIface is the driver.Stmt interface
// types (and the like) that go-sqlite3/driver.stmt
// conforms to. Useful so you don't need
// to repeatedly perform checks yourself.
type stmtIface interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

// RowsIface is the driver.Rows interface
// types (and the like) that go-sqlite3/driver.rows
// conforms to. Useful so you don't need
// to repeatedly perform checks yourself.
type rowsIface interface {
	driver.Rows
	driver.RowsColumnTypeDatabaseTypeName
}
