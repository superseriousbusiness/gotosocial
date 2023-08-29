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
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type DomainBlockTestSuite struct {
	AdminStandardTestSuite
}

func (suite *DomainBlockTestSuite) TestCreateDomainBlock() {
	var (
		ctx            = context.Background()
		adminAcct      = suite.testAccounts["admin_account"]
		domain         = "fossbros-anonymous.io"
		obfuscate      = false
		publicComment  = ""
		privateComment = ""
		subscriptionID = ""
	)

	apiBlock, actionID, errWithCode := suite.adminProcessor.DomainBlockCreate(
		ctx,
		adminAcct,
		domain,
		obfuscate,
		publicComment,
		privateComment,
		subscriptionID,
	)
	suite.NoError(errWithCode)
	suite.NotNil(apiBlock)
	suite.NotEmpty(actionID)

	// Wait for action to finish.
	if !testrig.WaitFor(func() bool {
		return suite.adminProcessor.Actions.TotalRunning() == 0
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
}

func TestDomainBlockTestSuite(t *testing.T) {
	suite.Run(t, new(DomainBlockTestSuite))
}
