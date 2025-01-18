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

package status_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/processing/interactionrequests"
	"github.com/superseriousbusiness/gotosocial/internal/processing/polls"
	"github.com/superseriousbusiness/gotosocial/internal/processing/status"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusStandardTestSuite struct {
	suite.Suite
	db            db.DB
	typeConverter *typeutils.Converter
	tc            transport.Controller
	storage       *storage.Driver
	state         state.State
	mediaManager  *media.Manager
	federator     *federation.Federator

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
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.typeConverter = typeutils.NewConverter(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)

	suite.tc = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../testrig/media"))
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, suite.tc, suite.mediaManager)

	visFilter := visibility.NewFilter(&suite.state)
	intFilter := interaction.NewFilter(&suite.state)
	testrig.StartTimelines(
		&suite.state,
		visFilter,
		suite.typeConverter,
	)

	common := common.New(&suite.state, suite.mediaManager, suite.typeConverter, suite.federator, visFilter)
	polls := polls.New(&common, &suite.state, suite.typeConverter)
	intReqs := interactionrequests.New(&common, &suite.state, suite.typeConverter)

	suite.status = status.New(
		&suite.state,
		&common,
		&polls,
		&intReqs,
		suite.federator,
		suite.typeConverter,
		visFilter,
		intFilter,
		processing.GetParseMentionFunc(
			&suite.state,
			suite.federator,
		),
	)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *StatusStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}
