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

package favourites_test

import (
	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/favourites"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type FavouritesStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	tc           *typeutils.Converter
	mediaManager *media.Manager
	federator    *federation.Federator
	emailSender  email.Sender
	processor    *processing.Processor
	storage      *storage.Driver
	state        state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testFollows      map[string]*gtsmodel.Follow

	// module being tested
	favModule *favourites.Module
}

func (suite *FavouritesStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFollows = testrig.NewTestFollows()
}

func (suite *FavouritesStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	suite.tc = typeutils.NewConverter(&suite.state)

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../../testrig/media")), suite.mediaManager)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.favModule = favourites.New(suite.processor)
}

func (suite *FavouritesStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *FavouritesStandardTestSuite) TestProcessFave() {}
