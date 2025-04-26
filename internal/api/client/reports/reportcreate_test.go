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

package reports_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/reports"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ReportCreateTestSuite struct {
	ReportsStandardTestSuite
}

func (suite *ReportCreateTestSuite) createReport(expectedHTTPStatus int, expectedBody string, form *apimodel.ReportCreateRequest) (*apimodel.Report, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodPost, config.GetProtocol()+"://"+config.GetHost()+"/api/"+reports.BasePath, nil)
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values{
		"account_id":   {form.AccountID},
		"status_ids[]": form.StatusIDs,
		"comment":      {form.Comment},
		"forward":      {strconv.FormatBool(form.Forward)},
		"category":     {form.Category},
		"rule_ids[]":   form.RuleIDs,
	}

	// trigger the handler
	suite.reportsModule.ReportPOSTHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.Report{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *ReportCreateTestSuite) ReportOK(form *apimodel.ReportCreateRequest, report *apimodel.Report) {
	suite.Equal(form.AccountID, report.TargetAccount.ID)
	suite.Equal(form.StatusIDs, report.StatusIDs)
	suite.Equal(form.Comment, report.Comment)
	suite.Equal(form.Forward, report.Forwarded)
}

func (suite *ReportCreateTestSuite) TestCreateReport1() {
	targetAccount := suite.testAccounts["remote_account_1"]

	form := &apimodel.ReportCreateRequest{
		AccountID: targetAccount.ID,
		StatusIDs: []string{},
		Comment:   "",
		Forward:   false,
	}

	report, err := suite.createReport(http.StatusOK, "", form)
	suite.NoError(err)
	suite.NotEmpty(report)
	suite.ReportOK(form, report)
}

func (suite *ReportCreateTestSuite) TestCreateReport2() {
	targetAccount := suite.testAccounts["remote_account_1"]
	targetStatus := suite.testStatuses["remote_account_1_status_1"]

	form := &apimodel.ReportCreateRequest{
		AccountID: targetAccount.ID,
		StatusIDs: []string{targetStatus.ID},
		Comment:   "noooo don't post your so sexy aha",
		Forward:   true,
	}

	report, err := suite.createReport(http.StatusOK, "", form)
	suite.NoError(err)
	suite.NotEmpty(report)
	suite.ReportOK(form, report)
}

func (suite *ReportCreateTestSuite) TestCreateReport3() {
	form := &apimodel.ReportCreateRequest{}

	report, err := suite.createReport(http.StatusBadRequest, `{"error":"Bad Request: account_id must be set"}`, form)
	suite.NoError(err)
	suite.Nil(report)
}

func (suite *ReportCreateTestSuite) TestCreateReport4() {
	form := &apimodel.ReportCreateRequest{
		AccountID: "boobs",
		StatusIDs: []string{},
		Comment:   "",
		Forward:   true,
	}

	report, err := suite.createReport(http.StatusBadRequest, `{"error":"Bad Request: account_id was not valid"}`, form)
	suite.NoError(err)
	suite.Nil(report)
}

func (suite *ReportCreateTestSuite) TestCreateReport5() {
	testAccount := suite.testAccounts["local_account_1"]
	form := &apimodel.ReportCreateRequest{
		AccountID: testAccount.ID,
	}

	report, err := suite.createReport(http.StatusBadRequest, `{"error":"Bad Request: cannot report your own account"}`, form)
	suite.NoError(err)
	suite.Nil(report)
}

func (suite *ReportCreateTestSuite) TestCreateReport6() {
	targetAccount := suite.testAccounts["remote_account_1"]

	form := &apimodel.ReportCreateRequest{
		AccountID: targetAccount.ID,
		Comment:   "netus et malesuada fames ac turpis egestas sed tempus urna et pharetra pharetra massa massa ultricies mi quis hendrerit dolor magna eget est lorem ipsum dolor sit amet consectetur adipiscing elit pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas integer eget aliquet nibh praesent tristique magna sit amet purus gravida quis blandit turpis cursus in hac habitasse platea dictumst quisque sagittis purus sit amet volutpat consequat mauris nunc congue nisi vitae suscipit tellus mauris a diam maecenas sed enim ut sem viverra aliquet eget sit amet tellus cras adipiscing enim eu turpis egestas pretium aenean pharetra magna ac placerat vestibulum lectus mauris ultrices eros in cursus turpis massa tincidunt dui ut ornare lectus sit amet est placerat in egestas erat imperdiet sed euismod nisi porta lorem mollis aliquam ut porttitor leo a diam sollicitudin tempor id eu nisl nunc mi ipsum faucibus vitae aliquet nec ullamcorper sit amet risus nullam eget felis eget nunc lobortis mattis aliquam faucibus purus in massa tempor nec feugiat nisl pretium fusce id velit ut tortor pretium viverra suspendisse potenti nullam ac tortor vitae purus faucibus ornare suspendisse sed nisi lacus sed viverra tellus in hac habitasse platea dictumst vestibulum rhoncus est pellentesque elit ullamcorper dignissim cras tincidunt lobortis feugiat vivamus at augue eget arcu dictum varius duis at consectetur lorem donec massa sapien faucibus et molestie ac feugiat sed lectus vestibulum mattis ullamcorper velit sed ullamcorper morbi tincidunt ornare massa eget ",
	}

	report, err := suite.createReport(http.StatusBadRequest, `{"error":"Bad Request: comment length must be no more than 1000 chars, provided comment was 1588 chars"}`, form)
	suite.NoError(err)
	suite.Nil(report)
}

func (suite *ReportCreateTestSuite) TestCreateReport7() {
	form := &apimodel.ReportCreateRequest{
		AccountID: "01GPGH5ENXWE5K65YNNXYWAJA4",
	}

	report, err := suite.createReport(http.StatusBadRequest, `{"error":"Bad Request: account with ID 01GPGH5ENXWE5K65YNNXYWAJA4 does not exist"}`, form)
	suite.NoError(err)
	suite.Nil(report)
}

func TestReportCreateTestSuite(t *testing.T) {
	suite.Run(t, &ReportCreateTestSuite{})
}
