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
	"code.superseriousbusiness.org/gotosocial/testrig"
)

func (suite *FiltersTestSuite) getFilterStatus(
	filterStatusID string,
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
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.StatusPath+"/"+filterStatusID, nil)
	ctx.Request.Header.Set("accept", "application/json")

	ctx.AddParam("id", filterStatusID)

	// trigger the handler
	suite.filtersModule.FilterStatusGETHandler(ctx)

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

func (suite *FiltersTestSuite) TestGetFilterStatus() {
	expectedFilterStatus := suite.testFilterStatuses["local_account_1_filter_3_status_1"]

	filterStatus, err := suite.getFilterStatus(expectedFilterStatus.ID, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(filterStatus)
	suite.Equal(expectedFilterStatus.ID, filterStatus.ID)
	suite.Equal(expectedFilterStatus.StatusID, filterStatus.StatusID)
}

func (suite *FiltersTestSuite) TestGetAnotherAccountsFilterStatus() {
	id := suite.testFilterStatuses["local_account_2_filter_1_status_1"].ID

	_, err := suite.getFilterStatus(id, http.StatusNotFound, `{"error":"Not Found: filter not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestGetNonexistentFilterStatus() {
	id := "not_even_a_real_ULID"

	_, err := suite.getFilterStatus(id, http.StatusNotFound, `{"error":"Not Found: filter status not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
