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

package search_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/search"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type SearchGetTestSuite struct {
	SearchStandardTestSuite
}

func (suite *SearchGetTestSuite) getSearch(
	requestingAccount *gtsmodel.Account,
	token *gtsmodel.Token,
	apiVersion string,
	user *gtsmodel.User,
	maxID *string,
	minID *string,
	limit *int,
	offset *int,
	query string,
	queryType *string,
	resolve *bool,
	following *bool,
	fromAccountID *string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.SearchResult, error) {
	var (
		recorder   = httptest.NewRecorder()
		ctx, _     = testrig.CreateGinTestContext(recorder, nil)
		requestURL = testrig.URLMustParse("/api" + search.BasePath)
		queryParts []string
	)

	// Put the request together.
	ctx.AddParam(apiutil.APIVersionKey, apiVersion)

	if maxID != nil {
		queryParts = append(queryParts, apiutil.MaxIDKey+"="+url.QueryEscape(*maxID))
	}

	if minID != nil {
		queryParts = append(queryParts, apiutil.MinIDKey+"="+url.QueryEscape(*minID))
	}

	if limit != nil {
		queryParts = append(queryParts, apiutil.LimitKey+"="+strconv.Itoa(*limit))
	}

	if offset != nil {
		queryParts = append(queryParts, apiutil.SearchOffsetKey+"="+strconv.Itoa(*offset))
	}

	queryParts = append(queryParts, apiutil.SearchQueryKey+"="+url.QueryEscape(query))

	if queryType != nil {
		queryParts = append(queryParts, apiutil.SearchTypeKey+"="+url.QueryEscape(*queryType))
	}

	if resolve != nil {
		queryParts = append(queryParts, apiutil.SearchResolveKey+"="+strconv.FormatBool(*resolve))
	}

	if following != nil {
		queryParts = append(queryParts, apiutil.SearchFollowingKey+"="+strconv.FormatBool(*following))
	}

	if fromAccountID != nil {
		queryParts = append(queryParts, apiutil.AccountIDKey+"="+url.QueryEscape(*fromAccountID))
	}

	requestURL.RawQuery = strings.Join(queryParts, "&")
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURL.String(), nil)
	ctx.Set(oauth.SessionAuthorizedAccount, requestingAccount)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, user)

	// Trigger the function being tested.
	suite.searchModule.SearchGETHandler(ctx)

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

	searchResult := &apimodel.SearchResult{}
	if err := json.Unmarshal(b, searchResult); err != nil {
		suite.FailNow(err.Error())
	}

	return searchResult, nil
}

func (suite *SearchGetTestSuite) bodgeLocalInstance(domain string) {
	// Set new host.
	config.SetHost(domain)

	// Copy instance account to not mess up other tests.
	instanceAccount := &gtsmodel.Account{}
	*instanceAccount = *suite.testAccounts["instance_account"]

	// Set username of instance account to given domain.
	instanceAccount.Username = domain
	if err := suite.db.UpdateAccount(context.Background(), instanceAccount, "username"); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByURI() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "https://unknown-instance.com/users/brand_new_person"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestring() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "@brand_new_person@unknown-instance.com"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringUppercase() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "@Some_User@example.org"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringNoLeadingAt() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "brand_new_person@unknown-instance.com"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringNoResolve() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@brand_new_person@unknown-instance.com"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringSpecialChars() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@üser@ëxample.org"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(searchResult.Accounts); l != 1 {
		suite.FailNow("", "expected %d accounts, got %d", 1, l)
	}
	suite.Equal("üser@ëxample.org", searchResult.Accounts[0].Acct)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringSpecialCharsPunycode() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@üser@xn--xample-ova.org"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(searchResult.Accounts); l != 1 {
		suite.FailNow("", "expected %d accounts, got %d", 1, l)
	}
	suite.Equal("üser@ëxample.org", searchResult.Accounts[0].Acct)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByNamestring() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@the_mighty_zork"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByNamestringWithDomain() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@the_mighty_zork@localhost:8080"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchNonexistingLocalAccountByNamestringResolveTrue() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "@somone_made_up@localhost:8080"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByURI() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "http://localhost:8080/users/the_mighty_zork"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByURL() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "http://localhost:8080/@the_mighty_zork"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchNonexistingLocalAccountByURL() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "http://localhost:8080/@the_shmighty_shmork"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
}

func (suite *SearchGetTestSuite) TestSearchStatusByURL() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"
		queryType          *string = func() *string { i := "statuses"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Statuses, 1) {
		suite.FailNow("expected 1 status in search results but got 0")
	}

	gotStatus := searchResult.Statuses[0]
	suite.NotNil(gotStatus)
}

func (suite *SearchGetTestSuite) TestSearchBlockedDomainURL() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "https://replyguys.com/@someone"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchBlockedDomainNamestring() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "@someone@replyguys.com"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchAAny() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "a"
		queryType          *string = nil // Return anything.
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 6)
	suite.Len(searchResult.Statuses, 9)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchAAnyFollowingOnly() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "a"
		queryType          *string = nil // Return anything.
		following          *bool   = func() *bool { i := true; return &i }()
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 2)
	suite.Len(searchResult.Statuses, 9)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchAStatuses() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "a"
		queryType          *string = func() *string { i := "statuses"; return &i }() // Only statuses.
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 9)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchHiStatusesWithAccountIDInQueryParam() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "hi"
		queryType          *string = func() *string { i := "statuses"; return &i }() // Only statuses.
		following          *bool   = nil
		fromAccountID      *string = func() *string { i := suite.testAccounts["local_account_2"].ID; return &i }()
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 1)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchHiStatusesWithAccountIDInQueryText() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "hi from:1happyturtle"
		queryType          *string = func() *string { i := "statuses"; return &i }() // Only statuses.
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 1)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchAAccounts() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "a"
		queryType          *string = func() *string { i := "accounts"; return &i }() // Only accounts.
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 6)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchAccountsLimit1() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = func() *int { i := 1; return &i }()
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "a"
		queryType          *string = func() *string { i := "accounts"; return &i }() // Only accounts.
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 1)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalInstanceAccountByURI() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "http://localhost:8080/users/localhost:8080"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should be able to get instance
	// account by exact URI.
	suite.Len(searchResult.Accounts, 1)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalInstanceAccountFull() {
	// Namestring excludes ':' in usernames, so we
	// need to fiddle with the instance account a
	// bit to get it to look like a different domain.
	newDomain := "example.org"
	suite.bodgeLocalInstance(newDomain)

	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@" + newDomain + "@" + newDomain
		queryType          *string = nil
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should be able to get instance
	// account by full namestring.
	suite.Len(searchResult.Accounts, 1)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalInstanceAccountPartial() {
	// Namestring excludes ':' in usernames, so we
	// need to fiddle with the instance account a
	// bit to get it to look like a different domain.
	newDomain := "example.org"
	suite.bodgeLocalInstance(newDomain)

	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@" + newDomain
		queryType          *string = nil
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Query was a partial namestring from our
	// instance, instance account should be
	// excluded from results.
	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalInstanceAccountEvenMorePartial() {
	// Namestring excludes ':' in usernames, so we
	// need to fiddle with the instance account a
	// bit to get it to look like a different domain.
	newDomain := "example.org"
	suite.bodgeLocalInstance(newDomain)

	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = newDomain
		queryType          *string = nil
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Query was just 'example.org' which doesn't
	// look like a namestring, so search should
	// fall back to text search and therefore give
	// 0 results back.
	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchRemoteInstanceAccountPartial() {
	// Insert an instance account that's not
	// from our instance, and try to search
	// for it with a partial namestring.
	theirDomain := "example.org"

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if err := suite.db.PutAccount(context.Background(), &gtsmodel.Account{
		ID:                    "01H6RWPG8T6DNW6VNXPBCJBH5S",
		Username:              theirDomain,
		Domain:                theirDomain,
		URI:                   "http://" + theirDomain + "/users/" + theirDomain,
		URL:                   "http://" + theirDomain + "/@" + theirDomain,
		PublicKeyURI:          "http://" + theirDomain + "/users/" + theirDomain + "#main-key",
		InboxURI:              "http://" + theirDomain + "/users/" + theirDomain + "/inbox",
		OutboxURI:             "http://" + theirDomain + "/users/" + theirDomain + "/outbox",
		FollowersURI:          "http://" + theirDomain + "/users/" + theirDomain + "/followers",
		FollowingURI:          "http://" + theirDomain + "/users/" + theirDomain + "/following",
		FeaturedCollectionURI: "http://" + theirDomain + "/users/" + theirDomain + "/collections/featured",
		ActorType:             gtsmodel.AccountActorTypePerson,
		PrivateKey:            key,
		PublicKey:             &key.PublicKey,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "@" + theirDomain
		queryType          *string = nil
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Search for instance account from
	// another domain should return 0 results.
	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchBadQueryType() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "whatever"
		queryType          *string = func() *string { i := "aaaaaaaaaaa"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusBadRequest
		expectedBody               = `{"error":"Bad Request: search query type aaaaaaaaaaa was not recognized, valid options are ['', 'accounts', 'statuses', 'hashtags']"}`
	)

	_, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *SearchGetTestSuite) TestSearchEmptyQuery() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = ""
		queryType          *string = func() *string { i := "aaaaaaaaaaa"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusBadRequest
		expectedBody               = `{"error":"Bad Request: required key q was not set or had empty value"}`
	)

	_, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *SearchGetTestSuite) TestSearchHashtagV1() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "#welcome"
		queryType          *string = func() *string { i := "hashtags"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = `{"accounts":[],"statuses":[],"hashtags":[{"name":"welcome","url":"http://localhost:8080/tags/welcome","history":[]}]}`
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 1)
}

func (suite *SearchGetTestSuite) TestSearchHashtagV2() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "#welcome"
		queryType          *string = func() *string { i := "hashtags"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = `{"accounts":[],"statuses":[],"hashtags":["welcome"]}`
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv1,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 1)
}

func (suite *SearchGetTestSuite) TestSearchHashtagButWithAccountSearch() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "#welcome"
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ``
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchNotHashtagButWithTypeHashtag() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = nil
		query                      = "welco"
		queryType          *string = func() *string { i := "hashtags"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ``
	)

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 1)
}

func (suite *SearchGetTestSuite) TestSearchBlockedAccountFullNamestring() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		targetAccount              = suite.testAccounts["remote_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "@" + targetAccount.Username + "@" + targetAccount.Domain
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	// Block the account
	// we're about to search.
	if err := suite.db.PutBlock(
		context.Background(),
		&gtsmodel.Block{
			ID:              id.NewULID(),
			URI:             "https://example.org/nooooooo",
			AccountID:       requestingAccount.ID,
			TargetAccountID: targetAccount.ID,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Search was for full namestring;
	// we should still be able to see
	// the account we've blocked.
	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchBlockedAccountPartialNamestring() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		targetAccount              = suite.testAccounts["remote_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = "@" + targetAccount.Username
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	// Block the account
	// we're about to search.
	if err := suite.db.PutBlock(
		context.Background(),
		&gtsmodel.Block{
			ID:              id.NewULID(),
			URI:             "https://example.org/nooooooo",
			AccountID:       requestingAccount.ID,
			TargetAccountID: targetAccount.ID,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Search was for partial namestring;
	// we should not be able to see
	// the account we've blocked.
	if !suite.Empty(searchResult.Accounts) {
		suite.FailNow("expected 0 accounts in search results")
	}
}

func (suite *SearchGetTestSuite) TestSearchBlockedAccountURI() {
	var (
		requestingAccount          = suite.testAccounts["local_account_1"]
		targetAccount              = suite.testAccounts["remote_account_1"]
		token                      = suite.testTokens["local_account_1"]
		user                       = suite.testUsers["local_account_1"]
		maxID              *string = nil
		minID              *string = nil
		limit              *int    = nil
		offset             *int    = nil
		resolve            *bool   = func() *bool { i := true; return &i }()
		query                      = targetAccount.URI
		queryType          *string = func() *string { i := "accounts"; return &i }()
		following          *bool   = nil
		fromAccountID      *string = nil
		expectedHTTPStatus         = http.StatusOK
		expectedBody               = ""
	)

	// Block the account
	// we're about to search.
	if err := suite.db.PutBlock(
		context.Background(),
		&gtsmodel.Block{
			ID:              id.NewULID(),
			URI:             "https://example.org/nooooooo",
			AccountID:       requestingAccount.ID,
			TargetAccountID: targetAccount.ID,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	searchResult, err := suite.getSearch(
		requestingAccount,
		token,
		apiutil.APIv2,
		user,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		resolve,
		following,
		fromAccountID,
		expectedHTTPStatus,
		expectedBody)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Search was for precise URI;
	// we should still be able to see
	// the account we've blocked.
	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func TestSearchGetTestSuite(t *testing.T) {
	suite.Run(t, &SearchGetTestSuite{})
}
