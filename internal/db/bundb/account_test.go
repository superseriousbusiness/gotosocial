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

package bundb_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type AccountTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AccountTestSuite) TestGetAccountByIDWithExtras() {
	account, err := suite.db.GetAccountByID(context.Background(), suite.testAccounts["local_account_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(account)
	suite.NotNil(account.AvatarMediaAttachment)
	suite.NotEmpty(account.AvatarMediaAttachment.URL)
	suite.NotNil(account.HeaderMediaAttachment)
	suite.NotEmpty(account.HeaderMediaAttachment.URL)
}

func (suite *AccountTestSuite) TestUpdateAccount() {
	testAccount := suite.testAccounts["local_account_1"]

	testAccount.DisplayName = "new display name!"

	_, err := suite.db.UpdateAccount(context.Background(), testAccount)
	suite.NoError(err)

	updated, err := suite.db.GetAccountByID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal("new display name!", updated.DisplayName)
	suite.WithinDuration(time.Now(), updated.UpdatedAt, 5*time.Second)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
