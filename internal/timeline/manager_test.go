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
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ManagerTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *ManagerTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *ManagerTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestLog()
	testrig.InitTestConfig()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.filter = visibility.NewFilter(&suite.state)

	testrig.StandardDBSetup(suite.db, nil)

	manager := timeline.NewManager(
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)
	suite.manager = manager
}

func (suite *ManagerTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *ManagerTestSuite) TestManagerIntegration() {
	ctx := context.Background()

	testAccount := suite.testAccounts["local_account_1"]

	// should start at 0
	indexedLen := suite.manager.GetIndexedLength(ctx, testAccount.ID)
	suite.Equal(0, indexedLen)

	// oldestIndexed should be empty string since there's nothing indexed
	oldestIndexed := suite.manager.GetOldestIndexedID(ctx, testAccount.ID)
	suite.Empty(oldestIndexed)

	// get hometimeline
	statuses, err := suite.manager.GetTimeline(ctx, testAccount.ID, "", "", "", 20, false)
	suite.NoError(err)
	suite.Len(statuses, 16)

	// now wipe the last status from all timelines, as though it had been deleted by the owner
	err = suite.manager.WipeItemFromAllTimelines(ctx, "01F8MH75CBF9JFX4ZAD54N0W0R")
	suite.NoError(err)

	// timeline should be shorter
	indexedLen = suite.manager.GetIndexedLength(ctx, testAccount.ID)
	suite.Equal(15, indexedLen)

	// oldest should now be different
	oldestIndexed = suite.manager.GetOldestIndexedID(ctx, testAccount.ID)
	suite.Equal("01F8MH82FYRXD2RC6108DAJ5HB", oldestIndexed)

	// delete the new oldest status specifically from this timeline, as though local_account_1 had muted or blocked it
	removed, err := suite.manager.Remove(ctx, testAccount.ID, "01F8MH82FYRXD2RC6108DAJ5HB")
	suite.NoError(err)
	suite.Equal(1, removed) // 1 status should be removed

	// timeline should be shorter
	indexedLen = suite.manager.GetIndexedLength(ctx, testAccount.ID)
	suite.Equal(14, indexedLen)

	// oldest should now be different
	oldestIndexed = suite.manager.GetOldestIndexedID(ctx, testAccount.ID)
	suite.Equal("01F8MHAAY43M6RJ473VQFCVH37", oldestIndexed)

	// now remove all entries by local_account_2 from the timeline
	err = suite.manager.WipeItemsFromAccountID(ctx, testAccount.ID, suite.testAccounts["local_account_2"].ID)
	suite.NoError(err)

	// timeline should be shorter
	indexedLen = suite.manager.GetIndexedLength(ctx, testAccount.ID)
	suite.Equal(7, indexedLen)

	// ingest and prepare another one into the timeline
	status := suite.testStatuses["local_account_2_status_1"]
	ingested, err := suite.manager.IngestOne(ctx, testAccount.ID, status)
	suite.NoError(err)
	suite.True(ingested)

	// timeline should be longer now
	indexedLen = suite.manager.GetIndexedLength(ctx, testAccount.ID)
	suite.Equal(8, indexedLen)

	// try to ingest same status again
	ingested, err = suite.manager.IngestOne(ctx, testAccount.ID, status)
	suite.NoError(err)
	suite.False(ingested) // should be false since it's a duplicate
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
