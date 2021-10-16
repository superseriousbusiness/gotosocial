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

package federatingdb_test

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatingDBTestSuite struct {
	suite.Suite
	config       *config.Config
	db           db.DB
	tc           typeutils.TypeConverter
	federatingDB federatingdb.DB

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
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
}

func (suite *FederatingDBTestSuite) SetupTest() {
	testrig.InitTestLog()
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.federatingDB = testrig.NewTestFederatingDB(suite.db)
	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *FederatingDBTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func createTestContext(receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account, fromFederatorChan chan messages.FromFederator) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, util.APReceivingAccount, receivingAccount)
	ctx = context.WithValue(ctx, util.APRequestingAccount, requestingAccount)
	ctx = context.WithValue(ctx, util.APFromFederatorChanKey, fromFederatorChan)
	return ctx
}
