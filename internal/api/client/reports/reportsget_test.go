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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/reports"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ReportsGetTestSuite struct {
	ReportsStandardTestSuite
}

func (suite *ReportsGetTestSuite) getReports(
	account *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	expectedHTTPStatus int,
	resolved *bool,
	targetAccountID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) ([]*apimodel.Report, string, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, account)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, user)

	// create the request URI
	requestPath := reports.BasePath + "?" + apiutil.LimitKey + "=" + strconv.Itoa(limit)
	if resolved != nil {
		requestPath = requestPath + "&" + apiutil.ResolvedKey + "=" + strconv.FormatBool(*resolved)
	}
	if targetAccountID != "" {
		requestPath = requestPath + "&" + apiutil.TargetAccountIDKey + "=" + targetAccountID
	}
	if maxID != "" {
		requestPath = requestPath + "&" + apiutil.MaxIDKey + "=" + maxID
	}
	if sinceID != "" {
		requestPath = requestPath + "&" + apiutil.SinceIDKey + "=" + sinceID
	}
	if minID != "" {
		requestPath = requestPath + "&" + apiutil.MinIDKey + "=" + minID
	}
	baseURI := config.GetProtocol() + "://" + config.GetHost()
	requestURI := baseURI + "/api/" + requestPath

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.reportsModule.ReportsGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		return nil, "", fmt.Errorf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	resp := []*apimodel.Report{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, "", err
	}

	return resp, result.Header.Get("Link"), nil
}

func (suite *ReportsGetTestSuite) TestGetReports() {
	testAccount := suite.testAccounts["local_account_2"]
	testToken := suite.testTokens["local_account_2"]
	testUser := suite.testUsers["local_account_2"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, nil, "", "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
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
    "rule_ids": [
      "01GP3AWY4CRDVRNZKW0TEAMB51",
      "01GP3DFY9XQ1TJMZT5BGAZPXX3"
    ],
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "acct": "foss_satan@fossbros-anonymous.io",
      "display_name": "big gerald",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2021-09-26T10:52:36.000Z",
      "note": "i post about like, i dunno, stuff, or whatever!!!!",
      "url": "http://fossbros-anonymous.io/@foss_satan",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "header_description": "Flat gray background (default header).",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 4,
      "last_status_at": "2024-11-01",
      "emojis": [],
      "fields": [],
      "group": false
    }
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="next", <http://localhost:8080/api/v1/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestGetReports2() {
	testAccount := suite.testAccounts["local_account_2"]
	testToken := suite.testTokens["local_account_2"]
	testUser := suite.testUsers["local_account_2"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, nil, "", "01GP3AWY4CRDVRNZKW0TEAMB5R", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[]`, string(b))
	suite.Empty(link)
}

func (suite *ReportsGetTestSuite) TestGetReports3() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, nil, "", "", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[]`, string(b))
	suite.Empty(link)
}

func (suite *ReportsGetTestSuite) TestGetReports4() {
	testAccount := suite.testAccounts["local_account_2"]
	testToken := suite.testTokens["local_account_2"]
	testUser := suite.testUsers["local_account_2"]
	resolved := util.Ptr(false)

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, resolved, "", "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
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
    "rule_ids": [
      "01GP3AWY4CRDVRNZKW0TEAMB51",
      "01GP3DFY9XQ1TJMZT5BGAZPXX3"
    ],
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "acct": "foss_satan@fossbros-anonymous.io",
      "display_name": "big gerald",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2021-09-26T10:52:36.000Z",
      "note": "i post about like, i dunno, stuff, or whatever!!!!",
      "url": "http://fossbros-anonymous.io/@foss_satan",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "header_description": "Flat gray background (default header).",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 4,
      "last_status_at": "2024-11-01",
      "emojis": [],
      "fields": [],
      "group": false
    }
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R&resolved=false>; rel="next", <http://localhost:8080/api/v1/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R&resolved=false>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestGetReports5() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]
	resolved := util.Ptr(true)

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, resolved, "", "", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[]`, string(b))
	suite.Empty(link)
}

func (suite *ReportsGetTestSuite) TestGetReports6() {
	testAccount := suite.testAccounts["local_account_2"]
	testToken := suite.testTokens["local_account_2"]
	testUser := suite.testUsers["local_account_2"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, nil, "01F8MH5ZK5VRH73AKHQM6Y9VNX", "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
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
    "rule_ids": [
      "01GP3AWY4CRDVRNZKW0TEAMB51",
      "01GP3DFY9XQ1TJMZT5BGAZPXX3"
    ],
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "acct": "foss_satan@fossbros-anonymous.io",
      "display_name": "big gerald",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2021-09-26T10:52:36.000Z",
      "note": "i post about like, i dunno, stuff, or whatever!!!!",
      "url": "http://fossbros-anonymous.io/@foss_satan",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "header_description": "Flat gray background (default header).",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 4,
      "last_status_at": "2024-11-01",
      "emojis": [],
      "fields": [],
      "group": false
    }
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="next", <http://localhost:8080/api/v1/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestGetReports7() {
	testAccount := suite.testAccounts["local_account_2"]
	testToken := suite.testTokens["local_account_2"]
	testUser := suite.testUsers["local_account_2"]
	resolved := util.Ptr(false)

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, resolved, "01F8MH5ZK5VRH73AKHQM6Y9VNX", "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
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
    "rule_ids": [
      "01GP3AWY4CRDVRNZKW0TEAMB51",
      "01GP3DFY9XQ1TJMZT5BGAZPXX3"
    ],
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "acct": "foss_satan@fossbros-anonymous.io",
      "display_name": "big gerald",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2021-09-26T10:52:36.000Z",
      "note": "i post about like, i dunno, stuff, or whatever!!!!",
      "url": "http://fossbros-anonymous.io/@foss_satan",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "header_description": "Flat gray background (default header).",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 4,
      "last_status_at": "2024-11-01",
      "emojis": [],
      "fields": [],
      "group": false
    }
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R&resolved=false&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="next", <http://localhost:8080/api/v1/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R&resolved=false&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="prev"`, link)
}

func TestReportsGetTestSuite(t *testing.T) {
	suite.Run(t, &ReportsGetTestSuite{})
}
