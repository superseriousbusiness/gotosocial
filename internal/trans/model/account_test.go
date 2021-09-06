/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package trans_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	trans "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

type AccountTestSuite struct {
	ModelTestSuite
}

func (suite *AccountTestSuite) TestAccountsIdempotent() {
	// we should be able to get all accounts with the simple trans.Account struct
	accounts := []*trans.Account{}
	err := suite.db.GetAll(context.Background(), &accounts)
	suite.NoError(err)
	suite.NotEmpty(accounts)

	// we should be able to marshal the accounts to json with no problems
	b, err := json.Marshal(&accounts)
	suite.NoError(err)
	suite.NotNil(b)
	suite.T().Log(string(b))

	// the json should be idempotent
	mAccounts := []*trans.Account{}
	err = json.Unmarshal(b, &mAccounts)
	suite.NoError(err)
	suite.NotEmpty(mAccounts)
	suite.EqualValues(accounts, mAccounts)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, &AccountTestSuite{})
}
