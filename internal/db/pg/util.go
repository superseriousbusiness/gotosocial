package pg

import (
	"strings"

	"database/sql"

	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// processErrorResponse parses the given error and returns an appropriate DBError.
func processErrorResponse(err error) db.Error {
	switch err {
	case nil:
		return nil
	case sql.ErrNoRows:
		return db.ErrNoEntries
	default:
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return db.ErrAlreadyExists
		}
		return err
	}
}
