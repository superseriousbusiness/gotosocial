package bundb

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// dbConn wrapps a bun.DB conn to provide SQL-type specific additional functionality
type DBConn struct {
	errProc func(error) db.Error // errProc is the SQL-type specific error processor
	log     *logrus.Logger       // log is the logger passed with this DBConn
	*bun.DB                      // DB is the underlying bun.DB connection
}

// WrapDBConn @TODO
func WrapDBConn(dbConn *bun.DB, log *logrus.Logger) *DBConn {
	var errProc func(error) db.Error
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
		log:     log,
		DB:      dbConn,
	}
}

// ProcessError processes an error to replace any known values with our own db.Error types,
// making it easier to catch specific situations (e.g. no rows, already exists, etc)
func (conn *DBConn) ProcessError(err error) db.Error {
	switch {
	case err == nil:
		return nil
	case err == sql.ErrNoRows:
		return db.ErrNoEntries
	default:
		return conn.errProc(err)
	}
}

// Exists checks the results of a SelectQuery for the existence of the data in question, masking ErrNoEntries errors
func (conn *DBConn) Exists(ctx context.Context, query *bun.SelectQuery) (bool, db.Error) {
	// Get the select query result
	count, err := query.Count(ctx)

	// Process error as our own and check if it exists
	switch err := conn.ProcessError(err); err {
	case nil:
		return (count != 0), nil
	case db.ErrNoEntries:
		return false, nil
	default:
		return false, err
	}
}

// NotExists is the functional opposite of conn.Exists()
func (conn *DBConn) NotExists(ctx context.Context, query *bun.SelectQuery) (bool, db.Error) {
	// Simply inverse of conn.exists()
	exists, err := conn.Exists(ctx, query)
	return !exists, err
}
