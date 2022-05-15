/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type BasicTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *BasicTestSuite) TestGetAccountByID() {
	testAccount := suite.testAccounts["local_account_1"]

	a := &gtsmodel.Account{}
	err := suite.db.GetByID(context.Background(), testAccount.ID, a)
	suite.NoError(err)
}

func (suite *BasicTestSuite) TestGetAllStatuses() {
	s := []*gtsmodel.Status{}
	err := suite.db.GetAll(context.Background(), &s)
	suite.NoError(err)
	suite.Len(s, 16)
}

func (suite *BasicTestSuite) TestGetAllNotNull() {
	where := []db.Where{{
		Key:   "domain",
		Value: nil,
		Not:   true,
	}}

	a := []*gtsmodel.Account{}

	err := suite.db.GetWhere(context.Background(), where, &a)
	suite.NoError(err)
	suite.NotEmpty(a)

	for _, acct := range a {
		suite.NotEmpty(acct.Domain)
	}
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
