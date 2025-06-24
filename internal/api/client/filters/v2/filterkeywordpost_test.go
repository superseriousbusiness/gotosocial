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
	"net/url"
	"strconv"
	"strings"

	filtersV2 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v2"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/stream"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

func (suite *FiltersTestSuite) postFilterKeyword(
	filterID string,
	keyword *string,
	wholeWord *bool,
	requestJson *string,
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
	ctx.Request = httptest.NewRequest(http.MethodPost, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.BasePath+"/"+filterID+"/keywords", nil)
	ctx.Request.Header.Set("accept", "application/json")
	if requestJson != nil {
		ctx.Request.Header.Set("content-type", "application/json")
		ctx.Request.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		ctx.Request.Form = make(url.Values)
		if keyword != nil {
			ctx.Request.Form["keyword"] = []string{*keyword}
		}
		if wholeWord != nil {
			ctx.Request.Form["whole_word"] = []string{strconv.FormatBool(*wholeWord)}
		}
	}

	ctx.AddParam("id", filterID)

	// trigger the handler
	suite.filtersModule.FilterKeywordPOSTHandler(ctx)

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

func (suite *FiltersTestSuite) TestPostFilterKeywordFull() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterID := suite.testFilters["local_account_1_filter_1"].ID
	keyword := "fnords"
	wholeWord := true
	filterKeyword, err := suite.postFilterKeyword(filterID, &keyword, &wholeWord, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(keyword, filterKeyword.Keyword)
	suite.Equal(wholeWord, filterKeyword.WholeWord)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterKeywordFullJSON() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterID := suite.testFilters["local_account_1_filter_1"].ID
	requestJson := `{
		"keyword": "fnords",
		"whole_word": true
	}`
	filterKeyword, err := suite.postFilterKeyword(filterID, nil, nil, &requestJson, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("fnords", filterKeyword.Keyword)
	suite.True(filterKeyword.WholeWord)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterKeywordMinimal() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterID := suite.testFilters["local_account_1_filter_1"].ID
	keyword := "fnords"
	filterKeyword, err := suite.postFilterKeyword(filterID, &keyword, nil, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(keyword, filterKeyword.Keyword)
	suite.False(filterKeyword.WholeWord)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterKeywordEmptyKeyword() {
	filterID := suite.testFilters["local_account_1_filter_1"].ID
	keyword := ""
	_, err := suite.postFilterKeyword(filterID, &keyword, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter keyword must be provided, and must be no more than 40 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterKeywordMissingKeyword() {
	filterID := suite.testFilters["local_account_1_filter_1"].ID
	_, err := suite.postFilterKeyword(filterID, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter keyword must be provided, and must be no more than 40 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Creating another filter keyword in the same filter with the same keyword should fail.
func (suite *FiltersTestSuite) TestPostFilterKeywordKeywordConflict() {
	filterID := suite.testFilters["local_account_1_filter_1"].ID
	keyword := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"].Keyword
	_, err := suite.postFilterKeyword(filterID, &keyword, nil, nil, http.StatusConflict, `{"error":"Conflict: duplicate keyword"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterKeywordAnotherAccountsFilter() {
	filterID := suite.testFilters["local_account_2_filter_1"].ID
	keyword := "fnords"
	_, err := suite.postFilterKeyword(filterID, &keyword, nil, nil, http.StatusNotFound, `{"error":"Not Found: filter not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterKeywordNonexistentFilter() {
	filterID := "not_even_a_real_ULID"
	keyword := "fnords"
	_, err := suite.postFilterKeyword(filterID, &keyword, nil, nil, http.StatusNotFound, `{"error":"Not Found: filter not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
