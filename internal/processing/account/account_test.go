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

package account_test

import (
	"context"

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
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	tc                  typeutils.TypeConverter
	storage             *storage.Driver
	mediaManager        media.Manager
	oauthServer         oauth.Server
	fromClientAPIChan   chan messages.FromClientAPI
	transportController transport.Controller
	federator           federation.Federator
	emailSender         email.Sender
	sentEmails          map[string]string

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
	testrig.InitTestLog()
	testrig.InitTestConfig()

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	clientWorker.SetProcessor(func(_ context.Context, msg messages.FromClientAPI) error {
		suite.fromClientAPIChan <- msg
		return nil
	})

	_ = fedWorker.Start()
	_ = clientWorker.Start()

	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.storage = testrig.NewInMemoryStorage()
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.fromClientAPIChan = make(chan messages.FromClientAPI, 100)
	suite.transportController = testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../testrig/media"), suite.db, fedWorker)
	suite.federator = testrig.NewTestFederator(suite.db, suite.transportController, suite.storage, suite.mediaManager, fedWorker)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", suite.sentEmails)
	suite.accountProcessor = account.New(suite.db, suite.tc, suite.mediaManager, suite.oauthServer, clientWorker, suite.federator, processing.GetParseMentionFunc(suite.db, suite.federator))
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *AccountStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}
