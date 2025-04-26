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

package bundb_test

import (
	"context"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"github.com/stretchr/testify/suite"
)

type SearchTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *SearchTestSuite) TestSearchAccountsTurtleAny() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "turtle", "", "", 10, false, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccounts1HappyWithPrefix() {
	testAccount := suite.testAccounts["local_account_1"]

	// Query will just look for usernames that start with "1happy".
	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "@1happy", "", "", 10, false, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccounts1HappyWithPrefixUpper() {
	testAccount := suite.testAccounts["local_account_1"]

	// Query will just look for usernames that start with "1HAPPY".
	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "@1HAPPY", "", "", 10, false, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccounts1HappyNoPrefix() {
	testAccount := suite.testAccounts["local_account_1"]

	// Query will do the full coalesce.
	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "1happy", "", "", 10, false, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsTurtleFollowing() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "turtle", "", "", 10, true, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsTurtleFollowingUpper() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "TURTLE", "", "", 10, true, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsPostFollowing() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "post", "", "", 10, true, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsPostAny() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "post", "", "", 10, false, 0)
	suite.NoError(err, db.ErrNoEntries)
	suite.Empty(accounts)
}

func (suite *SearchTestSuite) TestSearchAccountsFossAny() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "foss", "", "", 10, false, 0)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchStatuses() {
	testAccount := suite.testAccounts["local_account_1"]

	statuses, err := suite.db.SearchForStatuses(context.Background(), testAccount.ID, "hello", "", "", "", 10, 0)
	suite.NoError(err)
	suite.Len(statuses, 1)
}

func (suite *SearchTestSuite) TestSearchStatusesFromAccount() {
	testAccount := suite.testAccounts["local_account_1"]
	fromAccount := suite.testAccounts["local_account_2"]

	statuses, err := suite.db.SearchForStatuses(context.Background(), testAccount.ID, "hi", fromAccount.ID, "", "", 10, 0)
	suite.NoError(err)
	if suite.Len(statuses, 1) {
		suite.Equal(fromAccount.ID, statuses[0].AccountID)
	}
}

func (suite *SearchTestSuite) TestSearchTags() {
	// Search with full tag string.
	tags, err := suite.db.SearchForTags(context.Background(), "welcome", "", "", 10, 0)
	suite.NoError(err)
	suite.Len(tags, 1)

	// Search with partial tag string.
	tags, err = suite.db.SearchForTags(context.Background(), "wel", "", "", 10, 0)
	suite.NoError(err)
	suite.Len(tags, 1)

	// Search with end of tag string.
	tags, err = suite.db.SearchForTags(context.Background(), "come", "", "", 10, 0)
	suite.NoError(err)
	suite.Len(tags, 0)
}

func TestSearchTestSuite(t *testing.T) {
	suite.Run(t, new(SearchTestSuite))
}
