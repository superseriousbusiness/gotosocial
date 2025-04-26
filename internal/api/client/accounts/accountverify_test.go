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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/stretchr/testify/suite"
)

type AccountVerifyTestSuite struct {
	AccountStandardTestSuite
}

// accountVerifyGet calls the verify_credentials API method for a given account fixture name.
// Assumes token and user fixture names are the same as the account fixture name.
func (suite *AccountVerifyTestSuite) accountVerifyGet(fixtureName string) *apimodel.Account {
	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodGet, nil, accounts.VerifyPath, "")

	// override the account that we're authenticated as
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts[fixtureName])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens[fixtureName]))
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers[fixtureName])

	// call the handler
	suite.accountsModule.AccountVerifyGETHandler(ctx)

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

// We should see public account information and profile source for a normal user.
func (suite *AccountVerifyTestSuite) TestAccountVerifyGet() {
	fixtureName := "local_account_1"
	testAccount := suite.testAccounts[fixtureName]

	apimodelAccount := suite.accountVerifyGet(fixtureName)

	createdAt, err := time.Parse(time.RFC3339, apimodelAccount.CreatedAt)
	suite.NoError(err)

	suite.Equal(testAccount.ID, apimodelAccount.ID)
	suite.Equal(testAccount.Username, apimodelAccount.Username)
	suite.Equal(testAccount.Username, apimodelAccount.Acct)
	suite.Equal(testAccount.DisplayName, apimodelAccount.DisplayName)
	suite.Equal(*testAccount.Locked, apimodelAccount.Locked)
	suite.False(apimodelAccount.Bot)
	suite.WithinDuration(testAccount.CreatedAt, createdAt, 30*time.Second) // we lose a bit of accuracy serializing so fuzz this a bit
	suite.Equal(testAccount.URL, apimodelAccount.URL)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg", apimodelAccount.Avatar)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp", apimodelAccount.AvatarStatic)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg", apimodelAccount.Header)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp", apimodelAccount.HeaderStatic)
	suite.Equal(2, apimodelAccount.FollowersCount)
	suite.Equal(2, apimodelAccount.FollowingCount)
	suite.Equal(9, apimodelAccount.StatusesCount)
	suite.EqualValues(apimodel.VisibilityPublic, apimodelAccount.Source.Privacy)
	suite.Equal(testAccount.Settings.Language, apimodelAccount.Source.Language)
	suite.Equal(testAccount.NoteRaw, apimodelAccount.Source.Note)
}

// testAccountVerifyGetRole calls the verify_credentials API method for a given account fixture name,
// and checks the response for permissions appropriate to the role.
func (suite *AccountVerifyTestSuite) testAccountVerifyGetRole(fixtureName string) {
	testUser := suite.testUsers[fixtureName]

	apimodelAccount := suite.accountVerifyGet(fixtureName)

	if suite.NotNil(apimodelAccount.Role) {
		switch {
		case *testUser.Admin:
			suite.Equal("admin", string(apimodelAccount.Role.Name))
			suite.NotZero(apimodelAccount.Role.Permissions)
			suite.True(apimodelAccount.Role.Highlighted)

		case *testUser.Moderator:
			suite.Equal("moderator", string(apimodelAccount.Role.Name))
			suite.Zero(apimodelAccount.Role.Permissions)
			suite.True(apimodelAccount.Role.Highlighted)

		default:
			suite.Equal("user", string(apimodelAccount.Role.Name))
			suite.Zero(apimodelAccount.Role.Permissions)
			suite.False(apimodelAccount.Role.Highlighted)
		}
	}
}

// We should see a role for a normal user, and that role should not have any permissions.
func (suite *AccountVerifyTestSuite) TestAccountVerifyGetRoleUser() {
	suite.testAccountVerifyGetRole("local_account_1")
}

// We should see a role for an admin user, and that role should have some permissions.
func (suite *AccountVerifyTestSuite) TestAccountVerifyGetRoleAdmin() {
	suite.testAccountVerifyGetRole("admin_account")
}

func TestAccountVerifyTestSuite(t *testing.T) {
	suite.Run(t, new(AccountVerifyTestSuite))
}
