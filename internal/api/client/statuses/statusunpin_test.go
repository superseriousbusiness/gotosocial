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

package statuses_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusUnpinTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusUnpinTestSuite) createUnpin(
	expectedHTTPStatus int,
	expectedBody string,
	targetStatusID string,
) (*apimodel.Status, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["admin_account"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["admin_account"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["admin_account"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodPost, config.GetProtocol()+"://"+config.GetHost()+"/api/"+statuses.BasePath+"/"+targetStatusID+"/unpin", nil)
	ctx.Request.Header.Set("accept", "application/json")
	ctx.AddParam(statuses.IDKey, targetStatusID)

	// trigger the handler
	suite.statusModule.StatusUnpinPOSTHandler(ctx)

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
	if expectedBody != "" && string(b) != expectedBody {
		errs = append(errs, fmt.Sprintf("expected %s got %s", expectedBody, string(b)))
	}

	if len(errs) > 0 {
		return nil, errs.Combine()
	}

	resp := &apimodel.Status{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *StatusUnpinTestSuite) TestUnpinStatusOK() {
	// Unpin a pinned public status that this account owns.
	targetStatus := suite.testStatuses["admin_account_status_1"]

	resp, err := suite.createUnpin(http.StatusOK, "", targetStatus.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(resp.Pinned)
}

func (suite *StatusUnpinTestSuite) TestUnpinStatusNotFound() {
	// Unpin a pinned followers-only status owned by another account.
	targetStatus := suite.testStatuses["local_account_2_status_7"]

	if _, err := suite.createUnpin(http.StatusNotFound, `{"error":"Not Found"}`, targetStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *StatusUnpinTestSuite) TestUnpinStatusUnprocessable() {
	// Unpin a not-pinned status owned by another account.
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	if _, err := suite.createUnpin(
		http.StatusUnprocessableEntity,
		`{"error":"Unprocessable Entity: status 01F8MHAMCHF6Y650WCRSCP4WMY does not belong to account 01F8MH17FWEB39HZJ76B6VXSKF"}`,
		targetStatus.ID,
	); err != nil {
		suite.FailNow(err.Error())
	}
}

func TestStatusUnpinTestSuite(t *testing.T) {
	suite.Run(t, new(StatusUnpinTestSuite))
}
