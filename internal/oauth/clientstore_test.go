/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package oauth_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/pg"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/oauth2/v4/models"
)

type PgClientStoreTestSuite struct {
	suite.Suite
	db               db.DB
	testClientID     string
	testClientSecret string
	testClientDomain string
	testClientUserID string
}

const ()

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *PgClientStoreTestSuite) SetupSuite() {
	suite.testClientID = "test-client-id"
	suite.testClientSecret = "test-client-secret"
	suite.testClientDomain = "https://example.org"
	suite.testClientUserID = "test-client-user-id"
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *PgClientStoreTestSuite) SetupTest() {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	c := config.Empty()
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	db, err := pg.NewPostgresService(context.Background(), c, log)
	if err != nil {
		logrus.Panicf("error creating database connection: %s", err)
	}

	suite.db = db

	models := []interface{}{
		&oauth.Client{},
	}

	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}
}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *PgClientStoreTestSuite) TearDownTest() {
	models := []interface{}{
		&oauth.Client{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
	suite.db = nil
}

func (suite *PgClientStoreTestSuite) TestClientStoreSetAndGet() {
	// set a new client in the store
	cs := oauth.NewClientStore(suite.db)
	if err := cs.Set(context.Background(), suite.testClientID, models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID)); err != nil {
		suite.FailNow(err.Error())
	}

	// fetch that client from the store
	client, err := cs.GetByID(context.Background(), suite.testClientID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// check that the values are the same
	suite.NotNil(client)
	suite.EqualValues(models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID), client)
}

func (suite *PgClientStoreTestSuite) TestClientSetAndDelete() {
	// set a new client in the store
	cs := oauth.NewClientStore(suite.db)
	if err := cs.Set(context.Background(), suite.testClientID, models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID)); err != nil {
		suite.FailNow(err.Error())
	}

	// fetch the client from the store
	client, err := cs.GetByID(context.Background(), suite.testClientID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// check that the values are the same
	suite.NotNil(client)
	suite.EqualValues(models.New(suite.testClientID, suite.testClientSecret, suite.testClientDomain, suite.testClientUserID), client)
	if err := cs.Delete(context.Background(), suite.testClientID); err != nil {
		suite.FailNow(err.Error())
	}

	// try to get the deleted client; we should get an error
	deletedClient, err := cs.GetByID(context.Background(), suite.testClientID)
	suite.Assert().Nil(deletedClient)
	suite.Assert().EqualValues(db.ErrNoEntries{}, err)
}

func TestPgClientStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PgClientStoreTestSuite))
}
