/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package status_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StatusUnbookmarkTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusUnbookmarkTestSuite) TestUnbookmark() {
	ctx := context.Background()

	// bookmark a status
	bookmarkingAccount1 := suite.testAccounts["local_account_1"]
	targetStatus1 := suite.testStatuses["admin_account_status_1"]

	bookmark1, err := suite.status.Bookmark(ctx, bookmarkingAccount1, targetStatus1.ID)
	suite.NoError(err)
	suite.NotNil(bookmark1)
	suite.True(bookmark1.Bookmarked)
	suite.Equal(targetStatus1.ID, bookmark1.ID)

	bookmark2, err := suite.status.Unbookmark(ctx, bookmarkingAccount1, targetStatus1.ID)
	suite.NoError(err)
	suite.NotNil(bookmark2)
	suite.False(bookmark2.Bookmarked)
	suite.Equal(targetStatus1.ID, bookmark1.ID)
}

func TestStatusUnbookmarkTestSuite(t *testing.T) {
	suite.Run(t, new(StatusUnbookmarkTestSuite))
}
