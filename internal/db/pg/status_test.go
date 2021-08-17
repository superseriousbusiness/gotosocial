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

package pg_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusTestSuite struct {
	PGStandardTestSuite
}

func (suite *PGStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
}

func (suite *PGStandardTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *PGStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *PGStandardTestSuite) TestGetStatusByID() {
	status, err := suite.db.GetStatusByID(suite.testStatuses["local_account_1_status_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.Nil(status.BoostOf)
	suite.Nil(status.BoostOfAccount)
	suite.Nil(status.InReplyTo)
	suite.Nil(status.InReplyToAccount)
}

func (suite *PGStandardTestSuite) TestGetStatusByURI() {
	status, err := suite.db.GetStatusByURI(suite.testStatuses["local_account_1_status_1"].URI)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.Nil(status.BoostOf)
	suite.Nil(status.BoostOfAccount)
	suite.Nil(status.InReplyTo)
	suite.Nil(status.InReplyToAccount)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(PGStandardTestSuite))
}
