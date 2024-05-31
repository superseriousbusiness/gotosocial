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

func (suite *FiltersTestSuite) putFilter(filterID string, title *string, context *[]string, action *string, expiresIn *int, requestJson *string, expectedHTTPStatus int, expectedBody string) (*apimodel.FilterV2, error) {
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
	id := suite.testFilters["local_account_1_filter_2"].ID
	title := "messy synoptic varblabbles"
	context := []string{"home", "public"}
	action := "hide"
	expiresIn := 86400
	filter, err := suite.putFilter(id, &title, &context, &action, &expiresIn, nil, http.StatusOK, "")
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
	suite.Len(filter.Keywords, 3)
	suite.Len(filter.Statuses, 0)
}

func (suite *FiltersTestSuite) TestPutFilterFullJSON() {
	id := suite.testFilters["local_account_1_filter_2"].ID
	// Use a numeric literal with a fractional part to test the JSON-specific handling for non-integer "expires_in".
	requestJson := `{
		"title": "messy synoptic varblabbles",
		"context": ["home", "public"],
		"filter_action": "hide",
		"expires_in": 86400.1
	}`
	filter, err := suite.putFilter(id, nil, nil, nil, nil, &requestJson, http.StatusOK, "")
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
	suite.Len(filter.Keywords, 3)
	suite.Len(filter.Statuses, 0)
}

func (suite *FiltersTestSuite) TestPutFilterMinimal() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := "GNU/Linux"
	context := []string{"home"}
	filter, err := suite.putFilter(id, &title, &context, nil, nil, nil, http.StatusOK, "")
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
}

func (suite *FiltersTestSuite) TestPutFilterEmptyTitle() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := ""
	context := []string{"home"}
	_, err := suite.putFilter(id, &title, &context, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter title must be provided, and must be no more than 200 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterEmptyContext() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := "GNU/Linux"
	context := []string{}
	_, err := suite.putFilter(id, &title, &context, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: at least one filter context is required"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Changing our title to a title used by an existing filter should fail.
func (suite *FiltersTestSuite) TestPutFilterTitleConflict() {
	id := suite.testFilters["local_account_1_filter_1"].ID
	title := suite.testFilters["local_account_1_filter_2"].Title
	_, err := suite.putFilter(id, &title, nil, nil, nil, nil, http.StatusConflict, `{"error":"Conflict: you already have a filter with this title"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutAnotherAccountsFilter() {
	id := suite.testFilters["local_account_2_filter_1"].ID
	title := "GNU/Linux"
	context := []string{"home"}
	_, err := suite.putFilter(id, &title, &context, nil, nil, nil, http.StatusNotFound, `{"error":"Not Found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutNonexistentFilter() {
	id := "not_even_a_real_ULID"
	phrase := "GNU/Linux"
	context := []string{"home"}
	_, err := suite.putFilter(id, &phrase, &context, nil, nil, nil, http.StatusNotFound, `{"error":"Not Found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
