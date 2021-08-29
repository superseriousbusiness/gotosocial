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

package bundb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *StatusTestSuite) SetupSuite() {
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

func (suite *StatusTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *StatusTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *StatusTestSuite) TestGetStatusByID() {
	status, err := suite.db.GetStatusByID(context.Background(), suite.testStatuses["local_account_1_status_1"].ID)
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

func (suite *StatusTestSuite) TestGetStatusByURI() {
	status, err := suite.db.GetStatusByURI(context.Background(), suite.testStatuses["local_account_1_status_1"].URI)
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

func (suite *StatusTestSuite) TestGetStatusWithExtras() {
	status, err := suite.db.GetStatusByID(context.Background(), suite.testStatuses["admin_account_status_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.NotEmpty(status.Tags)
	suite.NotEmpty(status.Attachments)
	suite.NotEmpty(status.Emojis)
}

func (suite *StatusTestSuite) TestGetStatusWithMention() {
	status, err := suite.db.GetStatusByID(context.Background(), suite.testStatuses["local_account_2_status_5"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.NotEmpty(status.Mentions)
	suite.NotEmpty(status.MentionIDs)
	suite.NotNil(status.InReplyTo)
	suite.NotNil(status.InReplyToAccount)
}

func (suite *StatusTestSuite) TestGetStatusTwice() {
	before1 := time.Now()
	_, err := suite.db.GetStatusByURI(context.Background(), suite.testStatuses["local_account_1_status_1"].URI)
	suite.NoError(err)
	after1 := time.Now()
	duration1 := after1.Sub(before1)
	fmt.Println(duration1.Milliseconds())

	before2 := time.Now()
	_, err = suite.db.GetStatusByURI(context.Background(), suite.testStatuses["local_account_1_status_1"].URI)
	suite.NoError(err)
	after2 := time.Now()
	duration2 := after2.Sub(before2)
	fmt.Println(duration2.Milliseconds())

	// second retrieval should be several orders faster since it will be cached now
	suite.Less(duration2, duration1)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
