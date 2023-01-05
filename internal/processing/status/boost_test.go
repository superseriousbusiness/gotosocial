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

package status_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StatusBoostTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusBoostTestSuite) TestBoostOfBoost() {
	ctx := context.Background()

	// first boost a status, no big deal
	boostingAccount1 := suite.testAccounts["local_account_1"]
	application1 := suite.testApplications["application_1"]
	targetStatus1 := suite.testStatuses["admin_account_status_1"]

	boost1, err := suite.status.Boost(ctx, boostingAccount1, application1, targetStatus1.ID)
	suite.NoError(err)
	suite.NotNil(boost1)
	suite.Equal(targetStatus1.ID, boost1.Reblog.ID)

	// now take another account and boost that boost
	boostingAccount2 := suite.testAccounts["local_account_2"]
	application2 := suite.testApplications["application_2"]
	targetStatus2ID := boost1.ID

	boost2, err := suite.status.Boost(ctx, boostingAccount2, application2, targetStatus2ID)
	suite.NoError(err)
	suite.NotNil(boost2)
	// the boosted status should not be the boost,
	// but the original status that was boosted
	suite.Equal(targetStatus1.ID, boost2.Reblog.ID)
}

func TestStatusBoostTestSuite(t *testing.T) {
	suite.Run(t, new(StatusBoostTestSuite))
}
