package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

var testModels []interface{} = []interface{}{
	&model.Account{},
	&model.Application{},
	&model.Block{},
	&model.DomainBlock{},
	&model.EmailDomainBlock{},
	&model.Follow{},
	&model.FollowRequest{},
	&model.MediaAttachment{},
	&model.Mention{},
	&model.Status{},
	&model.Tag{},
	&model.User{},
	&oauth.Token{},
	&oauth.Client{},
}

var TestAccounts map[string]*model.Account = map[string]*model.Account{

	"test_account_1": {
		ID: "",
	},
}

// StandardDBSetup populates a given db with all the necessary tables/models for perfoming tests.
func StandardDBSetup(db db.DB) error {
	for _, m := range testModels {
		if err := db.CreateTable(m); err != nil {
			return err
		}
	}
	return nil
}

// StandardDBTeardown drops all the standard testing tables/models from the database to ensure it's clean for the next test.
func StandardDBTeardown(db db.DB) error {
	for _, m := range testModels {
		if err := db.DropTable(m); err != nil {
			return err
		}
	}
	return nil
}
