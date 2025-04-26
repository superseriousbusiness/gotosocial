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

package admin_test

import (
	adminactions "code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/interaction"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/processing/admin"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/subscriptions"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AdminStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	tc                  *typeutils.Converter
	storage             *storage.Driver
	state               state.State
	mediaManager        *media.Manager
	oauthServer         oauth.Server
	fromClientAPIChan   chan messages.FromClientAPI
	transportController transport.Controller
	federator           *federation.Federator
	emailSender         email.Sender
	sentEmails          map[string]string
	processor           *processing.Processor

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testFollows      map[string]*gtsmodel.Follow
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testEmojis       map[string]*gtsmodel.Emoji

	// module being tested
	adminProcessor *admin.Processor
}

func (suite *AdminStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testFollows = testrig.NewTestFollows()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testEmojis = testrig.NewTestEmojis()
}

func (suite *AdminStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = adminactions.New(suite.state.DB, &suite.state.Workers)
	suite.tc = typeutils.NewConverter(&suite.state)

	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.oauthServer = testrig.NewTestOauthServer(&suite.state)

	suite.transportController = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../testrig/media"))
	suite.federator = testrig.NewTestFederator(&suite.state, suite.transportController, suite.mediaManager)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", suite.sentEmails)

	suite.processor = processing.NewProcessor(
		cleaner.New(&suite.state),
		subscriptions.New(&suite.state, suite.transportController, suite.tc),
		suite.tc,
		suite.federator,
		suite.oauthServer,
		suite.mediaManager,
		&suite.state,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		visibility.NewFilter(&suite.state),
		interaction.NewFilter(&suite.state),
	)

	testrig.StartWorkers(&suite.state, suite.processor.Workers())
	suite.adminProcessor = suite.processor.Admin()

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *AdminStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}
