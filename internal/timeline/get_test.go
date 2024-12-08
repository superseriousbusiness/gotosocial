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
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
)

type GetTestSuite struct {
	TimelineStandardTestSuite
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

func (suite *GetTestSuite) emptyAccountFollows(ctx context.Context, accountID string) {
	// Get all of account's follows.
	follows, err := suite.state.DB.GetAccountFollows(
		gtscontext.SetBarebones(ctx),
		accountID,
		nil, // select all
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Remove each follow.
	for _, follow := range follows {
		if err := suite.state.DB.DeleteFollowByID(ctx, follow.ID); err != nil {
			suite.FailNow(err.Error())
		}
	}

	// Ensure no follows left.
	follows, err = suite.state.DB.GetAccountFollows(
		gtscontext.SetBarebones(ctx),
		accountID,
		nil, // select all
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if len(follows) != 0 {
		suite.FailNow("follows should be empty")
	}
}

func (suite *GetTestSuite) emptyAccountStatuses(ctx context.Context, accountID string) {
	// Get all of account's statuses.
	statuses, err := suite.state.DB.GetAccountStatuses(
		ctx,
		accountID,
		9999,
		false,
		false,
		id.Highest,
		id.Lowest,
		false,
		false,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Remove each status.
	for _, status := range statuses {
		if err := suite.state.DB.DeleteStatusByID(ctx, status.ID); err != nil {
			suite.FailNow(err.Error())
		}
	}
}

func (suite *GetTestSuite) TestGetNewTimelinePageDown() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = ""
		limit       = 5
		local       = false
	)

	// Get 5 from the top.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 5)

	// Get 5 from next maxID.
	maxID = statuses[len(statuses)-1].GetID()
	statuses, err = suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, maxID, id.Lowest, 5)
}

func (suite *GetTestSuite) TestGetNewTimelinePageUp() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = id.Lowest
		limit       = 5
		local       = false
	)

	// Get 5 from the back.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, minID, 5)

	// Page up from next minID.
	minID = statuses[0].GetID()
	statuses, err = suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, minID, 5)
}

func (suite *GetTestSuite) TestGetNewTimelineMoreThanPossible() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = ""
		limit       = 100
		local       = false
	)

	// Get 100 from the top.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 22)
}

func (suite *GetTestSuite) TestGetNewTimelineMoreThanPossiblePageUp() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = id.Lowest
		limit       = 100
		local       = false
	)

	// Get 100 from the back.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 22)
}

func (suite *GetTestSuite) TestGetNewTimelineNoFollowing() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = ""
		limit       = 10
		local       = false
	)

	suite.emptyAccountFollows(ctx, testAccount.ID)

	// Try to get 10 from the top of the timeline.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 9)

	for _, s := range statuses {
		if s.GetAccountID() != testAccount.ID {
			suite.FailNow("timeline with no follows should only contain posts by timeline owner account")
		}
	}
}

func (suite *GetTestSuite) TestGetNewTimelineNoFollowingNoStatuses() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = ""
		limit       = 5
		local       = false
	)

	suite.emptyAccountFollows(ctx, testAccount.ID)
	suite.emptyAccountStatuses(ctx, testAccount.ID)

	// Try to get 5 from the top of the timeline.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(statuses, id.Highest, id.Lowest, 0)
}

func (suite *GetTestSuite) TestGetNoParams() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = ""
		limit       = 10
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Get 10 statuses from the top (no params).
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, id.Lowest, 10)

	// First status should have the highest ID in the testrig.
	suite.Equal(suite.highestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetMaxID() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = "01F8MHBQCBTDKN6X5VHGMMN4MA"
		sinceID     = ""
		minID       = ""
		limit       = 10
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 10 with a max ID somewhere in the middle of the stack.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// We'll only get 6 statuses back.
	suite.checkStatuses(statuses, maxID, id.Lowest, 6)
}

func (suite *GetTestSuite) TestGetSinceID() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = "01F8MHBQCBTDKN6X5VHGMMN4MA"
		minID       = ""
		limit       = 10
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 10 with a since ID somewhere in the middle of the stack.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, sinceID, 10)

	// The first status in the stack should have the highest ID of all
	// in the testrig, because we're paging down.
	suite.Equal(suite.highestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetSinceIDOneOnly() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = "01F8MHBQCBTDKN6X5VHGMMN4MA"
		minID       = ""
		limit       = 1
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 1 with a since ID somewhere in the middle of the stack.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, sinceID, 1)

	// The one status we got back should have the highest ID of all in
	// the testrig, because using sinceID means we're paging down.
	suite.Equal(suite.highestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetMinID() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = "01F8MHBQCBTDKN6X5VHGMMN4MA"
		limit       = 5
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 5 with a min ID somewhere in the middle of the stack.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 5)

	// We're paging up so even the highest status ID in the pile
	// shouldn't be the highest ID we have.
	suite.NotEqual(suite.highestStatusID, statuses[0])
}

func (suite *GetTestSuite) TestGetMinIDOneOnly() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = "01F8MHBQCBTDKN6X5VHGMMN4MA"
		limit       = 1
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 1 with a min ID somewhere in the middle of the stack.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 1)

	// The one status we got back should have the an ID equal to the
	// one ID immediately newer than it.
	suite.Equal("01F8MHC0H0A7XHTVH5F596ZKBM", statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetMinIDFromLowestInTestrig() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = suite.lowestStatusID
		limit       = 1
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 1 with minID equal to the lowest status in the testrig.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 1)

	// The one status we got back should have an id higher than
	// the lowest status in the testrig, since minID is not inclusive.
	suite.Greater(statuses[0].GetID(), suite.lowestStatusID)
}

func (suite *GetTestSuite) TestGetMinIDFromLowestPossible() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = id.Lowest
		limit       = 1
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 1 with the lowest possible min ID.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(statuses, id.Highest, minID, 1)

	// The one status we got back should have the an ID equal to the
	// lowest ID status in the test rig.
	suite.Equal(suite.lowestStatusID, statuses[0].GetID())
}

func (suite *GetTestSuite) TestGetBetweenID() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = "01F8MHCP5P2NWYQ416SBA0XSEV"
		sinceID     = ""
		minID       = "01F8MHBQCBTDKN6X5VHGMMN4MA"
		limit       = 10
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 10 between these two IDs
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// There's only two statuses between these two IDs.
	suite.checkStatuses(statuses, maxID, minID, 2)
}

func (suite *GetTestSuite) TestGetBetweenIDImpossible() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = id.Lowest
		sinceID     = ""
		minID       = id.Highest
		limit       = 10
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Ask for 10 between these two IDs which present
	// an impossible query.
	statuses, err := suite.state.Timelines.Home.GetTimeline(
		ctx,
		testAccount.ID,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// We should have nothing back.
	suite.checkStatuses(statuses, maxID, minID, 0)
}

func (suite *GetTestSuite) TestGetTimelinesAsync() {
	var (
		ctx           = context.Background()
		accountToNuke = suite.testAccounts["local_account_1"]
		maxID         = ""
		sinceID       = ""
		minID         = ""
		limit         = 5
		local         = false
		multiplier    = 5
	)

	// Nuke one account's statuses and follows,
	// as though the account had just been created.
	suite.emptyAccountFollows(ctx, accountToNuke.ID)
	suite.emptyAccountStatuses(ctx, accountToNuke.ID)

	// Get 5 statuses from each timeline in
	// our testrig at the same time, five times.
	wg := new(sync.WaitGroup)
	wg.Add(len(suite.testAccounts) * multiplier)

	for i := 0; i < multiplier; i++ {
		go func() {
			for _, testAccount := range suite.testAccounts {
				if _, err := suite.state.Timelines.Home.GetTimeline(
					ctx,
					testAccount.ID,
					maxID,
					sinceID,
					minID,
					limit,
					local,
				); err != nil {
					suite.Fail(err.Error())
				}

				wg.Done()
			}
		}()
	}

	wg.Wait() // Wait until all get calls have returned.
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, new(GetTestSuite))
}
