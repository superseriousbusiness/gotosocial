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
	"net/url"
	"strconv"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AccountSearchTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountSearchTestSuite) getSearch(
	requestingAccount *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	limit *int,
	offset *int,
	query string,
	resolve *bool,
	following *bool,
	expectedHTTPStatus int,
	expectedBody string,
) ([]*apimodel.Account, error) {
	var (
		recorder   = httptest.NewRecorder()
		ctx, _     = testrig.CreateGinTestContext(recorder, nil)
		requestURL = testrig.URLMustParse("/api" + accounts.BasePath + "/search")
		queryParts []string
	)

	// Put the request together.
	if limit != nil {
		queryParts = append(queryParts, apiutil.LimitKey+"="+strconv.Itoa(*limit))
	}

	if offset != nil {
		queryParts = append(queryParts, apiutil.SearchOffsetKey+"="+strconv.Itoa(*offset))
	}

	queryParts = append(queryParts, apiutil.SearchQueryKey+"="+url.QueryEscape(query))

	if resolve != nil {
		queryParts = append(queryParts, apiutil.SearchResolveKey+"="+strconv.FormatBool(*resolve))
	}

	if following != nil {
		queryParts = append(queryParts, apiutil.SearchFollowingKey+"="+strconv.FormatBool(*following))
	}

	requestURL.RawQuery = strings.Join(queryParts, "&")
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURL.String(), nil)
	ctx.Set(oauth.SessionAuthorizedAccount, requestingAccount)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, user)

	// Trigger the function being tested.
	suite.accountsModule.AccountSearchGETHandler(ctx)

	// Read the result.
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	errs := gtserror.NewMultiError(2)

	// Check expected code + body.
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// If we got an expected body, return early.
	if expectedBody != "" && string(b) != expectedBody {
		errs.Appendf("expected %s got %s", expectedBody, string(b))
	}

	if err := errs.Combine(); err != nil {
		suite.FailNow("", "%v (body %s)", err, string(b))
	}

	accounts := []*apimodel.Account{}
	if err := json.Unmarshal(b, &accounts); err != nil {
		suite.FailNow(err.Error())
	}

	return accounts, nil
}

func (suite *AccountSearchTestSuite) TestSearchZorkOK() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "zork"
		following          *bool = nil
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 1 {
		suite.FailNow("", "expected length %d got %d", 1, l)
	}
}

func (suite *AccountSearchTestSuite) TestSearchZorkExactOK() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "@the_mighty_zork"
		following          *bool = nil
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 1 {
		suite.FailNow("", "expected length %d got %d", 1, l)
	}
}

func (suite *AccountSearchTestSuite) TestSearchZorkWithDomainOK() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "@the_mighty_zork@localhost:8080"
		following          *bool = nil
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 1 {
		suite.FailNow("", "expected length %d got %d", 1, l)
	}
}

func (suite *AccountSearchTestSuite) TestSearchFossSatanNotFollowing() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "foss_satan"
		following          *bool = func() *bool { i := false; return &i }()
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 1 {
		suite.FailNow("", "expected length %d got %d", 1, l)
	}
}

func (suite *AccountSearchTestSuite) TestSearchFossSatanFollowing() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "foss_satan"
		following          *bool = func() *bool { i := true; return &i }()
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 0 {
		suite.FailNow("", "expected length %d got %d", 0, l)
	}
}

func (suite *AccountSearchTestSuite) TestSearchBonkersQuery() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "aaaaa@aaaaaaaaa@aaaaa **** this won't@ return anything!@!!"
		following          *bool = nil
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 0 {
		suite.FailNow("", "expected length %d got %d", 0, l)
	}
}

func (suite *AccountSearchTestSuite) TestSearchAFollowing() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "a"
		following          *bool = nil
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 6 {
		suite.FailNow("", "expected length %d got %d", 6, l)
	}

	usernames := make([]string, 0, 6)
	for _, account := range accounts {
		usernames = append(usernames, account.Username)
	}

	suite.EqualValues([]string{"her_fuckin_maj", "media_mogul", "foss_satan", "1happyturtle", "the_mighty_zork", "admin"}, usernames)
}

func (suite *AccountSearchTestSuite) TestSearchANotFollowing() {
	var (
		requestingAccount        = suite.testAccounts["local_account_1"]
		token                    = suite.testTokens["local_account_1"]
		user                     = suite.testUsers["local_account_1"]
		limit              *int  = nil
		offset             *int  = nil
		resolve            *bool = nil
		query                    = "a"
		following          *bool = func() *bool { i := true; return &i }()
		expectedHTTPStatus       = http.StatusOK
		expectedBody             = ""
	)

	accounts, err := suite.getSearch(
		requestingAccount,
		token,
		user,
		limit,
		offset,
		query,
		resolve,
		following,
		expectedHTTPStatus,
		expectedBody,
	)

	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(accounts); l != 2 {
		suite.FailNow("", "expected length %d got %d", 2, l)
	}

	usernames := make([]string, 0, 2)
	for _, account := range accounts {
		usernames = append(usernames, account.Username)
	}

	suite.EqualValues([]string{"1happyturtle", "admin"}, usernames)
}

func TestAccountSearchTestSuite(t *testing.T) {
	suite.Run(t, new(AccountSearchTestSuite))
}
