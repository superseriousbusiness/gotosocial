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

package visibility_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type StatusVisibleTestSuite struct {
	FilterStandardTestSuite
}

func (suite *StatusVisibleTestSuite) TestOwnStatusVisible() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(visible)
}

func (suite *StatusVisibleTestSuite) TestOwnDMVisible() {
	ctx := context.Background()

	testStatusID := suite.testStatuses["local_account_2_status_6"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["local_account_2"]

	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(visible)
}

func (suite *StatusVisibleTestSuite) TestDMVisibleToTarget() {
	ctx := context.Background()

	testStatusID := suite.testStatuses["local_account_2_status_6"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["local_account_1"]

	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(visible)
}

func (suite *StatusVisibleTestSuite) TestDMNotVisibleIfNotMentioned() {
	ctx := context.Background()

	testStatusID := suite.testStatuses["local_account_2_status_6"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["admin_account"]

	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.False(visible)
}

func (suite *StatusVisibleTestSuite) TestStatusNotVisibleIfNotMutuals() {
	ctx := context.Background()

	suite.db.DeleteByID(ctx, suite.testFollows["local_account_2_local_account_1"].ID, &gtsmodel.Follow{})

	testStatusID := suite.testStatuses["local_account_1_status_4"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["local_account_2"]

	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.False(visible)
}

func (suite *StatusVisibleTestSuite) TestStatusNotVisibleIfNotFollowing() {
	ctx := context.Background()

	suite.db.DeleteByID(ctx, suite.testFollows["admin_account_local_account_1"].ID, &gtsmodel.Follow{})

	testStatusID := suite.testStatuses["local_account_1_status_5"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["admin_account"]

	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.False(visible)
}

func (suite *StatusVisibleTestSuite) TestStatusNotVisibleIfNotMutualsCached() {
	ctx := context.Background()
	testStatusID := suite.testStatuses["local_account_1_status_4"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["local_account_2"]

	// Perform a status visibility check while mutuals, this shsould be true.
	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)
	suite.True(visible)

	err = suite.db.DeleteFollowByID(ctx, suite.testFollows["local_account_2_local_account_1"].ID)
	suite.NoError(err)

	// Perform a status visibility check after unfollow, this should be false.
	visible, err = suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)
	suite.False(visible)
}

func (suite *StatusVisibleTestSuite) TestStatusNotVisibleIfNotFollowingCached() {
	ctx := context.Background()
	testStatusID := suite.testStatuses["local_account_1_status_5"].ID
	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)
	testAccount := suite.testAccounts["admin_account"]

	// Perform a status visibility check while following, this shsould be true.
	visible, err := suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)
	suite.True(visible)

	err = suite.db.DeleteFollowByID(ctx, suite.testFollows["admin_account_local_account_1"].ID)
	suite.NoError(err)

	// Perform a status visibility check after unfollow, this should be false.
	visible, err = suite.filter.StatusVisible(ctx, testAccount, testStatus)
	suite.NoError(err)
	suite.False(visible)
}

func (suite *StatusVisibleTestSuite) TestVisiblePending() {
	ctx := context.Background()

	// Copy the test status and mark
	// the copy as pending approval.
	//
	// This is a status from admin
	// that replies to zork.
	testStatus := new(gtsmodel.Status)
	*testStatus = *suite.testStatuses["admin_account_status_3"]
	testStatus.PendingApproval = util.Ptr(true)
	if err := suite.state.DB.UpdateStatus(ctx, testStatus); err != nil {
		suite.FailNow(err.Error())
	}

	for _, testCase := range []struct {
		acct    *gtsmodel.Account
		visible bool
	}{
		{
			acct:    suite.testAccounts["admin_account"],
			visible: true, // Own status, always visible.
		},
		{
			acct:    suite.testAccounts["local_account_1"],
			visible: true, // Reply to zork, always visible.
		},
		{
			acct:    suite.testAccounts["local_account_2"],
			visible: false, // None of their business.
		},
		{
			acct:    suite.testAccounts["remote_account_1"],
			visible: false, // None of their business.
		},
		{
			acct:    nil,   // Unauthed request.
			visible: false, // None of their business.
		},
	} {
		visible, err := suite.filter.StatusVisible(ctx, testCase.acct, testStatus)
		suite.NoError(err)
		suite.Equal(testCase.visible, visible)
	}

	// Update the status to mark it as approved.
	testStatus.PendingApproval = util.Ptr(false)
	testStatus.ApprovedByURI = "http://localhost:8080/some/accept/uri"
	if err := suite.state.DB.UpdateStatus(ctx, testStatus); err != nil {
		suite.FailNow(err.Error())
	}

	for _, testCase := range []struct {
		acct    *gtsmodel.Account
		visible bool
	}{
		{
			acct:    suite.testAccounts["admin_account"],
			visible: true, // Own status, always visible.
		},
		{
			acct:    suite.testAccounts["local_account_1"],
			visible: true, // Reply to zork, always visible.
		},
		{
			acct:    suite.testAccounts["local_account_2"],
			visible: true, // Should be visible now.
		},
		{
			acct:    suite.testAccounts["remote_account_1"],
			visible: true, // Should be visible now.
		},
		{
			acct:    nil,  // Unauthed request.
			visible: true, // Should be visible now (public status).
		},
	} {
		visible, err := suite.filter.StatusVisible(ctx, testCase.acct, testStatus)
		suite.NoError(err)
		suite.Equal(testCase.visible, visible)
	}
}

func (suite *StatusVisibleTestSuite) TestVisibleLocalOnly() {
	ctx := context.Background()

	// Local-only, Public status.
	testStatus := suite.testStatuses["local_account_2_status_4"]

	for _, testCase := range []struct {
		acct    *gtsmodel.Account
		visible bool
	}{
		{
			acct:    suite.testAccounts["local_account_2"],
			visible: true, // Own status, always visible.
		},
		{
			acct:    nil,
			visible: false, // No auth, should not be visible..
		},
		{
			acct:    suite.testAccounts["local_account_1"],
			visible: true, // Local account, should be visible.
		},
		{
			acct:    suite.testAccounts["remote_account_1"],
			visible: false, // Blocked account, should not be visible.
		},
		{
			acct:    suite.testAccounts["remote_account_2"],
			visible: false, // Remote account, should not be visible.
		},
	} {
		visible, err := suite.filter.StatusVisible(ctx, testCase.acct, testStatus)
		suite.NoError(err)
		suite.Equal(testCase.visible, visible)
	}
}

func TestStatusVisibleTestSuite(t *testing.T) {
	suite.Run(t, new(StatusVisibleTestSuite))
}
