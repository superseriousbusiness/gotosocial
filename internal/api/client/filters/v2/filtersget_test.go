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

func (suite *FiltersTestSuite) getFilters(
	expectedHTTPStatus int,
	expectedBody string,
) ([]*apimodel.FilterV2, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.BasePath, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.filtersModule.FiltersGETHandler(ctx)

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

	resp := make([]*apimodel.FilterV2, 0)
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestGetFilters() {
	// Map of filter IDs to filter keyword and status IDs.
	expectedFilters := map[string]*struct {
		keywordIDs []string
		statusIDs  []string
	}{}

	// Collect the sets of IDs we expect to see.
	accountID := suite.testAccounts["local_account_1"].ID
	for _, filter := range suite.testFilters {
		if filter.AccountID == accountID {
			expectedFilters[filter.ID] = &struct {
				keywordIDs []string
				statusIDs  []string
			}{}
		}
	}
	for _, filterKeyword := range suite.testFilterKeywords {
		expected, ok := expectedFilters[filterKeyword.FilterID]
		if !ok {
			continue
		}
		expected.keywordIDs = append(expected.keywordIDs, filterKeyword.ID)
	}
	for _, filterStatus := range suite.testFilterStatuses {
		expected, ok := expectedFilters[filterStatus.FilterID]
		if !ok {
			continue
		}
		expected.statusIDs = append(expected.statusIDs, filterStatus.ID)
	}
	suite.NotEmpty(expectedFilters)

	// Fetch all filters for the logged-in account.
	filters, err := suite.getFilters(http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(filters)

	// Check that we got the right ones.
	suite.Len(filters, len(expectedFilters))

	for _, filter := range filters {
		expectedIDsForFilter := expectedFilters[filter.ID]

		var actualFilterKeywordIDs []string
		for _, filterKeyword := range filter.Keywords {
			actualFilterKeywordIDs = append(actualFilterKeywordIDs, filterKeyword.ID)
		}
		suite.ElementsMatch(actualFilterKeywordIDs, expectedIDsForFilter.keywordIDs)

		var actualFilterStatusIDs []string
		for _, filterStatus := range filter.Statuses {
			actualFilterStatusIDs = append(actualFilterStatusIDs, filterStatus.ID)
		}
		suite.ElementsMatch(actualFilterStatusIDs, expectedIDsForFilter.statusIDs)
	}
}
