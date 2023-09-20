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
	"context"
	"sort"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type TimelineStandardTestSuite struct {
	suite.Suite
	state *state.State

	testAccounts    map[string]*gtsmodel.Account
	testStatuses    map[string]*gtsmodel.Status
	highestStatusID string
	lowestStatusID  string
}

func (suite *TimelineStandardTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *TimelineStandardTestSuite) SetupTest() {
	suite.state = new(state.State)

	suite.state.Caches.Init()
	testrig.StartWorkers(suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.DB = testrig.NewTestDB(suite.state)

	testrig.StartTimelines(
		suite.state,
		visibility.NewFilter(suite.state),
		typeutils.NewConverter(suite.state),
	)

	testrig.StandardDBSetup(suite.state.DB, nil)
}

func (suite *TimelineStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
	testrig.StopWorkers(suite.state)
}

func (suite *TimelineStandardTestSuite) fillTimeline(timelineID string) {
	// Put testrig statuses in a determinate order
	// since we can't trust a map to keep order.
	statuses := []*gtsmodel.Status{}
	for _, s := range suite.testStatuses {
		statuses = append(statuses, s)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ID > statuses[j].ID
	})

	// Statuses are now highest -> lowest.
	suite.highestStatusID = statuses[0].ID
	suite.lowestStatusID = statuses[len(statuses)-1].ID
	if suite.highestStatusID < suite.lowestStatusID {
		suite.FailNow("", "statuses weren't ordered properly by sort")
	}

	// Put all test statuses into the timeline; we don't
	// need to be fussy about who sees what for these tests.
	for _, status := range statuses {
		if _, err := suite.state.Timelines.Home.IngestOne(context.Background(), timelineID, status); err != nil {
			suite.FailNow(err.Error())
		}
	}
}
