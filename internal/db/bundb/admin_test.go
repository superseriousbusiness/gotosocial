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
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AdminTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AdminTestSuite) TestIsUsernameAvailableNo() {
	available, err := suite.db.IsUsernameAvailable(suite.T().Context(), "the_mighty_zork")
	suite.NoError(err)
	suite.False(available)
}

func (suite *AdminTestSuite) TestIsUsernameAvailableYes() {
	available, err := suite.db.IsUsernameAvailable(suite.T().Context(), "someone_completely_different")
	suite.NoError(err)
	suite.True(available)
}

func (suite *AdminTestSuite) TestIsEmailAvailableNo() {
	available, err := suite.db.IsEmailAvailable(suite.T().Context(), "zork@example.org")
	suite.NoError(err)
	suite.False(available)
}

func (suite *AdminTestSuite) TestIsEmailAvailableYes() {
	available, err := suite.db.IsEmailAvailable(suite.T().Context(), "someone@somewhere.com")
	suite.NoError(err)
	suite.True(available)
}

func (suite *AdminTestSuite) TestIsEmailAvailableDomainBlocked() {
	if err := suite.db.Put(suite.T().Context(), &gtsmodel.EmailDomainBlock{
		ID:                 "01GEEV2R2YC5GRSN96761YJE47",
		Domain:             "somewhere.com",
		CreatedByAccountID: suite.testAccounts["admin_account"].ID,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	available, err := suite.db.IsEmailAvailable(suite.T().Context(), "someone@somewhere.com")
	suite.EqualError(err, "email domain somewhere.com is blocked")
	suite.False(available)
}

func (suite *AdminTestSuite) TestCreateInstanceAccount() {
	// reinitialize db caches to clear
	suite.state.Caches.Init()
	// we need to take an empty db for this...
	testrig.StandardDBTeardown(suite.db)
	// ...with tables created but no data
	suite.db = testrig.NewTestDB(&suite.state)
	testrig.CreateTestTables(suite.db)

	// make sure there's no instance account in the db yet
	acct, err := suite.db.GetInstanceAccount(suite.T().Context(), "")
	suite.Error(err)
	suite.Nil(acct)

	// create it
	err = suite.db.CreateInstanceAccount(suite.T().Context())
	suite.NoError(err)

	// and now check it exists
	acct, err = suite.db.GetInstanceAccount(suite.T().Context(), "")
	suite.NoError(err)
	suite.NotNil(acct)
}

func (suite *AdminTestSuite) TestNewSignupWithNoInstanceApp() {
	ctx := suite.T().Context()

	// Delete the instance app.
	if err := suite.state.DB.DeleteApplicationByID(
		ctx,
		suite.testApplications["instance_application"].ID,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Try to create a new signup with no provided app ID,
	// it should fail as it can't fetch the instance app.
	_, err := suite.state.DB.NewSignup(ctx, gtsmodel.NewSignup{
		Username: "whatever",
		Email:    "whatever@wherever.org",
		Password: "really_good_password",
	})
	suite.EqualError(err, "NewSignup: instance application not yet created, run the server at least once *before* creating users")
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
