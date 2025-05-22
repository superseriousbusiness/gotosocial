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

package federation_test

import (
	"github.com/stretchr/testify/suite"

	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

type FederatorStandardTestSuite struct {
	suite.Suite
	storage             *storage.Driver
	state               state.State
	typeconverter       *typeutils.Converter
	transportController transport.Controller
	httpClient          *testrig.MockHTTPClient
	federator           *federation.Federator

	testAccounts   map[string]*gtsmodel.Account
	testStatuses   map[string]*gtsmodel.Status
	testActivities map[string]testrig.ActivityWithSignature
	testTombstones map[string]*gtsmodel.Tombstone
}

func (suite *FederatorStandardTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	suite.testTombstones = testrig.NewTestTombstones()
}

func (suite *FederatorStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.DB = testrig.NewTestDB(&suite.state)
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.typeconverter = typeutils.NewConverter(&suite.state)

	// Ensure it's possible to deref
	// main key of foss satan.
	fossSatanAS, err := suite.typeconverter.AccountToAS(suite.T().Context(), suite.testAccounts["remote_account_1"])
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.httpClient = testrig.NewMockHTTPClient(nil, "../../testrig/media", fossSatanAS)
	suite.httpClient.TestRemotePeople = testrig.NewTestFediPeople()
	suite.httpClient.TestRemoteStatuses = testrig.NewTestFediStatuses()

	suite.transportController = testrig.NewTestTransportController(&suite.state, suite.httpClient)
	suite.federator = testrig.NewTestFederator(&suite.state, suite.transportController, testrig.NewTestMediaManager(&suite.state))

	testrig.StandardDBSetup(suite.state.DB, nil)
	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
}

func (suite *FederatorStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}
