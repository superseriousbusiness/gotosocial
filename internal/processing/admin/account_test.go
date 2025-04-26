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

package admin_test

import (
	"context"
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AccountTestSuite struct {
	AdminStandardTestSuite
}

func (suite *AccountTestSuite) TestAccountActionSuspend() {
	var (
		ctx       = context.Background()
		adminAcct = suite.testAccounts["admin_account"]
		request   = &apimodel.AdminActionRequest{
			Category: gtsmodel.AdminActionCategoryAccount.String(),
			Type:     gtsmodel.AdminActionSuspend.String(),
			Text:     "stinky",
			TargetID: suite.testAccounts["local_account_1"].ID,
		}
	)

	actionID, errWithCode := suite.adminProcessor.AccountAction(
		ctx,
		adminAcct,
		request,
	)
	suite.NoError(errWithCode)
	suite.NotEmpty(actionID)

	// Wait for action to finish.
	if !testrig.WaitFor(func() bool {
		return suite.state.AdminActions.TotalRunning() == 0
	}) {
		suite.FailNow("timed out waiting for admin action(s) to finish")
	}

	// Ensure action marked as
	// completed in the database.
	adminAction, err := suite.db.GetAdminAction(ctx, actionID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotZero(adminAction.CompletedAt)
	suite.Empty(adminAction.Errors)

	// Ensure target account suspended.
	targetAcct, err := suite.db.GetAccountByID(ctx, request.TargetID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotZero(targetAcct.SuspendedAt)
}

func (suite *AccountTestSuite) TestAccountActionUnsupported() {
	var (
		ctx       = context.Background()
		adminAcct = suite.testAccounts["admin_account"]
		request   = &apimodel.AdminActionRequest{
			Category: gtsmodel.AdminActionCategoryAccount.String(),
			Type:     "pee pee poo poo",
			Text:     "stinky",
			TargetID: suite.testAccounts["local_account_1"].ID,
		}
	)

	actionID, errWithCode := suite.adminProcessor.AccountAction(
		ctx,
		adminAcct,
		request,
	)
	suite.EqualError(errWithCode, "admin action type pee pee poo poo is not supported for this endpoint, currently supported types are: [\"suspend\"]")
	suite.Empty(actionID)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
