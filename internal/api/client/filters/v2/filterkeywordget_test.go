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

package v2_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	filtersV2 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v2"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

func (suite *FiltersTestSuite) getFilterKeyword(
	filterKeywordID string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.FilterKeyword, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.KeywordPath+"/"+filterKeywordID, nil)
	ctx.Request.Header.Set("accept", "application/json")

	ctx.AddParam("id", filterKeywordID)

	// trigger the handler
	suite.filtersModule.FilterKeywordGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
		if expectedBody == "" {
			return nil, errs.Combine()
		}
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.FilterKeyword{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestGetFilterKeyword() {
	expectedFilterKeyword := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"]

	filterKeyword, err := suite.getFilterKeyword(expectedFilterKeyword.ID, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(filterKeyword)
	suite.Equal(expectedFilterKeyword.ID, filterKeyword.ID)
	suite.Equal(expectedFilterKeyword.Keyword, filterKeyword.Keyword)
	suite.Equal(util.PtrOrValue(expectedFilterKeyword.WholeWord, false), filterKeyword.WholeWord)
}

func (suite *FiltersTestSuite) TestGetAnotherAccountsFilterKeyword() {
	id := suite.testFilterKeywords["local_account_2_filter_1_keyword_1"].ID

	_, err := suite.getFilterKeyword(id, http.StatusNotFound, `{"error":"Not Found: filter not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestGetNonexistentFilterKeyword() {
	id := "not_even_a_real_ULID"

	_, err := suite.getFilterKeyword(id, http.StatusNotFound, `{"error":"Not Found: filter keyword not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
