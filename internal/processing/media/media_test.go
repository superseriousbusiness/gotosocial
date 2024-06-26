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

package media_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	mediaprocessing "github.com/superseriousbusiness/gotosocial/internal/processing/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	tc                  *typeutils.Converter
	storage             *storage.Driver
	state               state.State
	mediaManager        *media.Manager
	transportController transport.Controller

	// standard suite models
	testTokens            map[string]*gtsmodel.Token
	testClients           map[string]*gtsmodel.Client
	testApplications      map[string]*gtsmodel.Application
	testUsers             map[string]*gtsmodel.User
	testAccounts          map[string]*gtsmodel.Account
	testAttachments       map[string]*gtsmodel.MediaAttachment
	testStatuses          map[string]*gtsmodel.Status
	testRemoteAttachments map[string]testrig.RemoteAttachmentFile

	// module being tested
	mediaProcessor mediaprocessing.Processor
}

func (suite *MediaStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testRemoteAttachments = testrig.NewTestFediAttachments("../../../testrig/media")
}

func (suite *MediaStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.tc = typeutils.NewConverter(&suite.state)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.transportController = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../testrig/media"))

	federator := testrig.NewTestFederator(&suite.state, suite.transportController, suite.mediaManager)
	filter := visibility.NewFilter(&suite.state)
	common := common.New(&suite.state, suite.mediaManager, suite.tc, federator, filter)

	suite.mediaProcessor = mediaprocessing.New(&common, &suite.state, suite.tc, federator, suite.mediaManager, suite.transportController)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *MediaStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}
