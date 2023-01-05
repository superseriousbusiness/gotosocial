/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
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

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *PgClientStoreTestSuite) SetupSuite() {
	suite.testClientID = "01FCVB74EW6YBYAEY7QG9CQQF6"
	suite.testClientSecret = "4cc87402-259b-4a35-9485-2c8bf54f3763"
	suite.testClientDomain = "https://example.org"
	suite.testClientUserID = "01FEGYXKVCDB731QF9MVFXA4F5"
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *PgClientStoreTestSuite) SetupTest() {
	testrig.InitTestLog()
	testrig.InitTestConfig()
	suite.db = testrig.NewTestDB()
	testrig.StandardDBSetup(suite.db, nil)
}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *PgClientStoreTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
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
	suite.Assert().EqualValues(db.ErrNoEntries, err)
}

func TestPgClientStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PgClientStoreTestSuite))
}
