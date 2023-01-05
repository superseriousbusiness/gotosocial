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

package streaming_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/streaming"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StreamingTestSuite struct {
	suite.Suite
	testAccounts map[string]*gtsmodel.Account
	testTokens   map[string]*gtsmodel.Token
	db           db.DB
	oauthServer  oauth.Server

	streamingProcessor streaming.Processor
}

func (suite *StreamingTestSuite) SetupTest() {
	testrig.InitTestLog()
	testrig.InitTestConfig()

	suite.testAccounts = testrig.NewTestAccounts()
	suite.testTokens = testrig.NewTestTokens()
	suite.db = testrig.NewTestDB()
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.streamingProcessor = streaming.New(suite.db, suite.oauthServer)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *StreamingTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}
