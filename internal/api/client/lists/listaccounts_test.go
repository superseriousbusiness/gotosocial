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

package lists_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/lists"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ListAccountsTestSuite struct {
	ListsStandardTestSuite
}

func (suite *ListAccountsTestSuite) getListAccounts(
	expectedHTTPStatus int,
	expectedBody string,
	listID string,
	maxID string,
	sinceID string,
	minID string,
	limit *int,
) (
	[]*apimodel.Account,
	string, // Link header
	error,
) {

	var (
		recorder = httptest.NewRecorder()
		ctx, _   = testrig.CreateGinTestContext(recorder, nil)
	)

	// Prepare test context.
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// Inject path parameters.
	ctx.AddParam("id", listID)

	// Inject query parameters.
	requestPath := config.GetProtocol() + "://" + config.GetHost() + "/api/" + lists.BasePath + "/" + listID + "/accounts"

	if limit != nil {
		requestPath += "?limit=" + strconv.Itoa(*limit)
	} else {
		requestPath += "?limit=40"
	}
	if maxID != "" {
		requestPath += "&" + apiutil.MaxIDKey + "=" + maxID
	}
	if sinceID != "" {
		requestPath += "&" + apiutil.SinceIDKey + "=" + sinceID
	}
	if minID != "" {
		requestPath += "&" + apiutil.MinIDKey + "=" + minID
	}

	// Prepare test context request.
	request := httptest.NewRequest(http.MethodGet, requestPath, nil)
	request.Header.Set("accept", "application/json")
	ctx.Request = request

	// trigger the handler
	suite.listsModule.ListAccountsGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, "", errs.Combine()
	}

	resp := []*apimodel.Account{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, "", err
	}

	return resp, result.Header.Get("Link"), nil
}

func (suite *ListAccountsTestSuite) TestGetListAccountsPaginatedDefaultLimit() {
	var (
		expectedHTTPStatus      = 200
		expectedBody            = ""
		listID                  = suite.testLists["local_account_1_list_1"].ID
		maxID                   = ""
		minID                   = ""
		sinceID                 = ""
		limit              *int = nil
	)

	accounts, link, err := suite.getListAccounts(
		expectedHTTPStatus,
		expectedBody,
		listID,
		maxID,
		sinceID,
		minID,
		limit,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 2)
	suite.Equal(
		"<http://localhost:8080/api/v1/lists/01H0G8E4Q2J3FE3JDWJVWEDCD1/accounts?limit=40&max_id=01F8MH5NBDF2MV7CTC4Q5128HF>; rel=\"next\", <http://localhost:8080/api/v1/lists/01H0G8E4Q2J3FE3JDWJVWEDCD1/accounts?limit=40&min_id=01F8MH17FWEB39HZJ76B6VXSKF>; rel=\"prev\"",
		link,
	)
}

func (suite *ListAccountsTestSuite) TestGetListAccountsPaginatedNextPage() {
	var (
		expectedHTTPStatus      = 200
		expectedBody            = ""
		listID                  = suite.testLists["local_account_1_list_1"].ID
		maxID                   = ""
		minID                   = ""
		sinceID                 = ""
		limit              *int = func() *int { l := 1; return &l }() // Set to 1.
	)

	// First response, ask for 1 account.
	accounts, link, err := suite.getListAccounts(
		expectedHTTPStatus,
		expectedBody,
		listID,
		maxID,
		sinceID,
		minID,
		limit,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 1)
	suite.Equal(
		"<http://localhost:8080/api/v1/lists/01H0G8E4Q2J3FE3JDWJVWEDCD1/accounts?limit=1&max_id=01F8MH17FWEB39HZJ76B6VXSKF>; rel=\"next\", <http://localhost:8080/api/v1/lists/01H0G8E4Q2J3FE3JDWJVWEDCD1/accounts?limit=1&min_id=01F8MH17FWEB39HZJ76B6VXSKF>; rel=\"prev\"",
		link,
	)

	// Next response, ask for next 1 account.
	maxID = "01H0G8FFM1AGQDRNGBGGX8CYJQ"
	accounts, link, err = suite.getListAccounts(
		expectedHTTPStatus,
		expectedBody,
		listID,
		maxID,
		sinceID,
		minID,
		limit,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 1)
	suite.Equal(
		"<http://localhost:8080/api/v1/lists/01H0G8E4Q2J3FE3JDWJVWEDCD1/accounts?limit=1&max_id=01F8MH17FWEB39HZJ76B6VXSKF>; rel=\"next\", <http://localhost:8080/api/v1/lists/01H0G8E4Q2J3FE3JDWJVWEDCD1/accounts?limit=1&min_id=01F8MH17FWEB39HZJ76B6VXSKF>; rel=\"prev\"",
		link,
	)
}

func (suite *ListAccountsTestSuite) TestGetListAccountsUnpaginated() {
	var (
		expectedHTTPStatus      = 200
		expectedBody            = ""
		listID                  = suite.testLists["local_account_1_list_1"].ID
		maxID                   = ""
		minID                   = ""
		sinceID                 = ""
		limit              *int = func() *int { l := 0; return &l }() // Set to 0 explicitly.
	)

	accounts, link, err := suite.getListAccounts(
		expectedHTTPStatus,
		expectedBody,
		listID,
		maxID,
		sinceID,
		minID,
		limit,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 2)
	suite.Empty(link)
}

func TestListAccountsTestSuite(t *testing.T) {
	suite.Run(t, new(ListAccountsTestSuite))
}
