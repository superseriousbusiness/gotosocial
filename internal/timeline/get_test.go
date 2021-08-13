/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
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
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.tc = testrig.NewTestTypeConverter(suite.db)

	testrig.StandardDBSetup(suite.db, nil)

	// let's take local_account_1 as the timeline owner
	tl, err := timeline.NewTimeline(suite.testAccounts["local_account_1"].ID, suite.db, suite.tc, suite.log)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// prepare the timeline by just shoving all test statuses in it -- let's not be fussy about who sees what
	for _, s := range suite.testStatuses {
		_, err := tl.IndexAndPrepareOne(s.CreatedAt, s.ID, s.BoostOfID, s.AccountID, s.BoostOfAccountID)
		if err != nil {
			suite.FailNow(err.Error())
		}
	}

	suite.timeline = tl
}

func (suite *GetTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *GetTestSuite) TestGet() {
	// get 10 from the top
	statuses, err := suite.timeline.Get(10, "", "", "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	fmt.Println(len(statuses))
	fmt.Println(len(statuses))
	fmt.Println(len(statuses))
	fmt.Println(len(statuses))
	fmt.Println(len(statuses))
	fmt.Println(len(statuses))
	time.Sleep(1 * time.Second)
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, new(GetTestSuite))
}
