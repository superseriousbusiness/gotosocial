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

package status_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StatusBookmarkTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusBookmarkTestSuite) TestBookmark() {
	ctx := suite.T().Context()

	// bookmark a status
	bookmarkingAccount1 := suite.testAccounts["local_account_1"]
	targetStatus1 := suite.testStatuses["admin_account_status_1"]

	bookmark1, err := suite.status.BookmarkCreate(ctx, bookmarkingAccount1, targetStatus1.ID)
	suite.NoError(err)
	suite.NotNil(bookmark1)
	suite.True(bookmark1.Bookmarked)
	suite.Equal(targetStatus1.ID, bookmark1.ID)
}

func (suite *StatusBookmarkTestSuite) TestUnbookmark() {
	ctx := suite.T().Context()

	// bookmark a status
	bookmarkingAccount1 := suite.testAccounts["local_account_1"]
	targetStatus1 := suite.testStatuses["admin_account_status_1"]

	bookmark1, err := suite.status.BookmarkCreate(ctx, bookmarkingAccount1, targetStatus1.ID)
	suite.NoError(err)
	suite.NotNil(bookmark1)
	suite.True(bookmark1.Bookmarked)
	suite.Equal(targetStatus1.ID, bookmark1.ID)

	bookmark2, err := suite.status.BookmarkRemove(ctx, bookmarkingAccount1, targetStatus1.ID)
	suite.NoError(err)
	suite.NotNil(bookmark2)
	suite.False(bookmark2.Bookmarked)
	suite.Equal(targetStatus1.ID, bookmark1.ID)
}

func TestStatusBookmarkTestSuite(t *testing.T) {
	suite.Run(t, new(StatusBookmarkTestSuite))
}
