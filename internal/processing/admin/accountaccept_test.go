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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	acct, errWithCode := suite.adminProcessor.AccountApprove(
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

	// Check DB entry too.
	dbUser, err := suite.state.DB.GetUserByID(ctx, targetUser.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(*dbUser.Approved)
}

func TestAdminApproveTestSuite(t *testing.T) {
	suite.Run(t, new(AdminApproveTestSuite))
}
