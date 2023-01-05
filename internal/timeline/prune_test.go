/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package timeline_test

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type PruneTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *PruneTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *PruneTestSuite) SetupTest() {
	testrig.InitTestLog()
	testrig.InitTestConfig()

	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.filter = visibility.NewFilter(suite.db)

	testrig.StandardDBSetup(suite.db, nil)

	// let's take local_account_1 as the timeline owner
	tl, err := timeline.NewTimeline(
		context.Background(),
		suite.testAccounts["local_account_1"].ID,
		processing.StatusGrabFunction(suite.db),
		processing.StatusFilterFunction(suite.db, suite.filter),
		processing.StatusPrepareFunction(suite.db, suite.tc),
		processing.StatusSkipInsertFunction(),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// put the status IDs in a determinate order since we can't trust a map to keep its order
	statuses := []*gtsmodel.Status{}
	for _, s := range suite.testStatuses {
		statuses = append(statuses, s)
	}
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ID > statuses[j].ID
	})

	// prepare the timeline by just shoving all test statuses in it -- let's not be fussy about who sees what
	for _, s := range statuses {
		_, err := tl.IndexAndPrepareOne(context.Background(), s.GetID(), s.BoostOfID, s.AccountID, s.BoostOfAccountID)
		if err != nil {
			suite.FailNow(err.Error())
		}
	}

	suite.timeline = tl
}

func (suite *PruneTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *PruneTestSuite) TestPrune() {
	// prune down to 5 prepared + 5 indexed
	suite.Equal(24, suite.timeline.Prune(5, 5))
	suite.Equal(5, suite.timeline.ItemIndexLength(context.Background()))
}

func (suite *PruneTestSuite) TestPruneTo0() {
	// prune down to 0 prepared + 0 indexed
	suite.Equal(34, suite.timeline.Prune(0, 0))
	suite.Equal(0, suite.timeline.ItemIndexLength(context.Background()))
}

func (suite *PruneTestSuite) TestPruneToInfinityAndBeyond() {
	// prune to 99999, this should result in no entries being pruned
	suite.Equal(0, suite.timeline.Prune(99999, 99999))
	suite.Equal(17, suite.timeline.ItemIndexLength(context.Background()))
}

func TestPruneTestSuite(t *testing.T) {
	suite.Run(t, new(PruneTestSuite))
}
