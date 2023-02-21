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

package status_test

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/processing/status"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusStandardTestSuite struct {
	suite.Suite
	db            db.DB
	typeConverter typeutils.TypeConverter
	tc            transport.Controller
	storage       *storage.Driver
	mediaManager  media.Manager
	federator     federation.Federator
	clientWorker  *concurrency.WorkerPool[messages.FromClientAPI]

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

	// module being tested
	status status.Processor
}

func (suite *StatusStandardTestSuite) SetupSuite() {
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

func (suite *StatusStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	suite.db = testrig.NewTestDB()
	suite.typeConverter = testrig.NewTestTypeConverter(suite.db)
	suite.clientWorker = concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	suite.tc = testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../testrig/media"), suite.db, fedWorker)
	suite.storage = testrig.NewInMemoryStorage()
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.federator = testrig.NewTestFederator(suite.db, suite.tc, suite.storage, suite.mediaManager, fedWorker)
	suite.status = status.New(suite.db, suite.typeConverter, suite.clientWorker, processing.GetParseMentionFunc(suite.db, suite.federator))
	suite.clientWorker.SetProcessor(func(ctx context.Context, msg messages.FromClientAPI) error { return nil })
	suite.NoError(suite.clientWorker.Start())

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *StatusStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}
