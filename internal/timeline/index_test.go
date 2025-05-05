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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type IndexTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *IndexTestSuite) TestOldestIndexedItemIDEmpty() {
	var (
		ctx           = context.Background()
		testAccountID = suite.testAccounts["local_account_1"].ID
	)

	// the oldest indexed post should be an empty string since there's nothing indexed yet
	postID := suite.state.Timelines.Home.GetOldestIndexedID(ctx, testAccountID)
	suite.Empty(postID)

	// indexLength should be 0
	suite.Zero(0, suite.state.Timelines.Home.GetIndexedLength(ctx, testAccountID))
}

func (suite *IndexTestSuite) TestIndexAlreadyIndexed() {
	var (
		ctx           = context.Background()
		testAccountID = suite.testAccounts["local_account_1"].ID
		testStatus    = suite.testStatuses["local_account_1_status_1"]
	)

	// index one post -- it should be indexed
	indexed, err := suite.state.Timelines.Home.IngestOne(ctx, testAccountID, testStatus)
	suite.NoError(err)
	suite.True(indexed)

	// try to index the same post again -- it should not be indexed
	indexed, err = suite.state.Timelines.Home.IngestOne(ctx, testAccountID, testStatus)
	suite.NoError(err)
	suite.False(indexed)
}

func (suite *IndexTestSuite) TestIndexBoostOfAlreadyIndexed() {
	var (
		ctx               = context.Background()
		testAccountID     = suite.testAccounts["local_account_1"].ID
		testStatus        = suite.testStatuses["local_account_1_status_1"]
		boostOfTestStatus = &gtsmodel.Status{
			CreatedAt:        time.Now(),
			ID:               "01FD4TA6G2Z6M7W8NJQ3K5WXYD",
			BoostOfID:        testStatus.ID,
			AccountID:        "01FD4TAY1C0NGEJVE9CCCX7QKS",
			BoostOfAccountID: testStatus.AccountID,
		}
	)

	// index one post -- it should be indexed
	indexed, err := suite.state.Timelines.Home.IngestOne(ctx, testAccountID, testStatus)
	suite.NoError(err)
	suite.True(indexed)

	// try to index the a boost of that post -- it should not be indexed
	indexed, err = suite.state.Timelines.Home.IngestOne(ctx, testAccountID, boostOfTestStatus)
	suite.NoError(err)
	suite.False(indexed)
}

func TestIndexTestSuite(t *testing.T) {
	suite.Run(t, new(IndexTestSuite))
}
