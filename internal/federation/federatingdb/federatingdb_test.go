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

package federatingdb_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatingDBTestSuite struct {
	suite.Suite
	db           db.DB
	tc           *typeutils.Converter
	federatingDB federatingdb.DB
	state        state.State

	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testBlocks       map[string]*gtsmodel.Block
	testActivities   map[string]testrig.ActivityWithSignature
}

func (suite *FederatingDBTestSuite) getFederatorMsg(timeout time.Duration) (*messages.FromFediAPI, bool) {
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, timeout)
	defer cncl()
	return suite.state.Workers.Federator.Queue.PopCtx(ctx)
}

func (suite *FederatingDBTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testBlocks = testrig.NewTestBlocks()
}

func (suite *FederatingDBTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	suite.db = testrig.NewTestDB(&suite.state)

	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	suite.tc = typeutils.NewConverter(&suite.state)

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		suite.tc,
	)

	suite.federatingDB = testrig.NewTestFederatingDB(&suite.state)
	testrig.StandardDBSetup(suite.db, suite.testAccounts)

	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
}

func (suite *FederatingDBTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StopWorkers(&suite.state)
}

func createTestContext(receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) context.Context {
	ctx := context.Background()
	ctx = gtscontext.SetReceivingAccount(ctx, receivingAccount)
	ctx = gtscontext.SetRequestingAccount(ctx, requestingAccount)
	return ctx
}
