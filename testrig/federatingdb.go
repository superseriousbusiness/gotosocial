package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
)

// NewTestFederatingDB returns a federating DB with the underlying db
func NewTestFederatingDB(db db.DB) federatingdb.DB {
	return federatingdb.New(db, NewTestConfig())
}
