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

package timeline_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type TimelineStandardTestSuite struct {
	suite.Suite
	db    db.DB
	state state.State

	// standard suite models
	testAccounts map[string]*gtsmodel.Account
	testStatuses map[string]*gtsmodel.Status

	// module being tested
	timeline timeline.Processor
}

func (suite *TimelineStandardTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *TimelineStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)

	suite.timeline = timeline.New(
		&suite.state,
		typeutils.NewConverter(&suite.state),
		visibility.NewFilter(&suite.state),
	)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *TimelineStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StopWorkers(&suite.state)
}
