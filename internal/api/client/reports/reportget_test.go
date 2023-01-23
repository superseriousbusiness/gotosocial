/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package reports_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/reports"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ReportGetTestSuite struct {
	ReportsStandardTestSuite
}

func (suite *ReportGetTestSuite) getReport(expectedHTTPStatus int, expectedBody string, reportID string) (*apimodel.Report, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_2"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_2"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_2"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_2"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+reports.BasePath+"/"+reportID, nil)
	ctx.Request.Header.Set("accept", "application/json")
	ctx.AddParam("id", reportID)

	// trigger the handler
	suite.reportsModule.ReportGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.MultiError{}

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs = append(errs, fmt.Sprintf("expected %d got %d", expectedHTTPStatus, resultCode))
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs = append(errs, fmt.Sprintf("expected %s got %s", expectedBody, string(b)))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.Report{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *ReportGetTestSuite) TestGetReport1() {
	targetReport := suite.testReports["local_account_2_report_remote_account_1"]

	report, err := suite.getReport(http.StatusOK, "", targetReport.ID)
	suite.NoError(err)
	suite.NotNil(report)

	b, err := json.MarshalIndent(&report, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01GP3AWY4CRDVRNZKW0TEAMB5R",
  "created_at": "2022-05-14T10:20:03.000Z",
  "action_taken": false,
  "action_taken_at": null,
  "action_taken_comment": null,
  "category": "other",
  "comment": "dark souls sucks, please yeet this nerd",
  "forwarded": true,
  "status_ids": [
    "01FVW7JHQFSFK166WWKR8CBA6M"
  ],
  "rule_ids": [],
  "target_account": {
    "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
    "username": "foss_satan",
    "acct": "foss_satan@fossbros-anonymous.io",
    "display_name": "big gerald",
    "locked": false,
    "bot": false,
    "created_at": "2021-09-26T10:52:36.000Z",
    "note": "i post about like, i dunno, stuff, or whatever!!!!",
    "url": "http://fossbros-anonymous.io/@foss_satan",
    "avatar": "",
    "avatar_static": "",
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 1,
    "last_status_at": "2021-09-20T10:40:37.000Z",
    "emojis": [],
    "fields": []
  }
}`, string(b))
}

func (suite *ReportGetTestSuite) TestGetReport2() {
	targetReport := suite.testReports["remote_account_1_report_local_account_2"]
	report, err := suite.getReport(http.StatusNotFound, `{"error":"Not Found"}`, targetReport.ID)
	suite.NoError(err)
	suite.Nil(report)
}

func (suite *ReportGetTestSuite) TestGetReport3() {
	report, err := suite.getReport(http.StatusBadRequest, `{"error":"Bad Request: no report id specified"}`, "")
	suite.NoError(err)
	suite.Nil(report)
}

func (suite *ReportGetTestSuite) TestGetReport4() {
	report, err := suite.getReport(http.StatusNotFound, `{"error":"Not Found"}`, "01GPJWHQS1BG0SF0WZ1SABC4RZ")
	suite.NoError(err)
	suite.Nil(report)
}

func TestReportGetTestSuite(t *testing.T) {
	suite.Run(t, &ReportGetTestSuite{})
}
