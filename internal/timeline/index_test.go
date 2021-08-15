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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type IndexTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *IndexTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *IndexTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.tc = testrig.NewTestTypeConverter(suite.db)

	testrig.StandardDBSetup(suite.db, nil)

	// let's take local_account_1 as the timeline owner, and start with an empty timeline
	tl, err := timeline.NewTimeline(suite.testAccounts["local_account_1"].ID, suite.db, suite.tc, suite.log)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.timeline = tl
}

func (suite *IndexTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *IndexTestSuite) TestIndexBeforeLowID() {
	// index 10 before the lowest status ID possible
	err := suite.timeline.IndexBefore("00000000000000000000000000", true, 10)
	suite.NoError(err)

	// the oldest indexed post should be the lowest one we have in our testrig
	postID, err := suite.timeline.OldestIndexedPostID()
	suite.NoError(err)
	suite.Equal("01F8MHAAY43M6RJ473VQFCVH37", postID)

	// indexLength should only be 9 because that's all this user has hometimelineable
	indexLength := suite.timeline.PostIndexLength()
	suite.Equal(9, indexLength)
}

func (suite *IndexTestSuite) TestIndexBeforeHighID() {
	// index 10 before the highest status ID possible
	err := suite.timeline.IndexBefore("ZZZZZZZZZZZZZZZZZZZZZZZZZZ", true, 10)
	suite.NoError(err)

	// the oldest indexed post should be empty
	postID, err := suite.timeline.OldestIndexedPostID()
	suite.NoError(err)
	suite.Empty(postID)

	// indexLength should be 0
	indexLength := suite.timeline.PostIndexLength()
	suite.Equal(0, indexLength)
}

func (suite *IndexTestSuite) TestIndexBehindHighID() {
	// index 10 behind the highest status ID possible
	err := suite.timeline.IndexBehind("ZZZZZZZZZZZZZZZZZZZZZZZZZZ", true, 10)
	suite.NoError(err)

	// the newest indexed post should be the highest one we have in our testrig
	postID, err := suite.timeline.NewestIndexedPostID()
	suite.NoError(err)
	suite.Equal("01FCTA44PW9H1TB328S9AQXKDS", postID)

	// indexLength should only be 11 because that's all this user has hometimelineable
	indexLength := suite.timeline.PostIndexLength()
	suite.Equal(11, indexLength)
}

func (suite *IndexTestSuite) TestIndexBehindLowID() {
	// index 10 behind the lowest status ID possible
	err := suite.timeline.IndexBehind("00000000000000000000000000", true, 10)
	suite.NoError(err)

	// the newest indexed post should be empty
	postID, err := suite.timeline.NewestIndexedPostID()
	suite.NoError(err)
	suite.Empty(postID)

	// indexLength should be 0
	indexLength := suite.timeline.PostIndexLength()
	suite.Equal(0, indexLength)
}

func (suite *IndexTestSuite) TestOldestIndexedPostIDEmpty() {
	// the oldest indexed post should be an empty string since there's nothing indexed yet
	postID, err := suite.timeline.OldestIndexedPostID()
	suite.NoError(err)
	suite.Empty(postID)

	// indexLength should be 0
	indexLength := suite.timeline.PostIndexLength()
	suite.Equal(0, indexLength)
}

func (suite *IndexTestSuite) TestNewestIndexedPostIDEmpty() {
	// the newest indexed post should be an empty string since there's nothing indexed yet
	postID, err := suite.timeline.NewestIndexedPostID()
	suite.NoError(err)
	suite.Empty(postID)

	// indexLength should be 0
	indexLength := suite.timeline.PostIndexLength()
	suite.Equal(0, indexLength)
}

func (suite *IndexTestSuite) TestIndexAlreadyIndexed() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	// index one post -- it should be indexed
	indexed, err := suite.timeline.IndexOne(testStatus.CreatedAt, testStatus.ID, testStatus.BoostOfID, testStatus.AccountID, testStatus.BoostOfAccountID)
	suite.NoError(err)
	suite.True(indexed)

	// try to index the same post again -- it should not be indexed
	indexed, err = suite.timeline.IndexOne(testStatus.CreatedAt, testStatus.ID, testStatus.BoostOfID, testStatus.AccountID, testStatus.BoostOfAccountID)
	suite.NoError(err)
	suite.False(indexed)
}

func (suite *IndexTestSuite) TestIndexAndPrepareAlreadyIndexedAndPrepared() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	// index and prepare one post -- it should be indexed
	indexed, err := suite.timeline.IndexAndPrepareOne(testStatus.CreatedAt, testStatus.ID, testStatus.BoostOfID, testStatus.AccountID, testStatus.BoostOfAccountID)
	suite.NoError(err)
	suite.True(indexed)

	// try to index and prepare the same post again -- it should not be indexed
	indexed, err = suite.timeline.IndexAndPrepareOne(testStatus.CreatedAt, testStatus.ID, testStatus.BoostOfID, testStatus.AccountID, testStatus.BoostOfAccountID)
	suite.NoError(err)
	suite.False(indexed)
}

func (suite *IndexTestSuite) TestIndexBoostOfAlreadyIndexed() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	boostOfTestStatus := &gtsmodel.Status{
		CreatedAt:        time.Now(),
		ID:               "01FD4TA6G2Z6M7W8NJQ3K5WXYD",
		BoostOfID:        testStatus.ID,
		AccountID:        "01FD4TAY1C0NGEJVE9CCCX7QKS",
		BoostOfAccountID: testStatus.AccountID,
	}

	// index one post -- it should be indexed
	indexed, err := suite.timeline.IndexOne(testStatus.CreatedAt, testStatus.ID, testStatus.BoostOfID, testStatus.AccountID, testStatus.BoostOfAccountID)
	suite.NoError(err)
	suite.True(indexed)

	// try to index the a boost of that post -- it should not be indexed
	indexed, err = suite.timeline.IndexOne(boostOfTestStatus.CreatedAt, boostOfTestStatus.ID, boostOfTestStatus.BoostOfID, boostOfTestStatus.AccountID, boostOfTestStatus.BoostOfAccountID)
	suite.NoError(err)
	suite.False(indexed)
}

func TestIndexTestSuite(t *testing.T) {
	suite.Run(t, new(IndexTestSuite))
}
