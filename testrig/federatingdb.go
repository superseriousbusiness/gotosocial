package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// NewTestFederatingDB returns a federating DB with the underlying db
func NewTestFederatingDB(db db.DB, fedWorker *concurrency.WorkerPool[messages.FromFederator]) federatingdb.DB {
	return federatingdb.New(db, fedWorker)
}
