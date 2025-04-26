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
	"strings"

	filtersV2 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v2"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/stream"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

func (suite *FiltersTestSuite) postFilterStatus(
	filterID string,
	statusID *string,
	requestJson *string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.FilterStatus, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodPost, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.BasePath+"/"+filterID+"/statuses", nil)
	ctx.Request.Header.Set("accept", "application/json")
	if requestJson != nil {
		ctx.Request.Header.Set("content-type", "application/json")
		ctx.Request.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		ctx.Request.Form = make(url.Values)
		if statusID != nil {
			ctx.Request.Form["status_id"] = []string{*statusID}
		}
	}

	ctx.AddParam("id", filterID)

	// trigger the handler
	suite.filtersModule.FilterStatusPOSTHandler(ctx)

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

	resp := &apimodel.FilterStatus{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestPostFilterStatus() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterID := suite.testFilters["local_account_1_filter_1"].ID
	statusID := suite.testStatuses["admin_account_status_1"].ID
	filterStatus, err := suite.postFilterStatus(filterID, &statusID, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(statusID, filterStatus.StatusID)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterStatusJSON() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterID := suite.testFilters["local_account_1_filter_1"].ID
	requestJson := `{
		"status_id": "01F8MH75CBF9JFX4ZAD54N0W0R"
	}`
	filterStatus, err := suite.postFilterStatus(filterID, nil, &requestJson, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(suite.testStatuses["admin_account_status_1"].ID, filterStatus.StatusID)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterStatusEmptyStatusID() {
	filterID := suite.testFilters["local_account_1_filter_1"].ID
	statusID := ""
	_, err := suite.postFilterStatus(filterID, &statusID, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: status_id must be provided"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterStatusInvalidStatusID() {
	filterID := suite.testFilters["local_account_1_filter_1"].ID
	statusID := "112401162517176488" // ma'am, that's clearly a Mastodon ID, this is a Wendy's
	_, err := suite.postFilterStatus(filterID, &statusID, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: status_id didn't match the expected ULID format for an ID (26 characters from the set 0123456789ABCDEFGHJKMNPQRSTVWXYZ)"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterStatusMissingStatusID() {
	filterID := suite.testFilters["local_account_1_filter_1"].ID
	_, err := suite.postFilterStatus(filterID, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: status_id must be provided"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Creating another filter status in the same filter with the same status ID should fail.
func (suite *FiltersTestSuite) TestPostFilterStatusStatusIDConflict() {
	filterID := suite.testFilters["local_account_1_filter_3"].ID
	statusID := suite.testFilterStatuses["local_account_1_filter_3_status_1"].StatusID
	_, err := suite.postFilterStatus(filterID, &statusID, nil, http.StatusConflict, `{"error":"Conflict: duplicate status"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterStatusAnotherAccountsFilter() {
	filterID := suite.testFilters["local_account_2_filter_1"].ID
	statusID := suite.testStatuses["admin_account_status_1"].ID
	_, err := suite.postFilterStatus(filterID, &statusID, nil, http.StatusNotFound, `{"error":"Not Found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterStatusNonexistentFilter() {
	filterID := "not_even_a_real_ULID"
	statusID := suite.testStatuses["admin_account_status_1"].ID
	_, err := suite.postFilterStatus(filterID, &statusID, nil, http.StatusNotFound, `{"error":"Not Found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
