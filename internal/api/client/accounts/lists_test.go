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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ListsTestSuite struct {
	AccountStandardTestSuite
}

func (suite *ListsTestSuite) getLists(targetAccountID string, expectedHTTPStatus int, expectedBody string) []*apimodel.List {
	var (
		recorder = httptest.NewRecorder()
		ctx, _   = testrig.CreateGinTestContext(recorder, nil)
		request  = httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/accounts/"+targetAccountID+"/lists", nil)
	)

	// Set up the test context.
	ctx.Request = request
	ctx.AddParam("id", targetAccountID)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// Trigger the handler.
	suite.accountsModule.AccountListsGETHandler(ctx)

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

	// Return list response.
	resp := new([]*apimodel.List)
	if err := json.Unmarshal(b, resp); err != nil {
		suite.FailNow(err.Error())
	}

	return *resp
}

func (suite *ListsTestSuite) TestGetListsHit() {
	targetAccount := suite.testAccounts["admin_account"]
	suite.getLists(targetAccount.ID, http.StatusOK, `[{"id":"01H0G8E4Q2J3FE3JDWJVWEDCD1","title":"Cool Ass Posters From This Instance","replies_policy":"followed","exclusive":false}]`)
}

func (suite *ListsTestSuite) TestGetListsNoHit() {
	targetAccount := suite.testAccounts["remote_account_1"]
	suite.getLists(targetAccount.ID, http.StatusOK, `[]`)
}

func TestListsTestSuite(t *testing.T) {
	suite.Run(t, new(ListsTestSuite))
}
