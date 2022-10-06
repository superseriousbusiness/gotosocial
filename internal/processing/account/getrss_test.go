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

package account_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GetRSSTestSuite struct {
	AccountStandardTestSuite
}

func (suite *GetRSSTestSuite) TestGetAccountRSSAdmin() {
	testAccount := suite.testAccounts["local_account_1"]
	rssFeed, err := suite.accountProcessor.GetRSSFeedForUsername(context.Background(), testAccount, "admin")
	suite.NoError(err)
	suite.Equal(``, rssFeed)
}

func (suite *GetRSSTestSuite) TestGetAccountRSSZork() {
	testAccount := suite.testAccounts["admin_account"]
	rssFeed, err := suite.accountProcessor.GetRSSFeedForUsername(context.Background(), testAccount, "the_mighty_zork")
	suite.NoError(err)
	suite.Equal(``, rssFeed)
}

func TestGetRSSTestSuite(t *testing.T) {
	suite.Run(t, new(GetRSSTestSuite))
}
