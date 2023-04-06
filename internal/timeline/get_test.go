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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type GetTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *GetTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *GetTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.filter = visibility.NewFilter(&suite.state)

	testrig.StandardDBSetup(suite.db, nil)

	// Take local_account_1 as the timeline owner, it
	// doesn't really matter too much for these tests.
	tl := timeline.NewTimeline(
		context.Background(),
		suite.testAccounts["local_account_1"].ID,
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)

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
	for _, s := range statuses {
		_, err := tl.IndexAndPrepareOne(context.Background(), s.GetID(), s.BoostOfID, s.AccountID, s.BoostOfAccountID)
		if err != nil {
			suite.FailNow(err.Error())
		}
	}

	suite.timeline = tl
}

func (suite *GetTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *GetTestSuite) checkStatuses(statuses []timeline.Preparable, maxID string, minID string, expectedLength int) {
	if l := len(statuses); l != expectedLength {
		suite.FailNow("", "expected %d statuses in slice, got %d", expectedLength, l)
	} else if l == 0 {
		// Can't test empty slice.
		return
	}

	// Check ordering + bounds of statuses.
	highest := statuses[0].GetID()
	for _, status := range statuses {
		id := status.GetID()

		if id >= maxID {
			suite.FailNow("", "%s greater than maxID %s", id, maxID)
		}

		if id <= minID {
			suite.FailNow("", "%s smaller than minID %s", id, minID)
		}

		if id > highest {
			suite.FailNow("", "statuses in slice were not ordered highest -> lowest ID")
		}

		highest = id
	}
}

func (suite *GetTestSuite) TestGetNewTimelinePageDown() {
	// Take a fresh timeline for this test.
	// This tests whether indexing works
	// properly against uninitialized timelines.
	tl := timeline.NewTimeline(
		context.Background(),
		suite.testAccounts["local_account_1"].ID,
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)

	// Get 5 from the top.
	statuses, err := tl.Get(context.Background(), 5, "", "", "", true)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 5)

	// Get 5 from next maxID.
	nextMaxID := statuses[len(statuses)-1].GetID()
	statuses, err = tl.Get(context.Background(), 5, nextMaxID, "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, nextMaxID, id.Lowest, 5)
}

func (suite *GetTestSuite) TestGetNewTimelinePageUp() {
	// Take a fresh timeline for this test.
	// This tests whether indexing works
	// properly against uninitialized timelines.
	tl := timeline.NewTimeline(
		context.Background(),
		suite.testAccounts["local_account_1"].ID,
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)

	// Get 5 from the back.
	statuses, err := tl.Get(context.Background(), 5, "", "", id.Lowest, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 5)

	// Page upwards.
	nextMinID := statuses[len(statuses)-1].GetID()
	statuses, err = tl.Get(context.Background(), 5, "", "", nextMinID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, nextMinID, 5)
}

func (suite *GetTestSuite) TestGetNewTimelineMoreThanPossible() {
	// Take a fresh timeline for this test.
	// This tests whether indexing works
	// properly against uninitialized timelines.
	tl := timeline.NewTimeline(
		context.Background(),
		suite.testAccounts["local_account_1"].ID,
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)

	// Get 100 from the top.
	statuses, err := tl.Get(context.Background(), 100, id.Highest, "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 16)
}

func (suite *GetTestSuite) TestGetNewTimelineMoreThanPossiblePageUp() {
	// Take a fresh timeline for this test.
	// This tests whether indexing works
	// properly against uninitialized timelines.
	tl := timeline.NewTimeline(
		context.Background(),
		suite.testAccounts["local_account_1"].ID,
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)

	// Get 100 from the back.
	statuses, err := tl.Get(context.Background(), 100, "", "", id.Lowest, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 16)
}

func (suite *GetTestSuite) TestGetNoParams() {
	// Get 10 statuses from the top (no params).
	statuses, err := suite.timeline.Get(context.Background(), 10, "", "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, id.Lowest, 10)

	// First status should have the highest ID in the testrig.
	suite.Equal(suite.highestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetMaxID() {
	// Ask for 10 with a max ID somewhere in the middle of the stack.
	maxID := "01F8MHBQCBTDKN6X5VHGMMN4MA"

	statuses, err := suite.timeline.Get(context.Background(), 10, maxID, "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// We'll only get 6 statuses back.
	suite.checkStatuses(statuses, maxID, id.Lowest, 6)
}

func (suite *GetTestSuite) TestGetSinceID() {
	// Ask for 10 with a since ID somewhere in the middle of the stack.
	sinceID := "01F8MHBQCBTDKN6X5VHGMMN4MA"
	statuses, err := suite.timeline.Get(context.Background(), 10, "", sinceID, "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, sinceID, 10)

	// The first status in the stack should have the highest ID of all
	// in the testrig, because we're paging down.
	suite.Equal(suite.highestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetSinceIDOneOnly() {
	// Ask for 1 with a since ID somewhere in the middle of the stack.
	sinceID := "01F8MHBQCBTDKN6X5VHGMMN4MA"
	statuses, err := suite.timeline.Get(context.Background(), 1, "", sinceID, "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, sinceID, 1)

	// The one status we got back should have the highest ID of all in
	// the testrig, because using sinceID means we're paging down.
	suite.Equal(suite.highestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetMinID() {
	// Ask for 5 with a min ID somewhere in the middle of the stack.
	minID := "01F8MHBQCBTDKN6X5VHGMMN4MA"
	statuses, err := suite.timeline.Get(context.Background(), 5, "", "", minID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 5)

	// We're paging up so even the highest status ID in the pile
	// shouldn't be the highest ID we have.
	suite.NotEqual(suite.highestStatusID, statuses[0])
}

func (suite *GetTestSuite) TestGetMinIDOneOnly() {
	// Ask for 1 with a min ID somewhere in the middle of the stack.
	minID := "01F8MHBQCBTDKN6X5VHGMMN4MA"
	statuses, err := suite.timeline.Get(context.Background(), 1, "", "", minID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 1)

	// The one status we got back should have the an ID equal to the
	// one ID immediately newer than it.
	suite.Equal("01F8MHC0H0A7XHTVH5F596ZKBM", statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetMinIDFromLowestInTestrig() {
	// Ask for 1 with minID equal to the lowest status in the testrig.
	minID := suite.lowestStatusID
	statuses, err := suite.timeline.Get(context.Background(), 1, "", "", minID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 1)

	// The one status we got back should have an id higher than
	// the lowest status in the testrig, since minID is not inclusive.
	suite.Greater(statuses[0].GetID(), suite.lowestStatusID)
}

func (suite *GetTestSuite) TestGetMinIDFromLowestPossible() {
	// Ask for 1 with the lowest possible min ID.
	minID := id.Lowest
	statuses, err := suite.timeline.Get(context.Background(), 1, "", "", minID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 1)

	// The one status we got back should have the an ID equal to the
	// lowest ID status in the test rig.
	suite.Equal(suite.lowestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetBetweenID() {
	// Ask for 10 between these two IDs
	maxID := "01F8MHCP5P2NWYQ416SBA0XSEV"
	minID := "01F8MHBQCBTDKN6X5VHGMMN4MA"

	statuses, err := suite.timeline.Get(context.Background(), 10, maxID, "", minID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// There's only two statuses between these two IDs.
	suite.checkStatuses(statuses, maxID, minID, 2)
}

func (suite *GetTestSuite) TestGetBetweenIDImpossible() {
	// Ask for 10 between these two IDs which present
	// an impossible query.
	maxID := id.Lowest
	minID := id.Highest

	statuses, err := suite.timeline.Get(context.Background(), 10, maxID, "", minID, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// We should have nothing back.
	suite.checkStatuses(statuses, maxID, minID, 0)
}

func (suite *GetTestSuite) TestLastGot() {
	// LastGot should be zero
	suite.Zero(suite.timeline.LastGot())

	// Get some from the top
	_, err := suite.timeline.Get(context.Background(), 10, "", "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// LastGot should be updated
	suite.WithinDuration(time.Now(), suite.timeline.LastGot(), 1*time.Second)
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, new(GetTestSuite))
}
