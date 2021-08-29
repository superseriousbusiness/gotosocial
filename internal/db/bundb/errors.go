package bundb

import (
	"github.com/jackc/pgconn"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// processPostgresError processes an error, replacing any postgres specific errors with our own error type
func processPostgresError(err error) db.Error {
	// Attempt to cast as postgres
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return err
	}

	// Handle supplied error code:
	// (https://www.postgresql.org/docs/10/errcodes-appendix.html)
	switch pgErr.Code {
	case "23505" /* unique_violation */ :
		return db.ErrAlreadyExists
	default:
		return err
	}
}

// processSQLiteError processes an error, replacing any sqlite specific errors with our own error type
func processSQLiteError(err error) db.Error {
	// Attempt to cast as sqlite
	sqliteErr, ok := err.(*sqlite.Error)
	if !ok {
		return err
	}

	// Handle supplied error code:
	switch sqliteErr.Code() {
	case sqlite3.SQLITE_CONSTRAINT_UNIQUE:
		return db.ErrAlreadyExists
	default:
		return err
	}
}
