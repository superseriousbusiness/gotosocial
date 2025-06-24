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

package v1_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	filtersV1 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v1"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

func (suite *FiltersTestSuite) getFilter(
	filterKeywordID string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.FilterV1, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV1.BasePath+"/"+filterKeywordID, nil)
	ctx.Request.Header.Set("accept", "application/json")

	ctx.AddParam("id", filterKeywordID)

	// trigger the handler
	suite.filtersModule.FilterGETHandler(ctx)

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

	resp := &apimodel.FilterV1{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestGetFilter() {
	// v1 filters map to individual filter keywords, but also use the settings of the associated filter.
	expectedFilterGtsModel := suite.testFilters["local_account_1_filter_1"]
	expectedFilterKeywordGtsModel := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"]

	filter, err := suite.getFilter(expectedFilterKeywordGtsModel.ID, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(filter)
	suite.Equal(expectedFilterGtsModel.Action == gtsmodel.FilterActionHide, filter.Irreversible)
	suite.Equal(expectedFilterKeywordGtsModel.ID, filter.ID)
	suite.Equal(expectedFilterKeywordGtsModel.Keyword, filter.Phrase)
}

func (suite *FiltersTestSuite) TestGetAnotherAccountsFilter() {
	id := suite.testFilterKeywords["local_account_2_filter_1_keyword_1"].ID

	_, err := suite.getFilter(id, http.StatusNotFound, `{"error":"Not Found: filter not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestGetNonexistentFilter() {
	id := "not_even_a_real_ULID"

	_, err := suite.getFilter(id, http.StatusNotFound, `{"error":"Not Found: filter keyword not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
