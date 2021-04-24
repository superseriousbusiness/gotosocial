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

package federation_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ProtocolTestSuite struct {
	suite.Suite
	config    *config.Config
	db        db.DB
	log       *logrus.Logger
	federator *federation.Federator
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *ProtocolTestSuite) SetupSuite() {
	// setup standard items
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()

	// setup module being tested
	suite.federator = federation.NewFederator(suite.db, suite.log, suite.config).(*federation.Federator)
}

func (suite *ProtocolTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *ProtocolTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *ProtocolTestSuite) TestPostInboxRequestBodyHook() {

	// setup
	// recorder := httptest.NewRecorder()
	// ctx := context.Background()
	// request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/the_mighty_zork/inbox", nil) // the endpoint we're hitting

	// activity :=

	// _, err := suite.federator.PostInboxRequestBodyHook(ctx, request, nil)
	// assert.NoError(suite.T(), err)

	// check response
	// suite.EqualValues(http.StatusOK, recorder.Code)

	// result := recorder.Result()
	// defer result.Body.Close()
	// b, err := ioutil.ReadAll(result.Body)
	// assert.NoError(suite.T(), err)

}

func TestProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolTestSuite))
}
