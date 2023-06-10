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
	"testing"

	"github.com/stretchr/testify/suite"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

type UnprepareTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *UnprepareTestSuite) TestUnprepareFromFave() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["local_account_1"]
		maxID       = ""
		sinceID     = ""
		minID       = ""
		limit       = 1
		local       = false
	)

	suite.fillTimeline(testAccount.ID)

	// Get first status from the top (no params).
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

	if len(statuses) != 1 {
		suite.FailNow("couldn't get top status")
	}

	targetStatus := statuses[0].(*apimodel.Status)

	// Check fave stats of the top status.
	suite.Equal(0, targetStatus.FavouritesCount)
	suite.False(targetStatus.Favourited)

	// Fave the top status from testAccount.
	if err := suite.state.DB.PutStatusFave(ctx, &gtsmodel.StatusFave{
		ID:              id.NewULID(),
		AccountID:       testAccount.ID,
		TargetAccountID: targetStatus.Account.ID,
		StatusID:        targetStatus.ID,
		URI:             "https://example.org/some/activity/path",
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// Repeat call to get first status from the top.
	// Get first status from the top (no params).
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

	if len(statuses) != 1 {
		suite.FailNow("couldn't get top status")
	}

	targetStatus = statuses[0].(*apimodel.Status)

	// We haven't yet uncached/unprepared the status,
	// we've only inserted the fave, so counts should
	// stay the same...
	suite.Equal(0, targetStatus.FavouritesCount)
	suite.False(targetStatus.Favourited)

	// Now call unprepare.
	suite.state.Timelines.Home.UnprepareItemFromAllTimelines(ctx, targetStatus.ID)

	// Now a Get should trigger a fresh prepare of the
	// target status, and the counts should be updated.
	// Repeat call to get first status from the top.
	// Get first status from the top (no params).
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

	if len(statuses) != 1 {
		suite.FailNow("couldn't get top status")
	}

	targetStatus = statuses[0].(*apimodel.Status)

	suite.Equal(1, targetStatus.FavouritesCount)
	suite.True(targetStatus.Favourited)
}

func TestUnprepareTestSuite(t *testing.T) {
	suite.Run(t, new(UnprepareTestSuite))
}
