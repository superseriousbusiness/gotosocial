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

	filtersV2 "github.com/superseriousbusiness/gotosocial/internal/api/client/filters/v2"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func (suite *FiltersTestSuite) putFilterKeyword(
	filterKeywordID string,
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
	ctx.Request = httptest.NewRequest(http.MethodPut, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.KeywordPath+"/"+filterKeywordID, nil)
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

	ctx.AddParam("id", filterKeywordID)

	// trigger the handler
	suite.filtersModule.FilterKeywordPUTHandler(ctx)

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

func (suite *FiltersTestSuite) TestPutFilterKeywordFull() {
	filterKeywordID := suite.testFilters["local_account_1_filter_1_keyword_1"].ID
	keyword := "fnords"
	wholeWord := true
	filterKeyword, err := suite.putFilterKeyword(filterKeywordID, &keyword, &wholeWord, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(keyword, filterKeyword.Keyword)
	suite.Equal(wholeWord, filterKeyword.WholeWord)
}

func (suite *FiltersTestSuite) TestPutFilterKeywordFullJSON() {
	filterKeywordID := suite.testFilters["local_account_1_filter_1_keyword_1"].ID
	requestJson := `{
		"keyword": "fnords",
		"whole_word": true
	}`
	filterKeyword, err := suite.putFilterKeyword(filterKeywordID, nil, nil, &requestJson, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("fnords", filterKeyword.Keyword)
	suite.True(filterKeyword.WholeWord)
}

func (suite *FiltersTestSuite) TestPutFilterKeywordMinimal() {
	filterKeywordID := suite.testFilters["local_account_1_filter_1_keyword_1"].ID
	keyword := "fnords"
	filterKeyword, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(keyword, filterKeyword.Keyword)
	suite.False(filterKeyword.WholeWord)
}

func (suite *FiltersTestSuite) TestPutFilterKeywordEmptyKeyword() {
	filterKeywordID := suite.testFilters["local_account_1_filter_1_keyword_1"].ID
	keyword := ""
	_, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter keyword must be provided, and must be no more than 40 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterKeywordMissingKeyword() {
	filterKeywordID := suite.testFilters["local_account_1_filter_1_keyword_1"].ID
	_, err := suite.putFilterKeyword(filterKeywordID, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter keyword must be provided, and must be no more than 40 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Changing our filter keyword to the same keyword as another filter keyword in the same filter should fail.
func (suite *FiltersTestSuite) TestPutFilterKeywordKeywordConflict() {
	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_2_keyword_1"].ID
	conflictingKeyword := suite.testFilterKeywords["local_account_1_filter_2_keyword_2"].Keyword
	_, err := suite.putFilterKeyword(filterKeywordID, &conflictingKeyword, nil, nil, http.StatusConflict, `{"error":"Conflict: duplicate keyword"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterKeywordAnotherAccountsFilterKeyword() {
	filterKeywordID := suite.testFilters["local_account_2_filter_1_keyword_1"].ID
	keyword := "fnord"
	_, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusNotFound, `{"error":"Not Found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterKeywordNonexistentFilterKeyword() {
	filterKeywordID := "not_even_a_real_ULID"
	keyword := "fnord"
	_, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusNotFound, `{"error":"Not Found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
