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
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaStandardTestSuite struct {
	suite.Suite

	db                  db.DB
	storage             *storage.Driver
	state               state.State
	manager             *media.Manager
	transportController transport.Controller
	testAttachments     map[string]*gtsmodel.MediaAttachment
	testAccounts        map[string]*gtsmodel.Account
	testEmojis          map[string]*gtsmodel.Emoji
}

func (suite *MediaStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	suite.db = testrig.NewTestDB(&suite.state)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.state.Storage = suite.storage

	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
	testrig.StandardDBSetup(suite.db, nil)

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		typeutils.NewConverter(&suite.state),
	)

	suite.testAttachments = testrig.NewTestAttachments()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testEmojis = testrig.NewTestEmojis()
	suite.manager = testrig.NewTestMediaManager(&suite.state)
	suite.transportController = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../testrig/media"))
}

func (suite *MediaStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}
