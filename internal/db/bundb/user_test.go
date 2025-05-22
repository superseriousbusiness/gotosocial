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
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *UserTestSuite) TestGetAllUsers() {
	users, err := suite.db.GetAllUsers(suite.T().Context())
	suite.NoError(err)
	suite.Len(users, len(suite.testUsers))
}

func (suite *UserTestSuite) TestGetUser() {
	user, err := suite.db.GetUserByID(suite.T().Context(), suite.testUsers["local_account_1"].ID)
	suite.NoError(err)
	suite.NotNil(user)
}

func (suite *UserTestSuite) TestGetUserByEmailAddress() {
	user, err := suite.db.GetUserByEmailAddress(suite.T().Context(), suite.testUsers["local_account_1"].Email)
	suite.NoError(err)
	suite.NotNil(user)
}

func (suite *UserTestSuite) TestGetUserByAccountID() {
	user, err := suite.db.GetUserByAccountID(suite.T().Context(), suite.testAccounts["local_account_1"].ID)
	suite.NoError(err)
	suite.NotNil(user)
}

func (suite *UserTestSuite) TestUpdateUserSelectedColumns() {
	testUser := suite.testUsers["local_account_1"]

	updateUser := new(gtsmodel.User)
	*updateUser = *testUser
	updateUser.Email = "whatever"
	updateUser.Locale = "es"

	err := suite.db.UpdateUser(suite.T().Context(), updateUser)
	suite.NoError(err)

	dbUser, err := suite.db.GetUserByID(suite.T().Context(), testUser.ID)
	suite.NoError(err)
	suite.NotNil(dbUser)
	suite.Equal(updateUser.Email, dbUser.Email)
	suite.Equal(updateUser.Locale, dbUser.Locale)
	suite.Equal(testUser.AccountID, dbUser.AccountID)
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
