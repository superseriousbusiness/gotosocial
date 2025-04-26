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

package fileserver_test

import (
	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/fileserver"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type FileserverTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	storage      *storage.Driver
	state        state.State
	federator    *federation.Federator
	tc           *typeutils.Converter
	processor    *processing.Processor
	mediaManager *media.Manager
	oauthServer  oauth.Server
	emailSender  email.Sender

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
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
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(
		&suite.state,
		testrig.NewTestTransportController(
			&suite.state,
			testrig.NewMockHTTPClient(nil, "../../../testrig/media"),
		),
		suite.mediaManager,
	)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)

	suite.tc = typeutils.NewConverter(&suite.state)

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.oauthServer = testrig.NewTestOauthServer(&suite.state)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", nil)

	suite.fileServer = fileserver.New(suite.processor)
}

func (suite *FileserverTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")

	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

func (suite *FileserverTestSuite) TearDownSuite() {
	if err := suite.db.Close(); err != nil {
		log.Panicf(nil, "error closing db connection: %s", err)
	}
	testrig.StopWorkers(&suite.state)
}

func (suite *FileserverTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}
