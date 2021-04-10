package testrig

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

var testModels []interface{} = []interface{}{
	&gtsmodel.Account{},
	&gtsmodel.Application{},
	&gtsmodel.Block{},
	&gtsmodel.DomainBlock{},
	&gtsmodel.EmailDomainBlock{},
	&gtsmodel.Follow{},
	&gtsmodel.FollowRequest{},
	&gtsmodel.MediaAttachment{},
	&gtsmodel.Mention{},
	&gtsmodel.Status{},
	&gtsmodel.Tag{},
	&gtsmodel.User{},
	&oauth.Token{},
	&oauth.Client{},
}

// NewTestDB returns a new initialized, empty database for testing
func NewTestDB() db.DB {
	config := NewTestConfig()
	l := logrus.New()
	l.SetLevel(logrus.TraceLevel)
	testDB, err := db.New(context.Background(), config, l)
	if err != nil {
		panic(err)
	}
	return testDB
}

// StandardDBSetup populates a given db with all the necessary tables/models for perfoming tests.
func StandardDBSetup(db db.DB) {
	for _, m := range testModels {
		if err := db.CreateTable(m); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestTokens() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestClients() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestApplications() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestUsers() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestAccounts() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestAttachments() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestStatuses() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}
}

// StandardDBTeardown drops all the standard testing tables/models from the database to ensure it's clean for the next test.
func StandardDBTeardown(db db.DB) {
	for _, m := range testModels {
		if err := db.DropTable(m); err != nil {
			panic(err)
		}
	}
}
