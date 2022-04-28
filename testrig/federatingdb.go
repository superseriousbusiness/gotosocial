package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/worker"
)

// NewTestFederatingDB returns a federating DB with the underlying db
func NewTestFederatingDB(db db.DB, fedWorker *worker.Worker[messages.FromFederator]) federatingdb.DB {
	return federatingdb.New(db, fedWorker)
}
