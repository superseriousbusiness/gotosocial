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

package account_test

import (
	"git.iim.gay/grufwub/go-store/kv"
	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config              *config.Config
	db                  db.DB
	log                 *logrus.Logger
	tc                  typeutils.TypeConverter
	storage             *kv.KVStore
	mediaHandler        media.Handler
	oauthServer         oauth.Server
	fromClientAPIChan   chan messages.FromClientAPI
	httpClient          pub.HttpClient
	transportController transport.Controller
	federator           federation.Federator

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status

	// module being tested
	accountProcessor account.Processor
}

func (suite *AccountStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *AccountStandardTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.storage = testrig.NewTestStorage()
	suite.mediaHandler = testrig.NewTestMediaHandler(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.fromClientAPIChan = make(chan messages.FromClientAPI, 100)
	suite.httpClient = testrig.NewMockHTTPClient(nil)
	suite.transportController = testrig.NewTestTransportController(suite.httpClient, suite.db)
	suite.federator = testrig.NewTestFederator(suite.db, suite.transportController, suite.storage)
	suite.accountProcessor = account.New(suite.db, suite.tc, suite.mediaHandler, suite.oauthServer, suite.fromClientAPIChan, suite.federator, suite.config, suite.log)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *AccountStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}
