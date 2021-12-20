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
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AdminTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AdminTestSuite) TestCreateInstanceAccount() {
	// we need to take an empty db for this...
	testrig.StandardDBTeardown(suite.db)
	// ...with tables created but no data
	testrig.CreateTestTables(suite.db)

	// make sure there's no instance account in the db yet
	acct, err := suite.db.GetInstanceAccount(context.Background(), "")
	suite.Error(err)
	suite.Nil(acct)

	// create it
	err = suite.db.CreateInstanceAccount(context.Background())
	suite.NoError(err)

	// and now check it exists
	acct, err = suite.db.GetInstanceAccount(context.Background(), "")
	suite.NoError(err)
	suite.NotNil(acct)
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
