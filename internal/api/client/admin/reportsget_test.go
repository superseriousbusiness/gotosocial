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

package admin_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ReportsGetTestSuite struct {
	AdminStandardTestSuite
}

func (suite *ReportsGetTestSuite) getReports(
	account *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	expectedHTTPStatus int,
	expectedBody string,
	resolved *bool,
	accountID string,
	targetAccountID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) ([]*apimodel.AdminReport, string, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, account)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, user)

	// create the request URI
	requestPath := admin.ReportsPath + "?" + admin.LimitKey + "=" + strconv.Itoa(limit)
	if resolved != nil {
		requestPath = requestPath + "&" + admin.ResolvedKey + "=" + strconv.FormatBool(*resolved)
	}
	if accountID != "" {
		requestPath = requestPath + "&" + admin.AccountIDKey + "=" + accountID
	}
	if targetAccountID != "" {
		requestPath = requestPath + "&" + admin.TargetAccountIDKey + "=" + targetAccountID
	}
	if maxID != "" {
		requestPath = requestPath + "&" + admin.MaxIDKey + "=" + maxID
	}
	if sinceID != "" {
		requestPath = requestPath + "&" + admin.SinceIDKey + "=" + sinceID
	}
	if minID != "" {
		requestPath = requestPath + "&" + admin.MinIDKey + "=" + minID
	}
	baseURI := config.GetProtocol() + "://" + config.GetHost()
	requestURI := baseURI + "/api/" + requestPath

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.adminModule.ReportsGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	errs := gtserror.MultiError{}

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs = append(errs, fmt.Sprintf("expected %d got %d", expectedHTTPStatus, resultCode))
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs = append(errs, fmt.Sprintf("expected %s got %s", expectedBody, string(b)))
		}
		return nil, "", errs.Combine()
	}

	resp := []*apimodel.AdminReport{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, "", err
	}

	return resp, result.Header.Get("Link"), nil
}

func (suite *ReportsGetTestSuite) TestReportsGet1() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", nil, "", "", "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
    "id": "01GP3DFY9XQ1TJMZT5BGAZPXX7",
    "action_taken": true,
    "action_taken_at": "2022-05-15T15:01:56.000Z",
    "category": "other",
    "comment": "this is a turtle, not a person, therefore should not be a poster",
    "forwarded": true,
    "created_at": "2022-05-15T14:20:12.000Z",
    "updated_at": "2022-05-15T14:20:12.000Z",
    "account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "domain": "fossbros-anonymous.io",
      "created_at": "2021-09-26T10:52:36.000Z",
      "email": "",
      "ip": null,
      "ips": [],
      "locale": "",
      "invite_request": null,
      "role": "user",
      "confirmed": false,
      "approved": false,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
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
    },
    "target_account": {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "domain": null,
      "created_at": "2022-06-04T13:12:00.000Z",
      "email": "tortle.dude@example.org",
      "ip": "118.44.18.196",
      "ips": [],
      "locale": "en",
      "invite_request": "",
      "role": "user",
      "confirmed": true,
      "approved": true,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
        "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
        "username": "1happyturtle",
        "acct": "1happyturtle",
        "display_name": "happy little turtle :3",
        "locked": true,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.png",
        "header_static": "http://localhost:8080/assets/default_header.png",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 7,
        "last_status_at": "2021-10-20T10:40:37.000Z",
        "emojis": [],
        "fields": [],
        "role": "user"
      },
      "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
    },
    "assigned_account": {
      "id": "01F8MH17FWEB39HZJ76B6VXSKF",
      "username": "admin",
      "domain": null,
      "created_at": "2022-05-17T13:10:59.000Z",
      "email": "admin@example.org",
      "ip": "89.122.255.1",
      "ips": [],
      "locale": "en",
      "invite_request": "",
      "role": "admin",
      "confirmed": true,
      "approved": true,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
        "id": "01F8MH17FWEB39HZJ76B6VXSKF",
        "username": "admin",
        "acct": "admin",
        "display_name": "",
        "locked": false,
        "bot": false,
        "created_at": "2022-05-17T13:10:59.000Z",
        "note": "",
        "url": "http://localhost:8080/@admin",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.png",
        "header_static": "http://localhost:8080/assets/default_header.png",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 4,
        "last_status_at": "2021-10-20T10:41:37.000Z",
        "emojis": [],
        "fields": [],
        "enable_rss": true,
        "role": "admin"
      },
      "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
    },
    "action_taken_by_account": {
      "id": "01F8MH17FWEB39HZJ76B6VXSKF",
      "username": "admin",
      "domain": null,
      "created_at": "2022-05-17T13:10:59.000Z",
      "email": "admin@example.org",
      "ip": "89.122.255.1",
      "ips": [],
      "locale": "en",
      "invite_request": "",
      "role": "admin",
      "confirmed": true,
      "approved": true,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
        "id": "01F8MH17FWEB39HZJ76B6VXSKF",
        "username": "admin",
        "acct": "admin",
        "display_name": "",
        "locked": false,
        "bot": false,
        "created_at": "2022-05-17T13:10:59.000Z",
        "note": "",
        "url": "http://localhost:8080/@admin",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.png",
        "header_static": "http://localhost:8080/assets/default_header.png",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 4,
        "last_status_at": "2021-10-20T10:41:37.000Z",
        "emojis": [],
        "fields": [],
        "enable_rss": true,
        "role": "admin"
      },
      "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
    },
    "statuses": [],
    "rule_ids": [],
    "action_taken_comment": "user was warned not to be a turtle anymore"
  },
  {
    "id": "01GP3AWY4CRDVRNZKW0TEAMB5R",
    "action_taken": false,
    "action_taken_at": null,
    "category": "other",
    "comment": "dark souls sucks, please yeet this nerd",
    "forwarded": true,
    "created_at": "2022-05-14T10:20:03.000Z",
    "updated_at": "2022-05-14T10:20:03.000Z",
    "account": {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "domain": null,
      "created_at": "2022-06-04T13:12:00.000Z",
      "email": "tortle.dude@example.org",
      "ip": "118.44.18.196",
      "ips": [],
      "locale": "en",
      "invite_request": "",
      "role": "user",
      "confirmed": true,
      "approved": true,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
        "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
        "username": "1happyturtle",
        "acct": "1happyturtle",
        "display_name": "happy little turtle :3",
        "locked": true,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.png",
        "header_static": "http://localhost:8080/assets/default_header.png",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 7,
        "last_status_at": "2021-10-20T10:40:37.000Z",
        "emojis": [],
        "fields": [],
        "role": "user"
      },
      "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
    },
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "domain": "fossbros-anonymous.io",
      "created_at": "2021-09-26T10:52:36.000Z",
      "email": "",
      "ip": null,
      "ips": [],
      "locale": "",
      "invite_request": null,
      "role": "user",
      "confirmed": false,
      "approved": false,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
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
    },
    "assigned_account": null,
    "action_taken_by_account": null,
    "statuses": [
      {
        "id": "01FVW7JHQFSFK166WWKR8CBA6M",
        "created_at": "2021-09-20T10:40:37.000Z",
        "in_reply_to_id": null,
        "in_reply_to_account_id": null,
        "sensitive": false,
        "spoiler_text": "",
        "visibility": "unlisted",
        "language": "en",
        "uri": "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
        "url": "http://fossbros-anonymous.io/@foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
        "replies_count": 0,
        "reblogs_count": 0,
        "favourites_count": 0,
        "favourited": false,
        "reblogged": false,
        "muted": false,
        "bookmarked": false,
        "pinned": false,
        "content": "dark souls status bot: \"thoughts of dog\"",
        "reblog": null,
        "account": {
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
        },
        "media_attachments": [
          {
            "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
            "type": "image",
            "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
            "preview_remote_url": "http://fossbros-anonymous.io/attachments/small/a499f55b-2d1e-4acd-98d2-1ac2ba6d79b9.jpg",
            "meta": {
              "original": {
                "width": 472,
                "height": 291,
                "size": "472x291",
                "aspect": 1.6219932
              },
              "small": {
                "width": 472,
                "height": 291,
                "size": "472x291",
                "aspect": 1.6219932
              },
              "focus": {
                "x": 0,
                "y": 0
              }
            },
            "description": "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
            "blurhash": "LARysgM_IU_3~pD%M_Rj_39FIAt6"
          }
        ],
        "mentions": [],
        "tags": [],
        "emojis": [],
        "card": null,
        "poll": null
      }
    ],
    "rule_ids": [],
    "action_taken_comment": null
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=20&min_id=01GP3DFY9XQ1TJMZT5BGAZPXX7>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGet2() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]
	account := suite.testAccounts["local_account_2"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", nil, account.ID, "", "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
    "id": "01GP3AWY4CRDVRNZKW0TEAMB5R",
    "action_taken": false,
    "action_taken_at": null,
    "category": "other",
    "comment": "dark souls sucks, please yeet this nerd",
    "forwarded": true,
    "created_at": "2022-05-14T10:20:03.000Z",
    "updated_at": "2022-05-14T10:20:03.000Z",
    "account": {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "domain": null,
      "created_at": "2022-06-04T13:12:00.000Z",
      "email": "tortle.dude@example.org",
      "ip": "118.44.18.196",
      "ips": [],
      "locale": "en",
      "invite_request": "",
      "role": "user",
      "confirmed": true,
      "approved": true,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
        "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
        "username": "1happyturtle",
        "acct": "1happyturtle",
        "display_name": "happy little turtle :3",
        "locked": true,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.png",
        "header_static": "http://localhost:8080/assets/default_header.png",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 7,
        "last_status_at": "2021-10-20T10:40:37.000Z",
        "emojis": [],
        "fields": [],
        "role": "user"
      },
      "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
    },
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "domain": "fossbros-anonymous.io",
      "created_at": "2021-09-26T10:52:36.000Z",
      "email": "",
      "ip": null,
      "ips": [],
      "locale": "",
      "invite_request": null,
      "role": "user",
      "confirmed": false,
      "approved": false,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
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
    },
    "assigned_account": null,
    "action_taken_by_account": null,
    "statuses": [
      {
        "id": "01FVW7JHQFSFK166WWKR8CBA6M",
        "created_at": "2021-09-20T10:40:37.000Z",
        "in_reply_to_id": null,
        "in_reply_to_account_id": null,
        "sensitive": false,
        "spoiler_text": "",
        "visibility": "unlisted",
        "language": "en",
        "uri": "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
        "url": "http://fossbros-anonymous.io/@foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
        "replies_count": 0,
        "reblogs_count": 0,
        "favourites_count": 0,
        "favourited": false,
        "reblogged": false,
        "muted": false,
        "bookmarked": false,
        "pinned": false,
        "content": "dark souls status bot: \"thoughts of dog\"",
        "reblog": null,
        "account": {
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
        },
        "media_attachments": [
          {
            "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
            "type": "image",
            "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
            "preview_remote_url": "http://fossbros-anonymous.io/attachments/small/a499f55b-2d1e-4acd-98d2-1ac2ba6d79b9.jpg",
            "meta": {
              "original": {
                "width": 472,
                "height": 291,
                "size": "472x291",
                "aspect": 1.6219932
              },
              "small": {
                "width": 472,
                "height": 291,
                "size": "472x291",
                "aspect": 1.6219932
              },
              "focus": {
                "x": 0,
                "y": 0
              }
            },
            "description": "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
            "blurhash": "LARysgM_IU_3~pD%M_Rj_39FIAt6"
          }
        ],
        "mentions": [],
        "tags": [],
        "emojis": [],
        "card": null,
        "poll": null
      }
    ],
    "rule_ids": [],
    "action_taken_comment": null
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R&account_id=01F8MH5NBDF2MV7CTC4Q5128HF>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R&account_id=01F8MH5NBDF2MV7CTC4Q5128HF>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGet3() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]
	targetAccount := suite.testAccounts["remote_account_1"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", nil, "", targetAccount.ID, "", "", "", 20)
	suite.NoError(err)
	suite.NotEmpty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[
  {
    "id": "01GP3AWY4CRDVRNZKW0TEAMB5R",
    "action_taken": false,
    "action_taken_at": null,
    "category": "other",
    "comment": "dark souls sucks, please yeet this nerd",
    "forwarded": true,
    "created_at": "2022-05-14T10:20:03.000Z",
    "updated_at": "2022-05-14T10:20:03.000Z",
    "account": {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "domain": null,
      "created_at": "2022-06-04T13:12:00.000Z",
      "email": "tortle.dude@example.org",
      "ip": "118.44.18.196",
      "ips": [],
      "locale": "en",
      "invite_request": "",
      "role": "user",
      "confirmed": true,
      "approved": true,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
        "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
        "username": "1happyturtle",
        "acct": "1happyturtle",
        "display_name": "happy little turtle :3",
        "locked": true,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.png",
        "header_static": "http://localhost:8080/assets/default_header.png",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 7,
        "last_status_at": "2021-10-20T10:40:37.000Z",
        "emojis": [],
        "fields": [],
        "role": "user"
      },
      "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
    },
    "target_account": {
      "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
      "username": "foss_satan",
      "domain": "fossbros-anonymous.io",
      "created_at": "2021-09-26T10:52:36.000Z",
      "email": "",
      "ip": null,
      "ips": [],
      "locale": "",
      "invite_request": null,
      "role": "user",
      "confirmed": false,
      "approved": false,
      "disabled": false,
      "silenced": false,
      "suspended": false,
      "account": {
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
    },
    "assigned_account": null,
    "action_taken_by_account": null,
    "statuses": [
      {
        "id": "01FVW7JHQFSFK166WWKR8CBA6M",
        "created_at": "2021-09-20T10:40:37.000Z",
        "in_reply_to_id": null,
        "in_reply_to_account_id": null,
        "sensitive": false,
        "spoiler_text": "",
        "visibility": "unlisted",
        "language": "en",
        "uri": "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
        "url": "http://fossbros-anonymous.io/@foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
        "replies_count": 0,
        "reblogs_count": 0,
        "favourites_count": 0,
        "favourited": false,
        "reblogged": false,
        "muted": false,
        "bookmarked": false,
        "pinned": false,
        "content": "dark souls status bot: \"thoughts of dog\"",
        "reblog": null,
        "account": {
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
        },
        "media_attachments": [
          {
            "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
            "type": "image",
            "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
            "preview_remote_url": "http://fossbros-anonymous.io/attachments/small/a499f55b-2d1e-4acd-98d2-1ac2ba6d79b9.jpg",
            "meta": {
              "original": {
                "width": 472,
                "height": 291,
                "size": "472x291",
                "aspect": 1.6219932
              },
              "small": {
                "width": 472,
                "height": 291,
                "size": "472x291",
                "aspect": 1.6219932
              },
              "focus": {
                "x": 0,
                "y": 0
              }
            },
            "description": "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
            "blurhash": "LARysgM_IU_3~pD%M_Rj_39FIAt6"
          }
        ],
        "mentions": [],
        "tags": [],
        "emojis": [],
        "card": null,
        "poll": null
      }
    ],
    "rule_ids": [],
    "action_taken_comment": null
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGet4() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]
	resolved := testrig.FalseBool()
	targetAccount := suite.testAccounts["local_account_2"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", resolved, "", targetAccount.ID, "", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[]`, string(b))
	suite.Empty(link)
}

func (suite *ReportsGetTestSuite) TestReportsGet6() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	reports, _, err := suite.getReports(testAccount, testToken, testUser, http.StatusForbidden, `{"error":"Forbidden: user 01F8MGVGPHQ2D3P3X0454H54Z5 not an admin"}`, nil, "", "", "", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)
}

func TestReportsGetTestSuite(t *testing.T) {
	suite.Run(t, &ReportsGetTestSuite{})
}
