// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package stream_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/stream"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StreamTestSuite struct {
	suite.Suite
	testAccounts map[string]*gtsmodel.Account
	testStatuses map[string]*gtsmodel.Status
	testTokens   map[string]*gtsmodel.Token
	db           db.DB
	oauthServer  oauth.Server
	state        state.State

	streamProcessor stream.Processor
}

func (suite *StreamTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestLog()
	testrig.InitTestConfig()

	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTokens = testrig.NewTestTokens()
	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.oauthServer = testrig.NewTestOauthServer(&suite.state)
	suite.streamProcessor = stream.New(&suite.state, suite.oauthServer)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *StreamTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}
