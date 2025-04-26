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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AdminApproveTestSuite struct {
	AdminStandardTestSuite
}

func (suite *AdminApproveTestSuite) TestApprove() {
	var (
		ctx        = context.Background()
		adminAcct  = suite.testAccounts["admin_account"]
		targetAcct = suite.testAccounts["unconfirmed_account"]
		targetUser = new(gtsmodel.User)
	)

	// Copy user since we're modifying it.
	*targetUser = *suite.testUsers["unconfirmed_account"]

	// Approve the sign-up.
	acct, errWithCode := suite.adminProcessor.SignupApprove(
		ctx,
		adminAcct,
		targetAcct.ID,
	)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Account should be approved.
	suite.NotNil(acct)
	suite.True(acct.Approved)
	suite.Nil(acct.IP)

	// Wait for processor to
	// handle side effects.
	var (
		dbUser *gtsmodel.User
		err    error
	)
	if !testrig.WaitFor(func() bool {
		dbUser, err = suite.state.DB.GetUserByID(ctx, targetUser.ID)
		return err == nil && dbUser != nil && *dbUser.Approved
	}) {
		suite.FailNow("waiting for approved user")
	}
}

func TestAdminApproveTestSuite(t *testing.T) {
	suite.Run(t, new(AdminApproveTestSuite))
}
