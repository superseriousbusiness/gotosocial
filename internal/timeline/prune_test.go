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
)

type PruneTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *PruneTestSuite) TestPrune() {
	var (
		ctx                        = context.Background()
		testAccountID              = suite.testAccounts["local_account_1"].ID
		desiredPreparedItemsLength = 5
		desiredIndexedItemsLength  = 5
	)

	suite.fillTimeline(testAccountID)

	pruned, err := suite.state.Timelines.Home.Prune(ctx, testAccountID, desiredPreparedItemsLength, desiredIndexedItemsLength)
	suite.NoError(err)
	suite.Equal(20, pruned)
	suite.Equal(5, suite.state.Timelines.Home.GetIndexedLength(ctx, testAccountID))
}

func (suite *PruneTestSuite) TestPruneTwice() {
	var (
		ctx                        = context.Background()
		testAccountID              = suite.testAccounts["local_account_1"].ID
		desiredPreparedItemsLength = 5
		desiredIndexedItemsLength  = 5
	)

	suite.fillTimeline(testAccountID)

	pruned, err := suite.state.Timelines.Home.Prune(ctx, testAccountID, desiredPreparedItemsLength, desiredIndexedItemsLength)
	suite.NoError(err)
	suite.Equal(20, pruned)
	suite.Equal(5, suite.state.Timelines.Home.GetIndexedLength(ctx, testAccountID))

	// Prune same again, nothing should be pruned this time.
	pruned, err = suite.state.Timelines.Home.Prune(ctx, testAccountID, desiredPreparedItemsLength, desiredIndexedItemsLength)
	suite.NoError(err)
	suite.Equal(0, pruned)
	suite.Equal(5, suite.state.Timelines.Home.GetIndexedLength(ctx, testAccountID))
}

func (suite *PruneTestSuite) TestPruneTo0() {
	var (
		ctx                        = context.Background()
		testAccountID              = suite.testAccounts["local_account_1"].ID
		desiredPreparedItemsLength = 0
		desiredIndexedItemsLength  = 0
	)

	suite.fillTimeline(testAccountID)

	pruned, err := suite.state.Timelines.Home.Prune(ctx, testAccountID, desiredPreparedItemsLength, desiredIndexedItemsLength)
	suite.NoError(err)
	suite.Equal(25, pruned)
	suite.Equal(0, suite.state.Timelines.Home.GetIndexedLength(ctx, testAccountID))
}

func (suite *PruneTestSuite) TestPruneToInfinityAndBeyond() {
	var (
		ctx                        = context.Background()
		testAccountID              = suite.testAccounts["local_account_1"].ID
		desiredPreparedItemsLength = 9999999
		desiredIndexedItemsLength  = 9999999
	)

	suite.fillTimeline(testAccountID)

	pruned, err := suite.state.Timelines.Home.Prune(ctx, testAccountID, desiredPreparedItemsLength, desiredIndexedItemsLength)
	suite.NoError(err)
	suite.Equal(0, pruned)
	suite.Equal(25, suite.state.Timelines.Home.GetIndexedLength(ctx, testAccountID))
}

func TestPruneTestSuite(t *testing.T) {
	suite.Run(t, new(PruneTestSuite))
}
