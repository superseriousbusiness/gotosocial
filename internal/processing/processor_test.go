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

package processing_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ProcessingStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	storage             *storage.Driver
	mediaManager        media.Manager
	typeconverter       typeutils.TypeConverter
	httpClient          *testrig.MockHTTPClient
	transportController transport.Controller
	federator           federation.Federator
	oauthServer         oauth.Server
	emailSender         email.Sender

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testTags         map[string]*gtsmodel.Tag
	testMentions     map[string]*gtsmodel.Mention
	testAutheds      map[string]*oauth.Auth
	testBlocks       map[string]*gtsmodel.Block
	testActivities   map[string]testrig.ActivityWithSignature

	processor processing.Processor
}

func (suite *ProcessingStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
	suite.testAutheds = map[string]*oauth.Auth{
		"local_account_1": {
			Application: suite.testApplications["local_account_1"],
			User:        suite.testUsers["local_account_1"],
			Account:     suite.testAccounts["local_account_1"],
		},
	}
	suite.testBlocks = testrig.NewTestBlocks()
}

func (suite *ProcessingStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB()
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	suite.storage = testrig.NewInMemoryStorage()
	suite.typeconverter = testrig.NewTestTypeConverter(suite.db)
	suite.httpClient = testrig.NewMockHTTPClient(nil, "../../testrig/media")

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	suite.transportController = testrig.NewTestTransportController(suite.httpClient, suite.db, fedWorker)
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.federator = testrig.NewTestFederator(suite.db, suite.transportController, suite.storage, suite.mediaManager, fedWorker)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.emailSender = testrig.NewEmailSender("../../web/template/", nil)

	suite.processor = processing.NewProcessor(suite.typeconverter, suite.federator, suite.oauthServer, suite.mediaManager, suite.storage, suite.db, suite.emailSender, clientWorker, fedWorker)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
	if err := suite.processor.Start(); err != nil {
		panic(err)
	}
}

func (suite *ProcessingStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	if err := suite.processor.Stop(); err != nil {
		panic(err)
	}
}
