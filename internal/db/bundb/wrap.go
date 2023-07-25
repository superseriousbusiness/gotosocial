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

// WrappedDB wraps a bun database instance
// to provide common per-dialect SQL error
// conversions to common types, and retries
// on returned busy errors (SQLite only for now).
type WrappedDB struct {
	errHook func(error) error
	*bun.DB // underlying conn
}

// WrapDB wraps a bun database instance in our own WrappedDB type.
func WrapDB(db *bun.DB) *WrappedDB {
	var errProc func(error) error
	switch name := db.Dialect().Name(); name {
	case dialect.PG:
		errProc = processPostgresError
	case dialect.SQLite:
		errProc = processSQLiteError
	default:
		panic("unknown dialect name: " + name.String())
	}
	return &WrappedDB{
		errHook: errProc,
		DB:      db,
	}
}

// BeginTx wraps bun.DB.BeginTx() with retry-busy timeout.
func (db *WrappedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx bun.Tx, err error) {
	err = retryOnBusy(ctx, func() error {
		tx, err = db.DB.BeginTx(ctx, opts)
		err = db.ProcessError(err)
		return err
	})
	return
}

// ExecContext wraps bun.DB.ExecContext() with retry-busy timeout.
func (db *WrappedDB) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	err = retryOnBusy(ctx, func() error {
		result, err = db.DB.ExecContext(ctx, query, args...)
		err = db.ProcessError(err)
		return err
	})
	return
}

// QueryContext wraps bun.DB.QueryContext() with retry-busy timeout.
func (db *WrappedDB) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	err = retryOnBusy(ctx, func() error {
		rows, err = db.DB.QueryContext(ctx, query, args...)
		err = db.ProcessError(err)
		return err
	})
	return
}

// QueryRowContext wraps bun.DB.QueryRowContext() with retry-busy timeout.
func (db *WrappedDB) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	_ = retryOnBusy(ctx, func() error {
		row = db.DB.QueryRowContext(ctx, query, args...)
		err := db.ProcessError(row.Err())
		return err
	})
	return
}

// RunInTx is functionally the same as bun.DB.RunInTx() but with retry-busy timeouts.
func (db *WrappedDB) RunInTx(ctx context.Context, fn func(bun.Tx) error) error {
	// Attempt to start new transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			// Rollback (with retry-backoff).
			_ = retryOnBusy(ctx, func() error {
				err := tx.Rollback()
				return db.errHook(err)
			})
		}
	}()

	// Perform supplied transaction
	if err := fn(tx); err != nil {
		return db.errHook(err)
	}

	// Commit (with retry-backoff).
	err = retryOnBusy(ctx, func() error {
		err := tx.Commit()
		return db.errHook(err)
	})
	done = true
	return err
}

func (db *WrappedDB) NewValues(model interface{}) *bun.ValuesQuery {
	return bun.NewValuesQuery(db.DB, model).Conn(db)
}

func (db *WrappedDB) NewMerge() *bun.MergeQuery {
	return bun.NewMergeQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewSelect() *bun.SelectQuery {
	return bun.NewSelectQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewInsert() *bun.InsertQuery {
	return bun.NewInsertQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewUpdate() *bun.UpdateQuery {
	return bun.NewUpdateQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewDelete() *bun.DeleteQuery {
	return bun.NewDeleteQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewRaw(query string, args ...interface{}) *bun.RawQuery {
	return bun.NewRawQuery(db.DB, query, args...).Conn(db)
}

func (db *WrappedDB) NewCreateTable() *bun.CreateTableQuery {
	return bun.NewCreateTableQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewDropTable() *bun.DropTableQuery {
	return bun.NewDropTableQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewCreateIndex() *bun.CreateIndexQuery {
	return bun.NewCreateIndexQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewDropIndex() *bun.DropIndexQuery {
	return bun.NewDropIndexQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewTruncateTable() *bun.TruncateTableQuery {
	return bun.NewTruncateTableQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewAddColumn() *bun.AddColumnQuery {
	return bun.NewAddColumnQuery(db.DB).Conn(db)
}

func (db *WrappedDB) NewDropColumn() *bun.DropColumnQuery {
	return bun.NewDropColumnQuery(db.DB).Conn(db)
}

// ProcessError processes an error to replace any known values with our own error types,
// making it easier to catch specific situations (e.g. no rows, already exists, etc)
func (db *WrappedDB) ProcessError(err error) error {
	if err == nil {
		return nil
	}
	return db.errHook(err)
}

// Exists checks the results of a SelectQuery for the existence of the data in question, masking ErrNoEntries errors
func (db *WrappedDB) Exists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
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
func (db *WrappedDB) NotExists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
	exists, err := db.Exists(ctx, query)
	return !exists, err
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
