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
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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

func (suite *GetTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *GetTestSuite) TestGetDefault() {
	// lastGot should be zero
	suite.Zero(suite.timeline.LastGot())

	// get 10 20 the top and don't prepare the next query
	statuses, err := suite.timeline.Get(context.Background(), 20, "", "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we only have 16 statuses in the test suite
	suite.Len(statuses, 17)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}

	// lastGot should be up to date
	suite.WithinDuration(time.Now(), suite.timeline.LastGot(), 1*time.Second)
}

func (suite *GetTestSuite) TestGetDefaultPrepareNext() {
	// get 10 from the top and prepare the next query
	statuses, err := suite.timeline.Get(context.Background(), 10, "", "", "", true)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(statuses, 10)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}

	// sleep a second so the next query can run
	time.Sleep(1 * time.Second)
}

func (suite *GetTestSuite) TestGetMaxID() {
	// ask for 10 with a max ID somewhere in the middle of the stack
	statuses, err := suite.timeline.Get(context.Background(), 10, "01F8MHBQCBTDKN6X5VHGMMN4MA", "", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 6 statuses back, since we asked for a max ID that excludes some of our entries
	suite.Len(statuses, 6)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}
}

func (suite *GetTestSuite) TestGetMaxIDPrepareNext() {
	// ask for 10 with a max ID somewhere in the middle of the stack
	statuses, err := suite.timeline.Get(context.Background(), 10, "01F8MHBQCBTDKN6X5VHGMMN4MA", "", "", true)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 6 statuses back, since we asked for a max ID that excludes some of our entries
	suite.Len(statuses, 6)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}

	// sleep a second so the next query can run
	time.Sleep(1 * time.Second)
}

func (suite *GetTestSuite) TestGetMinID() {
	// ask for 15 with a min ID somewhere in the middle of the stack
	statuses, err := suite.timeline.Get(context.Background(), 10, "", "01F8MHBQCBTDKN6X5VHGMMN4MA", "", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 10 statuses back, since we asked for a min ID that excludes some of our entries
	suite.Len(statuses, 10)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}
}

func (suite *GetTestSuite) TestGetSinceID() {
	// ask for 15 with a since ID somewhere in the middle of the stack
	statuses, err := suite.timeline.Get(context.Background(), 15, "", "", "01F8MHBQCBTDKN6X5VHGMMN4MA", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 10 statuses back, since we asked for a since ID that excludes some of our entries
	suite.Len(statuses, 10)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}
}

func (suite *GetTestSuite) TestGetSinceIDPrepareNext() {
	// ask for 15 with a since ID somewhere in the middle of the stack
	statuses, err := suite.timeline.Get(context.Background(), 15, "", "", "01F8MHBQCBTDKN6X5VHGMMN4MA", true)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 10 statuses back, since we asked for a since ID that excludes some of our entries
	suite.Len(statuses, 10)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}

	// sleep a second so the next query can run
	time.Sleep(1 * time.Second)
}

func (suite *GetTestSuite) TestGetBetweenID() {
	// ask for 10 between these two IDs
	statuses, err := suite.timeline.Get(context.Background(), 10, "01F8MHCP5P2NWYQ416SBA0XSEV", "", "01F8MHBQCBTDKN6X5VHGMMN4MA", false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 2 statuses back, since there are only two statuses between the given IDs
	suite.Len(statuses, 2)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}
}

func (suite *GetTestSuite) TestGetBetweenIDPrepareNext() {
	// ask for 10 between these two IDs
	statuses, err := suite.timeline.Get(context.Background(), 10, "01F8MHCP5P2NWYQ416SBA0XSEV", "", "01F8MHBQCBTDKN6X5VHGMMN4MA", true)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// we should only get 2 statuses back, since there are only two statuses between the given IDs
	suite.Len(statuses, 2)

	// statuses should be sorted highest to lowest ID
	var highest string
	for i, s := range statuses {
		if i == 0 {
			highest = s.GetID()
		} else {
			suite.Less(s.GetID(), highest)
			highest = s.GetID()
		}
	}

	// sleep a second so the next query can run
	time.Sleep(1 * time.Second)
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, new(GetTestSuite))
}
