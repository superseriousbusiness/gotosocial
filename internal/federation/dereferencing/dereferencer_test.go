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

package dereferencing_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type DereferencerStandardTestSuite struct {
	suite.Suite
	db      db.DB
	storage *storage.Driver
	state   state.State
	client  *testrig.MockHTTPClient

	testRemoteStatuses    map[string]vocab.ActivityStreamsNote
	testRemotePeople      map[string]vocab.ActivityStreamsPerson
	testRemoteGroups      map[string]vocab.ActivityStreamsGroup
	testRemoteServices    map[string]vocab.ActivityStreamsService
	testRemoteAttachments map[string]testrig.RemoteAttachmentFile
	testAccounts          map[string]*gtsmodel.Account
	testEmojis            map[string]*gtsmodel.Emoji

	dereferencer dereferencing.Dereferencer
}

func (suite *DereferencerStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.testAccounts = testrig.NewTestAccounts()
	suite.testRemoteStatuses = testrig.NewTestFediStatuses()
	suite.testRemotePeople = testrig.NewTestFediPeople()
	suite.testRemoteGroups = testrig.NewTestFediGroups()
	suite.testRemoteServices = testrig.NewTestFediServices()
	suite.testRemoteAttachments = testrig.NewTestFediAttachments("../../../testrig/media")
	suite.testEmojis = testrig.NewTestEmojis()

	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	suite.db = testrig.NewTestDB(&suite.state)

	converter := typeutils.NewConverter(&suite.state)

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		converter,
	)

	suite.client = testrig.NewMockHTTPClient(nil, "../../../testrig/media")
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.state.Storage = suite.storage

	visFilter := visibility.NewFilter(&suite.state)
	intFilter := interaction.NewFilter(&suite.state)
	media := testrig.NewTestMediaManager(&suite.state)
	suite.dereferencer = dereferencing.NewDereferencer(
		&suite.state,
		converter,
		testrig.NewTestTransportController(
			&suite.state,
			suite.client,
		),
		visFilter,
		intFilter,
		media,
	)
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *DereferencerStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StopWorkers(&suite.state)
}
