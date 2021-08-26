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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type SessionTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *SessionTestSuite) SetupSuite() {
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

func (suite *SessionTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *SessionTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *SessionTestSuite) TestGetSession() {
	session, err := suite.db.GetSession(context.Background())
	suite.NoError(err)
	suite.NotNil(session)
	suite.NotEmpty(session.Auth)
	suite.NotEmpty(session.Crypt)
	suite.NotEmpty(session.ID)

	// TODO -- the same session should be returned with consecutive selects
	// right now there's an issue with bytea in bun, so uncomment this when that issue is fixed: https://github.com/uptrace/bun/issues/122
	// session2, err := suite.db.GetSession(context.Background())
	// suite.NoError(err)
	// suite.NotNil(session2)
	// suite.Equal(session.Auth, session2.Auth)
	// suite.Equal(session.Crypt, session2.Crypt)
	// suite.Equal(session.ID, session2.ID)
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}
