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

package visibility_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StatusBoostableTestSuite struct {
	FilterStandardTestSuite
}

func (suite *StatusBoostableTestSuite) TestOwnPublicBoostable() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOwnUnlockedBoostable() {
	testStatus := suite.testStatuses["local_account_1_status_2"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOwnMutualsOnlyNonInteractiveBoostable() {
	testStatus := suite.testStatuses["local_account_1_status_3"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOwnMutualsOnlyBoostable() {
	testStatus := suite.testStatuses["local_account_1_status_4"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOwnFollowersOnlyBoostable() {
	testStatus := suite.testStatuses["local_account_1_status_5"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOwnDirectNotBoostable() {
	testStatus := suite.testStatuses["local_account_2_status_6"]
	testAccount := suite.testAccounts["local_account_2"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.False(boostable)
}

func (suite *StatusBoostableTestSuite) TestOtherPublicBoostable() {
	testStatus := suite.testStatuses["local_account_2_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOtherUnlistedBoostable() {
	testStatus := suite.testStatuses["local_account_1_status_2"]
	testAccount := suite.testAccounts["local_account_2"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(boostable)
}

func (suite *StatusBoostableTestSuite) TestOtherFollowersOnlyNotBoostable() {
	testStatus := suite.testStatuses["local_account_2_status_7"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.False(boostable)
}

func (suite *StatusBoostableTestSuite) TestOtherDirectNotBoostable() {
	testStatus := suite.testStatuses["local_account_2_status_6"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.False(boostable)
}

func (suite *StatusBoostableTestSuite) TestRemoteFollowersOnlyNotVisibleError() {
	testStatus := suite.testStatuses["local_account_1_status_5"]
	testAccount := suite.testAccounts["remote_account_1"]
	ctx := context.Background()

	boostable, err := suite.filter.StatusBoostable(ctx, testStatus, testAccount)
	suite.Assert().Error(err)

	suite.False(boostable)
}

func TestStatusBoostableTestSuite(t *testing.T) {
	suite.Run(t, new(StatusBoostableTestSuite))
}
