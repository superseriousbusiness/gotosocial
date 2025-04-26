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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type AccountProfileTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountProfileTestSuite) deleteProfileAttachment(
	testAccountFixtureName string,
	profileSubpath string,
	handler func(*gin.Context),
	expectedHTTPStatus int,
) (*apimodel.Account, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts[testAccountFixtureName])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens[testAccountFixtureName]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers[testAccountFixtureName])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodDelete, config.GetProtocol()+"://"+config.GetHost()+"/api"+accounts.ProfileBasePath+"/"+profileSubpath, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	handler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	// check code
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		return nil, fmt.Errorf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	resp := &apimodel.Account{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// Delete the avatar of a user that has an avatar. Should succeed.
func (suite *AccountProfileTestSuite) TestDeleteAvatar() {
	account, err := suite.deleteProfileAttachment(
		"local_account_1",
		"avatar",
		suite.accountsModule.AccountAvatarDELETEHandler,
		http.StatusOK,
	)
	if suite.NoError(err) {
		// An empty URL is legal *only* in the test environment, which may have no default avatars.
		suite.True(account.Avatar == "" || strings.HasPrefix(account.Avatar, "http://localhost:8080/assets/default_avatars/"))
	}
}

// Delete the avatar of a user that doesn't have an avatar. Should succeed.
func (suite *AccountProfileTestSuite) TestDeleteNonexistentAvatar() {
	account, err := suite.deleteProfileAttachment(
		"admin_account",
		"avatar",
		suite.accountsModule.AccountAvatarDELETEHandler,
		http.StatusOK,
	)
	if suite.NoError(err) {
		// An empty URL is legal *only* in the test environment, which may have no default avatars.
		suite.True(account.Avatar == "" || strings.HasPrefix(account.Avatar, "http://localhost:8080/assets/default_avatars/"))
	}
}

// Delete the header of a user that has a header. Should succeed.
func (suite *AccountProfileTestSuite) TestDeleteHeader() {
	account, err := suite.deleteProfileAttachment(
		"local_account_2",
		"header",
		suite.accountsModule.AccountHeaderDELETEHandler,
		http.StatusOK,
	)
	if suite.NoError(err) {
		suite.Equal("http://localhost:8080/assets/default_header.webp", account.Header)
	}
}

// Delete the header of a user that doesn't have a header. Should succeed.
func (suite *AccountProfileTestSuite) TestDeleteNonexistentHeader() {
	account, err := suite.deleteProfileAttachment(
		"admin_account",
		"header",
		suite.accountsModule.AccountHeaderDELETEHandler,
		http.StatusOK,
	)
	if suite.NoError(err) {
		suite.Equal("http://localhost:8080/assets/default_header.webp", account.Header)
	}
}

func TestAccountProfileTestSuite(t *testing.T) {
	suite.Run(t, new(AccountProfileTestSuite))
}
