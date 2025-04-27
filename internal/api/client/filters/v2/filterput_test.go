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
	"slices"
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

func (suite *FiltersTestSuite) putFilter(filterID string, title *string, context *[]string, action *string, expiresIn *int, expiresInStr *string, keywordsAttributesKeyword *[]string, keywordsAttributesWholeWord *[]bool, keywordsAttributesDestroy *[]bool, statusesAttributesID *[]string, statusesAttributesStatusID *[]string, statusesAttributesDestroy *[]bool, requestJson *string, expectedHTTPStatus int, expectedBody string, keywordsAttributesID *[]string) (*apimodel.FilterV2, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodPut, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.BasePath+"/"+filterID, nil)
	ctx.Request.Header.Set("accept", "application/json")
	if requestJson != nil {
		ctx.Request.Header.Set("content-type", "application/json")
		ctx.Request.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		ctx.Request.Form = make(url.Values)
		if title != nil {
			ctx.Request.Form["title"] = []string{*title}
		}
		if context != nil {
			ctx.Request.Form["context[]"] = *context
		}
		if action != nil {
			ctx.Request.Form["filter_action"] = []string{*action}
		}
		if expiresIn != nil {
			ctx.Request.Form["expires_in"] = []string{strconv.Itoa(*expiresIn)}
		} else if expiresInStr != nil {
			ctx.Request.Form["expires_in"] = []string{*expiresInStr}
		}
		if keywordsAttributesID != nil {
			ctx.Request.Form["keywords_attributes[][id]"] = *keywordsAttributesID
		}
		if keywordsAttributesKeyword != nil {
			ctx.Request.Form["keywords_attributes[][keyword]"] = *keywordsAttributesKeyword
		}
		if keywordsAttributesWholeWord != nil {
			formatted := []string{}
			for _, value := range *keywordsAttributesWholeWord {
				formatted = append(formatted, strconv.FormatBool(value))
			}
			ctx.Request.Form["keywords_attributes[][whole_word]"] = formatted
		}
		if keywordsAttributesWholeWord != nil {
			formatted := []string{}
			for _, value := range *keywordsAttributesDestroy {
				formatted = append(formatted, strconv.FormatBool(value))
			}
			ctx.Request.Form["keywords_attributes[][_destroy]"] = formatted
		}
		if statusesAttributesID != nil {
			ctx.Request.Form["statuses_attributes[][id]"] = *statusesAttributesID
		}
		if statusesAttributesStatusID != nil {
			ctx.Request.Form["statuses_attributes[][status_id]"] = *statusesAttributesStatusID
		}
		if statusesAttributesDestroy != nil {
			formatted := []string{}
			for _, value := range *statusesAttributesDestroy {
				formatted = append(formatted, strconv.FormatBool(value))
			}
			ctx.Request.Form["statuses_attributes[][_destroy]"] = formatted
		}
	}

	ctx.AddParam("id", filterID)

	// trigger the handler
	suite.filtersModule.FilterPUTHandler(ctx)

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

	resp := &apimodel.FilterV2{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestPutFilterFull() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	id := suite.testFilters["local_account_1_filter_2"].ID
	title := "messy synoptic varblabbles"
	context := []string{"home", "public"}
	action := "hide"
	expiresIn := 86400
	// Tests attributes arrays that aren't the same length, just in case.
	keywordsAttributesID := []string{
		suite.testFilterKeywords["local_account_1_filter_2_keyword_1"].ID,
		suite.testFilterKeywords["local_account_1_filter_2_keyword_2"].ID,
	}
	keywordsAttributesKeyword := []string{"fū", "", "blah"}
	// If using the form version of this API, you have to always set whole_word to the previous value for that keyword;
	// there's no way to represent a nullable boolean in it.
	keywordsAttributesWholeWord := []bool{true, false, true}
	keywordsAttributesDestroy := []bool{false, true}
	statusesAttributesStatusID := []string{suite.testStatuses["remote_account_1_status_2"].ID}
	filter, err := suite.putFilter(id, &title, &context, &action, &expiresIn, nil, &keywordsAttributesKeyword, &keywordsAttributesWholeWord, &keywordsAttributesDestroy, nil, &statusesAttributesStatusID, nil, nil, http.StatusOK, "", &keywordsAttributesID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(title, filter.Title)
	filterContext := make([]string, 0, len(filter.Context))
	for _, c := range filter.Context {
		filterContext = append(filterContext, string(c))
	}
	suite.ElementsMatch(context, filterContext)
	suite.Equal(apimodel.FilterActionHide, filter.FilterAction)
	if suite.NotNil(filter.ExpiresAt) {
		suite.NotEmpty(*filter.ExpiresAt)
	}

	if suite.Len(filter.Keywords, 3) {
		slices.SortFunc(filter.Keywords, func(lhs, rhs apimodel.FilterKeyword) int {
			return strings.Compare(lhs.ID, rhs.ID)
		})

		suite.Equal("fū", filter.Keywords[0].Keyword)
		suite.True(filter.Keywords[0].WholeWord)

		suite.Equal("quux", filter.Keywords[1].Keyword)
		suite.True(filter.Keywords[1].WholeWord)

		suite.Equal("blah", filter.Keywords[2].Keyword)
		suite.True(filter.Keywords[1].WholeWord)
	}

	if suite.Len(filter.Statuses, 1) {
		slices.SortFunc(filter.Statuses, func(lhs, rhs apimodel.FilterStatus) int {
			return strings.Compare(lhs.ID, rhs.ID)
		})

		suite.Equal(suite.testStatuses["remote_account_1_status_2"].ID, filter.Statuses[0].StatusID)
	}

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPutFilterFullJSON() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	id := suite.testFilters["local_account_1_filter_2"].ID
	// Use a numeric literal with a fractional part to test the JSON-specific handling for non-integer "expires_in".
	requestJson := `{
		"title": "messy synoptic varblabbles",
		"context": ["home", "public"],
		"filter_action": "hide",
		"expires_in": 86400.1,
		"keywords_attributes": [
			{
				"id": "01HN277Y11ENG4EC1ERMAC9FH4",
				"keyword": "fū"
			},
			{
				"id": "01HN278494N88BA2FY4DZ5JTNS",
				"_destroy": true
			},
			{
				"keyword": "blah",
				"whole_word": true
			}
		],
		"statuses_attributes": [
			{
				"status_id": "01HEN2QRFA8H3C6QPN7RD4KSR6"
			}
		]
	}`
	filter, err := suite.putFilter(id, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &requestJson, http.StatusOK, "", nil)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("messy synoptic varblabbles", filter.Title)
	suite.ElementsMatch(
		[]apimodel.FilterContext{
			apimodel.FilterContextHome,
			apimodel.FilterContextPublic,
		},
		filter.Context,
	)
	suite.Equal(apimodel.FilterActionHide, filter.FilterAction)
	if suite.NotNil(filter.ExpiresAt) {
		suite.NotEmpty(*filter.ExpiresAt)
	}

	if suite.Len(filter.Keywords, 3) {
		slices.SortFunc(filter.Keywords, func(lhs, rhs apimodel.FilterKeyword) int {
			return strings.Compare(lhs.ID, rhs.ID)
		})

		suite.Equal("fū", filter.Keywords[0].Keyword)
		suite.True(filter.Keywords[0].WholeWord)

		suite.Equal("quux", filter.Keywords[1].Keyword)
		suite.True(filter.Keywords[1].WholeWord)

		suite.Equal("blah", filter.Keywords[2].Keyword)
		suite.True(filter.Keywords[1].WholeWord)
	}

	if suite.Len(filter.Statuses, 1) {
		slices.SortFunc(filter.Statuses, func(lhs, rhs apimodel.FilterStatus) int {
			return strings.Compare(lhs.ID, rhs.ID)
		})

		suite.Equal("01HEN2QRFA8H3C6QPN7RD4KSR6", filter.Statuses[0].StatusID)
	}

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPutFilterMinimal() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	id := suite.testFilters["local_account_1_filter_1"].ID
	title := "GNU/Linux"
	context := []string{"home"}
	filter, err := suite.putFilter(id, &title, &context, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, http.StatusOK, "", nil)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(title, filter.Title)
	filterContext := make([]string, 0, len(filter.Context))
	for _, c := range filter.Context {
		filterContext = append(filterContext, string(c))
	}
	suite.ElementsMatch(context, filterContext)
	suite.Equal(apimodel.FilterActionWarn, filter.FilterAction)
	suite.Nil(filter.ExpiresAt)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPutFilterEmptyTitle() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := ""
	context := []string{"home"}
	_, err := suite.putFilter(id, &title, &context, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter title must be provided, and must be no more than 200 chars"}`, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterEmptyContext() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := "GNU/Linux"
	context := []string{}
	_, err := suite.putFilter(id, &title, &context, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: at least one filter context is required"}`, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Changing our title to a title used by an existing filter should fail.
func (suite *FiltersTestSuite) TestPutFilterTitleConflict() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := suite.testFilters["local_account_1_filter_2"].Title
	_, err := suite.putFilter(id, &title, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, http.StatusConflict, `{"error":"Conflict: you already have a filter with this title"}`, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutAnotherAccountsFilter() {
	id := suite.testFilters["local_account_2_filter_1"].ID
	title := "GNU/Linux"
	context := []string{"home"}
	_, err := suite.putFilter(id, &title, &context, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, http.StatusNotFound, `{"error":"Not Found"}`, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutNonexistentFilter() {
	id := "not_even_a_real_ULID"
	phrase := "GNU/Linux"
	context := []string{"home"}
	_, err := suite.putFilter(id, &phrase, &context, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, http.StatusNotFound, `{"error":"Not Found"}`, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// setFilterExpiration sets filter expiration.
func (suite *FiltersTestSuite) setFilterExpiration(id string, expiresIn *int, expiresInStr *string, requestJson *string) *apimodel.FilterV2 {
	filter, err := suite.putFilter(id, nil, nil, nil, expiresIn, expiresInStr, nil, nil, nil, nil, nil, nil, requestJson, http.StatusOK, "", nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
	return filter
}

// Regression test for https://codeberg.org/superseriousbusiness/gotosocial/issues/3497
func (suite *FiltersTestSuite) TestPutFilterUnsetExpirationDateEmptyString() {
	id := suite.testFilters["local_account_1_filter_2"].ID

	// Setup: set an expiration date for the filter.
	expiresIn := 86400
	filter := suite.setFilterExpiration(id, &expiresIn, nil, nil)
	if !suite.NotNil(filter.ExpiresAt) {
		suite.FailNow("Test precondition failed")
	}

	// Unset the filter's expiration date by setting it to an empty string.
	expiresInStr := ""
	filter = suite.setFilterExpiration(id, nil, &expiresInStr, nil)
	suite.Nil(filter.ExpiresAt)
}

// Regression test related to https://codeberg.org/superseriousbusiness/gotosocial/issues/3497
func (suite *FiltersTestSuite) TestPutFilterUnsetExpirationDateNullJSON() {
	id := suite.testFilters["local_account_1_filter_3"].ID

	// Setup: set an expiration date for the filter.
	expiresIn := 86400
	filter := suite.setFilterExpiration(id, &expiresIn, nil, nil)
	if !suite.NotNil(filter.ExpiresAt) {
		suite.FailNow("Test precondition failed")
	}

	// Unset the filter's expiration date by setting it to a null literal.
	requestJson := `{
		"expires_in": null
	}`
	filter = suite.setFilterExpiration(id, nil, nil, &requestJson)
	suite.Nil(filter.ExpiresAt)
}

// Regression test related to https://codeberg.org/superseriousbusiness/gotosocial/issues/3497
func (suite *FiltersTestSuite) TestPutFilterUnalteredExpirationDateJSON() {
	id := suite.testFilters["local_account_1_filter_4"].ID

	// Setup: set an expiration date for the filter.
	expiresIn := 86400
	filter := suite.setFilterExpiration(id, &expiresIn, nil, nil)
	if !suite.NotNil(filter.ExpiresAt) {
		suite.FailNow("Test precondition failed")
	}

	// Update nothing. There should still be an expiration date.
	requestJson := `{}`
	filter = suite.setFilterExpiration(id, nil, nil, &requestJson)
	suite.NotNil(filter.ExpiresAt)
}
