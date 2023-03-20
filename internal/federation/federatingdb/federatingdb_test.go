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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatingDBTestSuite struct {
	suite.Suite
	db            db.DB
	tc            typeutils.TypeConverter
	fromFederator chan messages.FromFederator
	federatingDB  federatingdb.DB
	state         state.State

	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testBlocks       map[string]*gtsmodel.Block
	testActivities   map[string]testrig.ActivityWithSignature
}

func (suite *FederatingDBTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
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
	testrig.StartWorkers(&suite.state)

	suite.fromFederator = make(chan messages.FromFederator, 10)
	suite.state.Workers.EnqueueFederator = func(ctx context.Context, msgs ...messages.FromFederator) {
		for _, msg := range msgs {
			suite.fromFederator <- msg
		}
	}

	suite.db = testrig.NewTestDB(&suite.state)
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.federatingDB = testrig.NewTestFederatingDB(&suite.state)
	testrig.StandardDBSetup(suite.db, suite.testAccounts)

	suite.state.DB = suite.db
}

func (suite *FederatingDBTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StopWorkers(&suite.state)
	for suite.fromFederator != nil {
		select {
		case <-suite.fromFederator:
		default:
			return
		}
	}
}

func createTestContext(receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ap.ContextReceivingAccount, receivingAccount)
	ctx = context.WithValue(ctx, ap.ContextRequestingAccount, requestingAccount)
	return ctx
}
