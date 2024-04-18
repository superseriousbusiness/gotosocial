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

package sqlite

import (
	"context"
	"database/sql/driver"

	// linkname shenanigans
	_ "unsafe"

	// the library being unsafely linked to
	_ "github.com/ncruces/go-sqlite3/driver"

	// embed wasm sqlite binary
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/ncruces/go-sqlite3"
)

// ConnIface is the driver.Conn interface
// types (and the like) that go-sqlite3/driver.conn
// conforms to. Useful so you don't need
// to repeatedly perform checks yourself.
type ConnIface interface {
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.ExecerContext
}

// StmtIface is the driver.Stmt interface
// types (and the like) that go-sqlite3/driver.stmt
// conforms to. Useful so you don't need
// to repeatedly perform checks yourself.
type StmtIface interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

// RowsIface is the driver.Rows interface
// types (and the like) that go-sqlite3/driver.rows
// conforms to. Useful so you don't need
// to repeatedly perform checks yourself.
type RowsIface interface {
	driver.Rows
	driver.RowsColumnTypeDatabaseTypeName
}

type Driver struct {
	Init func(*sqlite3.Conn) error
}

// Open: implements database/sql/driver.Driver.
func (d Driver) Open(name string) (driver.Conn, error) {
	c, err := d.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

// OpenConnector: implements database/sql/driver.DriverContext.
func (d Driver) OpenConnector(name string) (driver.Connector, error) {
	return newConnector(name, d.Init)
}

type connector struct {
	init    func(*sqlite3.Conn) error
	name    string
	txBegin string
	tmRead  sqlite3.TimeFormat
	tmWrite sqlite3.TimeFormat
	pragmas bool
}

func (c *connector) Driver() driver.Driver { return Driver{Init: c.init} }

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	return connect(c, ctx)
}

//go:linkname newConnector github.com/ncruces/go-sqlite3/driver.newConnector
func newConnector(name string, init func(*sqlite3.Conn) error) (*connector, error)

//go:linkname connect github.com/ncruces/go-sqlite3/driver.(*connector).Connect
func connect(n *connector, ctx context.Context) (driver.Conn, error) //nolint
