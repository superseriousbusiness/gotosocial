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

package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20211113114307_init"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AdminTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AdminTestSuite) TestIsUsernameAvailableNo() {
	available, err := suite.db.IsUsernameAvailable(context.Background(), "the_mighty_zork")
	suite.NoError(err)
	suite.False(available)
}

func (suite *AdminTestSuite) TestIsUsernameAvailableYes() {
	available, err := suite.db.IsUsernameAvailable(context.Background(), "someone_completely_different")
	suite.NoError(err)
	suite.True(available)
}

func (suite *AdminTestSuite) TestIsEmailAvailableNo() {
	available, err := suite.db.IsEmailAvailable(context.Background(), "zork@example.org")
	suite.NoError(err)
	suite.False(available)
}

func (suite *AdminTestSuite) TestIsEmailAvailableYes() {
	available, err := suite.db.IsEmailAvailable(context.Background(), "someone@somewhere.com")
	suite.NoError(err)
	suite.True(available)
}

func (suite *AdminTestSuite) TestIsEmailAvailableDomainBlocked() {
	if err := suite.db.Put(context.Background(), &gtsmodel.EmailDomainBlock{
		ID:                 "01GEEV2R2YC5GRSN96761YJE47",
		Domain:             "somewhere.com",
		CreatedByAccountID: suite.testAccounts["admin_account"].ID,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	available, err := suite.db.IsEmailAvailable(context.Background(), "someone@somewhere.com")
	suite.EqualError(err, "email domain somewhere.com is blocked")
	suite.False(available)
}

func (suite *AdminTestSuite) TestCreateInstanceAccount() {
	// reinitialize test DB to clear caches
	suite.db = testrig.NewTestDB()
	// we need to take an empty db for this...
	testrig.StandardDBTeardown(suite.db)
	// ...with tables created but no data
	testrig.CreateTestTables(suite.db)

	// make sure there's no instance account in the db yet
	acct, err := suite.db.GetInstanceAccount(context.Background(), "")
	suite.Error(err)
	suite.Nil(acct)

	// create it
	err = suite.db.CreateInstanceAccount(context.Background())
	suite.NoError(err)

	// and now check it exists
	acct, err = suite.db.GetInstanceAccount(context.Background(), "")
	suite.NoError(err)
	suite.NotNil(acct)
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
