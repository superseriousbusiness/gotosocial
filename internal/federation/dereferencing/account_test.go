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

package dereferencing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *AccountTestSuite) TestDereferenceGroup() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	groupURL := testrig.URLMustParse("https://unknown-instance.com/groups/some_group")
	group, err := suite.dereferencer.GetRemoteAccount(context.Background(), fetchingAccount.Username, groupURL, false, false)
	suite.NoError(err)
	suite.NotNil(group)
	suite.NotNil(group)

	// group values should be set
	suite.Equal("https://unknown-instance.com/groups/some_group", group.URI)
	suite.Equal("https://unknown-instance.com/@some_group", group.URL)

	// group should be in the database
	dbGroup, err := suite.db.GetAccountByURI(context.Background(), group.URI)
	suite.NoError(err)
	suite.Equal(group.ID, dbGroup.ID)
	suite.Equal(ap.ActorGroup, dbGroup.ActorType)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
