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

package accounts_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/accounts"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type AccountGetTestSuite struct {
	AccountStandardTestSuite
}

// accountVerifyGet calls the get account API method for a given account fixture name.
func (suite *AccountGetTestSuite) getAccount(id string) *apimodel.Account {
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodGet, nil, accounts.BasePath+"/"+id, "")
	ctx.Params = gin.Params{
		gin.Param{
			Key:   accounts.IDKey,
			Value: id,
		},
	}

	// call the handler
	suite.accountsModule.AccountGETHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return apimodelAccount
}

// Fetching the currently logged-in account shows extra info,
// so we should see permissions, but this account is a regular user and should have no display role.
func (suite *AccountGetTestSuite) TestGetDisplayRoleForSelf() {
	apimodelAccount := suite.getAccount(suite.testAccounts["local_account_1"].ID)

	// Role should be set, but permissions should be empty.
	if suite.NotNil(apimodelAccount.Role) {
		role := apimodelAccount.Role
		suite.Equal("user", string(role.Name))
		suite.Zero(role.Permissions)
	}

	// Roles should not have anything in it.
	suite.Empty(apimodelAccount.Roles)
}

// We should not see a display role for an ordinary local account.
func (suite *AccountGetTestSuite) TestGetDisplayRoleForUserAccount() {
	apimodelAccount := suite.getAccount(suite.testAccounts["local_account_2"].ID)

	// Role should not be set.
	suite.Nil(apimodelAccount.Role)

	// Roles should not have anything in it.
	suite.Empty(apimodelAccount.Roles)
}

// We should be able to get a display role for an admin account.
func (suite *AccountGetTestSuite) TestGetDisplayRoleForAdminAccount() {
	apimodelAccount := suite.getAccount(suite.testAccounts["admin_account"].ID)

	// Role should not be set.
	suite.Nil(apimodelAccount.Role)

	// Roles should have exactly one display role.
	if suite.Len(apimodelAccount.Roles, 1) {
		role := apimodelAccount.Roles[0]
		suite.Equal("admin", string(role.Name))
		suite.NotEmpty(role.ID)
	}
}

func TestAccountGetTestSuite(t *testing.T) {
	suite.Run(t, new(AccountGetTestSuite))
}
