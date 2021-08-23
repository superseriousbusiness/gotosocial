package pg

import (
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// processErrorResponse parses the given error and returns an appropriate DBError.
func processErrorResponse(err error) db.Error {
	switch err {
	case nil:
		return nil
	case bun.ErrNoRows:
		return db.ErrNoEntries
	case bun.ErrMultiRows:
		return db.ErrMultipleEntries
	default:
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return db.ErrAlreadyExists
		}
		return err
	}
}
