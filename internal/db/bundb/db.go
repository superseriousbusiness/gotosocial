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
	"unsafe"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

// DB wraps a bun database instance
// to provide common per-dialect SQL error
// conversions to common types, and retries
// on returned busy (SQLite only).
type DB struct {
	// our own wrapped db type
	// with retry backoff support.
	// kept separate to the *bun.DB
	// type to be passed into query
	// builders as bun.IConn iface
	// (this prevents double firing
	// bun query hooks).
	//
	// also holds per-dialect
	// error hook function.
	raw rawdb

	// bun DB interface we use
	// for dialects, and improved
	// struct marshal/unmarshaling.
	bun *bun.DB
}

// WrapDB wraps a bun database instance in our database type.
func WrapDB(db *bun.DB) *DB {
	var errProc func(error) error
	switch name := db.Dialect().Name(); name {
	case dialect.PG:
		errProc = processPostgresError
	case dialect.SQLite:
		errProc = processSQLiteError
	default:
		panic("unknown dialect name: " + name.String())
	}
	return &DB{
		raw: rawdb{
			errHook: errProc,
			DB:      db.DB,
		},
		bun: db,
	}
}

// Dialect is a direct call-through to bun.DB.Dialect().
func (db *DB) Dialect() schema.Dialect { return db.bun.Dialect() }

// AddQueryHook is a direct call-through to bun.DB.AddQueryHook().
func (db *DB) AddQueryHook(hook bun.QueryHook) { db.bun.AddQueryHook(hook) }

// RegisterModels is a direct call-through to bun.DB.RegisterModels().
func (db *DB) RegisterModel(models ...any) { db.bun.RegisterModel(models...) }

// PingContext is a direct call-through to bun.DB.PingContext().
func (db *DB) PingContext(ctx context.Context) error { return db.bun.PingContext(ctx) }

// Close is a direct call-through to bun.DB.Close().
func (db *DB) Close() error { return db.bun.Close() }

// ExecContext wraps bun.DB.ExecContext() with retry-busy timeout and our own error processing.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	bundb := db.bun // use underlying *bun.DB interface for their query formatting
	err = retryOnBusy(ctx, func() error {
		result, err = bundb.ExecContext(ctx, query, args...)
		err = db.raw.errHook(err)
		return err
	})
	return
}

// QueryContext wraps bun.DB.ExecContext() with retry-busy timeout and our own error processing.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	bundb := db.bun // use underlying *bun.DB interface for their query formatting
	err = retryOnBusy(ctx, func() error {
		rows, err = bundb.QueryContext(ctx, query, args...)
		err = db.raw.errHook(err)
		return err
	})
	return
}

// QueryRowContext wraps bun.DB.ExecContext() with retry-busy timeout and our own error processing.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	bundb := db.bun // use underlying *bun.DB interface for their query formatting
	_ = retryOnBusy(ctx, func() error {
		row = bundb.QueryRowContext(ctx, query, args...)
		err := db.raw.errHook(row.Err())
		updateRowError(row, err) // set new error
		return err
	})
	return
}

// BeginTx wraps bun.DB.BeginTx() with retry-busy timeout and our own error processing.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx Tx, err error) {
	tx = Tx{db: db} // create wrapping tx type with our db ref
	bundb := db.bun // use *bun.DB interface to return bun.Tx type
	err = retryOnBusy(ctx, func() error {
		var btx bun.Tx
		btx, err = bundb.BeginTx(ctx, opts)
		err = db.raw.errHook(err)
		tx.tx = &btx // set bun tx ref
		return err
	})
	return
}

// RunInTx is functionally the same as bun.DB.RunInTx() but with retry-busy timeouts.
func (db *DB) RunInTx(ctx context.Context, fn func(Tx) error) error {
	// Attempt to start new transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			// Rollback (with retry-backoff).
			_ = retryOnBusy(ctx, tx.Rollback)
		}
	}()

	// Perform supplied transaction
	if err := fn(tx); err != nil {
		return db.raw.errHook(err)
	}

	// Commit (with retry-backoff).
	err = retryOnBusy(ctx, tx.Commit)
	done = true
	return err
}

func (db *DB) NewValues(model interface{}) *bun.ValuesQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewValuesQuery(db.bun, model).Conn(&db.raw)
}

func (db *DB) NewMerge() *bun.MergeQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewMergeQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewSelect() *bun.SelectQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewSelectQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewInsert() *bun.InsertQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewInsertQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewUpdate() *bun.UpdateQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewUpdateQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewDelete() *bun.DeleteQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewDeleteQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewRawQuery(db.bun, query, args...).Conn(&db.raw)
}

func (db *DB) NewCreateTable() *bun.CreateTableQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewCreateTableQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewDropTable() *bun.DropTableQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewDropTableQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewCreateIndex() *bun.CreateIndexQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewCreateIndexQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewDropIndex() *bun.DropIndexQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewDropIndexQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewTruncateTable() *bun.TruncateTableQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewTruncateTableQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewAddColumn() *bun.AddColumnQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewAddColumnQuery(db.bun).Conn(&db.raw)
}

func (db *DB) NewDropColumn() *bun.DropColumnQuery {
	// note: passing in rawdb as conn iface so no double query-hook
	// firing when passed through the bun.DB.Query___() functions.
	return bun.NewDropColumnQuery(db.bun).Conn(&db.raw)
}

// Exists checks the results of a SelectQuery for the existence of the data in question, masking ErrNoEntries errors.
func (db *DB) Exists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
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

// NotExists checks the results of a SelectQuery for the non-existence of the data in question, masking ErrNoEntries errors.
func (db *DB) NotExists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
	exists, err := db.Exists(ctx, query)
	return !exists, err
}

type rawdb struct {
	// dialect specific error
	// processing function hook.
	errHook func(error) error

	// embedded raw
	// db interface
	*sql.DB
}

// ExecContext wraps sql.DB.ExecContext() with retry-busy timeout.
func (db *rawdb) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	err = retryOnBusy(ctx, func() error {
		result, err = db.DB.ExecContext(ctx, query, args...)
		err = db.errHook(err)
		return err
	})
	return
}

// QueryContext wraps sql.DB.QueryContext() with retry-busy timeout.
func (db *rawdb) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	err = retryOnBusy(ctx, func() error {
		rows, err = db.DB.QueryContext(ctx, query, args...)
		err = db.errHook(err)
		return err
	})
	return
}

// QueryRowContext wraps sql.DB.QueryRowContext() with retry-busy timeout.
func (db *rawdb) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	_ = retryOnBusy(ctx, func() error {
		row = db.DB.QueryRowContext(ctx, query, args...)
		err := db.errHook(row.Err())
		return err
	})
	return
}

// Tx wraps a bun.Tx to add
// our own error processing
// through the rawdb.errHook.
type Tx struct {
	tx *bun.Tx
	db *DB
}

// ExecContext wraps bun.Tx.ExecContext() with our own error processing.
func (tx Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	res, err := tx.tx.ExecContext(ctx, query, args...)
	err = tx.db.raw.errHook(err)
	return res, err
}

// QueryContext wraps bun.Tx.QueryContext() with our own error processing.
func (tx Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := tx.tx.QueryContext(ctx, query, args...)
	err = tx.db.raw.errHook(err)
	return rows, err
}

// QueryRowContext wraps bun.Tx.QueryRowContext() with our own error processing.
func (tx Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row := tx.tx.QueryRowContext(ctx, query, args...)
	err := tx.db.raw.errHook(row.Err())
	updateRowError(row, err) // set new error
	return row
}

// Commit wraps bun.Tx.Commit() with our own error processing.
func (tx Tx) Commit() error {
	err := tx.tx.Commit()
	err = tx.db.raw.errHook(err)
	return err
}

// Rollback wraps bun.Tx.Rollback() with our own error processing.
func (tx Tx) Rollback() error {
	err := tx.tx.Rollback()
	err = tx.db.raw.errHook(err)
	return err
}

// Dialect is a direct call-through to bun.DB.Dialect().
func (tx Tx) Dialect() schema.Dialect {
	return tx.db.bun.Dialect()
}

func (tx Tx) NewValues(model interface{}) *bun.ValuesQuery {
	return bun.NewValuesQuery(tx.db.bun, model).Conn(tx)
}

func (tx Tx) NewMerge() *bun.MergeQuery {
	return bun.NewMergeQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewSelect() *bun.SelectQuery {
	return bun.NewSelectQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewInsert() *bun.InsertQuery {
	return bun.NewInsertQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewUpdate() *bun.UpdateQuery {
	return bun.NewUpdateQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewDelete() *bun.DeleteQuery {
	return bun.NewDeleteQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	return bun.NewRawQuery(tx.db.bun, query, args...).Conn(tx)
}

func (tx Tx) NewCreateTable() *bun.CreateTableQuery {
	return bun.NewCreateTableQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewDropTable() *bun.DropTableQuery {
	return bun.NewDropTableQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewCreateIndex() *bun.CreateIndexQuery {
	return bun.NewCreateIndexQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewDropIndex() *bun.DropIndexQuery {
	return bun.NewDropIndexQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewTruncateTable() *bun.TruncateTableQuery {
	return bun.NewTruncateTableQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewAddColumn() *bun.AddColumnQuery {
	return bun.NewAddColumnQuery(tx.db.bun).Conn(tx)
}

func (tx Tx) NewDropColumn() *bun.DropColumnQuery {
	return bun.NewDropColumnQuery(tx.db.bun).Conn(tx)
}

// updateRowError updates an sql.Row's internal error field using the unsafe package.
func updateRowError(sqlrow *sql.Row, err error) {
	type row struct {
		err  error
		rows *sql.Rows
	}
	(*row)(unsafe.Pointer(sqlrow)).err = err
}

// retryOnBusy will retry given function on returned 'errBusy'.
func retryOnBusy(ctx context.Context, fn func() error) error {
	var backoff time.Duration

	for i := 0; ; i++ {
		// Perform func.
		err := fn()

		if err != errBusy {
			// May be nil, or may be
			// some other error, either
			// way return here.
			return err
		}

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

		// Backoff for some time.
		case <-time.After(backoff):
		}
	}

	return gtserror.Newf("%w (waited > %s)", db.ErrBusyTimeout, backoff)
}
