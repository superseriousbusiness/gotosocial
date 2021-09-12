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

package processing_test

import (
	"context"

	"git.iim.gay/grufwub/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ProcessingStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config              *config.Config
	db                  db.DB
	log                 *logrus.Logger
	storage             *kv.KVStore
	typeconverter       typeutils.TypeConverter
	transportController transport.Controller
	federator           federation.Federator
	oauthServer         oauth.Server
	mediaHandler        media.Handler
	timelineManager     timeline.Manager

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
}

func (suite *ProcessingStandardTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.typeconverter = testrig.NewTestTypeConverter(suite.db)
	suite.transportController = testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	suite.federator = testrig.NewTestFederator(suite.db, suite.transportController, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.mediaHandler = testrig.NewTestMediaHandler(suite.db, suite.storage)
	suite.timelineManager = testrig.NewTestTimelineManager(suite.db)

	suite.processor = processing.NewProcessor(
		suite.config,
		suite.typeconverter,
		suite.federator,
		suite.oauthServer,
		suite.mediaHandler,
		suite.storage,
		suite.timelineManager,
		suite.db,
		suite.log)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
	if err := suite.processor.Start(context.Background()); err != nil {
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
