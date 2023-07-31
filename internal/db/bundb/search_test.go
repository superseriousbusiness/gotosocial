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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
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

func (suite *SearchTestSuite) TestSearchAccountsTurtleFollowing() {
	testAccount := suite.testAccounts["local_account_1"]

	accounts, err := suite.db.SearchForAccounts(context.Background(), testAccount.ID, "turtle", "", "", 10, true, 0)
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

	statuses, err := suite.db.SearchForStatuses(context.Background(), testAccount.ID, "hello", "", "", 10, 0)
	suite.NoError(err)
	suite.Len(statuses, 1)
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
