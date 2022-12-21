/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package fileserver_test

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FileserverTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	storage      *storage.Driver
	federator    federation.Federator
	tc           typeutils.TypeConverter
	processor    processing.Processor
	mediaManager media.Manager
	oauthServer  oauth.Server
	emailSender  email.Sender

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment

	// item being tested
	fileServer *fileserver.Module
}

/*
	TEST INFRASTRUCTURE
*/

func (suite *FileserverTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)

	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewInMemoryStorage()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../testrig/media"), suite.db, fedWorker), suite.storage, suite.mediaManager, fedWorker)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", nil)

	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender, testrig.NewTestMediaManager(suite.db, suite.storage), clientWorker, fedWorker)
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)

	suite.fileServer = fileserver.New(suite.processor)
}

func (suite *FileserverTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

func (suite *FileserverTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		log.Panicf("error closing db connection: %s", err)
	}
}

func (suite *FileserverTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}
