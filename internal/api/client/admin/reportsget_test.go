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

package admin_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
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
	requestPath := admin.ReportsPath + "?" + apiutil.LimitKey + "=" + strconv.Itoa(limit)
	if resolved != nil {
		requestPath = requestPath + "&" + apiutil.ResolvedKey + "=" + strconv.FormatBool(*resolved)
	}
	if accountID != "" {
		requestPath = requestPath + "&" + apiutil.AccountIDKey + "=" + accountID
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
	suite.adminModule.ReportsGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	errs := gtserror.NewMultiError(2)

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, "", errs.Combine()
	}

	resp := []*apimodel.AdminReport{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, "", err
	}

	return resp, result.Header.Get("Link"), nil
}

func (suite *ReportsGetTestSuite) TestReportsGetAll() {
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
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
    },
    "target_account": {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "domain": null,
      "created_at": "2022-06-04T13:12:00.000Z",
      "email": "tortle.dude@example.org",
      "ip": null,
      "ips": [],
      "locale": "en",
      "invite_request": null,
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
        "discoverable": false,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.webp",
        "header_static": "http://localhost:8080/assets/default_header.webp",
        "header_description": "Flat gray background (default header).",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 9,
        "last_status_at": "2024-11-01",
        "emojis": [],
        "fields": [
          {
            "name": "should you follow me?",
            "value": "maybe!",
            "verified_at": null
          },
          {
            "name": "age",
            "value": "120",
            "verified_at": null
          }
        ],
        "hide_collections": true,
        "group": false
      },
      "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
    },
    "assigned_account": {
      "id": "01F8MH17FWEB39HZJ76B6VXSKF",
      "username": "admin",
      "domain": null,
      "created_at": "2022-05-17T13:10:59.000Z",
      "email": "admin@example.org",
      "ip": null,
      "ips": [],
      "locale": "en",
      "invite_request": null,
      "role": {
        "id": "admin",
        "name": "admin",
        "color": "",
        "permissions": "546033",
        "highlighted": true
      },
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
        "discoverable": true,
        "bot": false,
        "created_at": "2022-05-17T13:10:59.000Z",
        "note": "",
        "url": "http://localhost:8080/@admin",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.webp",
        "header_static": "http://localhost:8080/assets/default_header.webp",
        "header_description": "Flat gray background (default header).",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 4,
        "last_status_at": "2021-10-20",
        "emojis": [],
        "fields": [],
        "enable_rss": true,
        "roles": [
          {
            "id": "admin",
            "name": "admin",
            "color": ""
          }
        ],
        "group": false
      },
      "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
    },
    "action_taken_by_account": {
      "id": "01F8MH17FWEB39HZJ76B6VXSKF",
      "username": "admin",
      "domain": null,
      "created_at": "2022-05-17T13:10:59.000Z",
      "email": "admin@example.org",
      "ip": null,
      "ips": [],
      "locale": "en",
      "invite_request": null,
      "role": {
        "id": "admin",
        "name": "admin",
        "color": "",
        "permissions": "546033",
        "highlighted": true
      },
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
        "discoverable": true,
        "bot": false,
        "created_at": "2022-05-17T13:10:59.000Z",
        "note": "",
        "url": "http://localhost:8080/@admin",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.webp",
        "header_static": "http://localhost:8080/assets/default_header.webp",
        "header_description": "Flat gray background (default header).",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 4,
        "last_status_at": "2021-10-20",
        "emojis": [],
        "fields": [],
        "enable_rss": true,
        "roles": [
          {
            "id": "admin",
            "name": "admin",
            "color": ""
          }
        ],
        "group": false
      },
      "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
    },
    "statuses": [],
    "rules": [],
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
      "ip": null,
      "ips": [],
      "locale": "en",
      "invite_request": null,
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
        "discoverable": false,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.webp",
        "header_static": "http://localhost:8080/assets/default_header.webp",
        "header_description": "Flat gray background (default header).",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 9,
        "last_status_at": "2024-11-01",
        "emojis": [],
        "fields": [
          {
            "name": "should you follow me?",
            "value": "maybe!",
            "verified_at": null
          },
          {
            "name": "age",
            "value": "120",
            "verified_at": null
          }
        ],
        "hide_collections": true,
        "group": false
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
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
    },
    "assigned_account": null,
    "action_taken_by_account": null,
    "statuses": [
      {
        "id": "01FVW7JHQFSFK166WWKR8CBA6M",
        "created_at": "2021-09-20T10:40:37.000Z",
        "edited_at": null,
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
        "content": "\u003cp\u003edark souls status bot: \"thoughts of dog\"\u003c/p\u003e",
        "reblog": null,
        "account": {
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
        },
        "media_attachments": [
          {
            "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
            "type": "image",
            "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.webp",
            "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
            "preview_remote_url": null,
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
            "blurhash": "L3Q9_@4n9E?axW4mD$Mx~q00Di%L"
          }
        ],
        "mentions": [],
        "tags": [],
        "emojis": [],
        "card": null,
        "poll": null,
        "interaction_policy": {
          "can_favourite": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          },
          "can_reply": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          },
          "can_reblog": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          }
        }
      }
    ],
    "rules": [
      {
        "id": "01GP3AWY4CRDVRNZKW0TEAMB51",
        "text": "Be gay"
      },
      {
        "id": "01GP3DFY9XQ1TJMZT5BGAZPXX3",
        "text": "Do crime"
      }
    ],
    "action_taken_comment": null
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=20&min_id=01GP3DFY9XQ1TJMZT5BGAZPXX7>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGetCreatedByAccount() {
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
      "ip": null,
      "ips": [],
      "locale": "en",
      "invite_request": null,
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
        "discoverable": false,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.webp",
        "header_static": "http://localhost:8080/assets/default_header.webp",
        "header_description": "Flat gray background (default header).",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 9,
        "last_status_at": "2024-11-01",
        "emojis": [],
        "fields": [
          {
            "name": "should you follow me?",
            "value": "maybe!",
            "verified_at": null
          },
          {
            "name": "age",
            "value": "120",
            "verified_at": null
          }
        ],
        "hide_collections": true,
        "group": false
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
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
    },
    "assigned_account": null,
    "action_taken_by_account": null,
    "statuses": [
      {
        "id": "01FVW7JHQFSFK166WWKR8CBA6M",
        "created_at": "2021-09-20T10:40:37.000Z",
        "edited_at": null,
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
        "content": "\u003cp\u003edark souls status bot: \"thoughts of dog\"\u003c/p\u003e",
        "reblog": null,
        "account": {
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
        },
        "media_attachments": [
          {
            "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
            "type": "image",
            "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.webp",
            "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
            "preview_remote_url": null,
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
            "blurhash": "L3Q9_@4n9E?axW4mD$Mx~q00Di%L"
          }
        ],
        "mentions": [],
        "tags": [],
        "emojis": [],
        "card": null,
        "poll": null,
        "interaction_policy": {
          "can_favourite": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          },
          "can_reply": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          },
          "can_reblog": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          }
        }
      }
    ],
    "rules": [
      {
        "id": "01GP3AWY4CRDVRNZKW0TEAMB51",
        "text": "Be gay"
      },
      {
        "id": "01GP3DFY9XQ1TJMZT5BGAZPXX3",
        "text": "Do crime"
      }
    ],
    "action_taken_comment": null
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?account_id=01F8MH5NBDF2MV7CTC4Q5128HF&limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="next", <http://localhost:8080/api/v1/admin/reports?account_id=01F8MH5NBDF2MV7CTC4Q5128HF&limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGetTargetAccount() {
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
      "ip": null,
      "ips": [],
      "locale": "en",
      "invite_request": null,
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
        "discoverable": false,
        "bot": false,
        "created_at": "2022-06-04T13:12:00.000Z",
        "note": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
        "url": "http://localhost:8080/@1happyturtle",
        "avatar": "",
        "avatar_static": "",
        "header": "http://localhost:8080/assets/default_header.webp",
        "header_static": "http://localhost:8080/assets/default_header.webp",
        "header_description": "Flat gray background (default header).",
        "followers_count": 1,
        "following_count": 1,
        "statuses_count": 9,
        "last_status_at": "2024-11-01",
        "emojis": [],
        "fields": [
          {
            "name": "should you follow me?",
            "value": "maybe!",
            "verified_at": null
          },
          {
            "name": "age",
            "value": "120",
            "verified_at": null
          }
        ],
        "hide_collections": true,
        "group": false
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
      "role": {
        "id": "user",
        "name": "user",
        "color": "",
        "permissions": "0",
        "highlighted": false
      },
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
    },
    "assigned_account": null,
    "action_taken_by_account": null,
    "statuses": [
      {
        "id": "01FVW7JHQFSFK166WWKR8CBA6M",
        "created_at": "2021-09-20T10:40:37.000Z",
        "edited_at": null,
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
        "content": "\u003cp\u003edark souls status bot: \"thoughts of dog\"\u003c/p\u003e",
        "reblog": null,
        "account": {
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
        },
        "media_attachments": [
          {
            "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
            "type": "image",
            "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
            "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.webp",
            "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
            "preview_remote_url": null,
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
            "blurhash": "L3Q9_@4n9E?axW4mD$Mx~q00Di%L"
          }
        ],
        "mentions": [],
        "tags": [],
        "emojis": [],
        "card": null,
        "poll": null,
        "interaction_policy": {
          "can_favourite": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          },
          "can_reply": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          },
          "can_reblog": {
            "automatic_approval": [
              "public",
              "me"
            ],
            "manual_approval": [],
            "always": [
              "public",
              "me"
            ],
            "with_approval": []
          }
        }
      }
    ],
    "rules": [
      {
        "id": "01GP3AWY4CRDVRNZKW0TEAMB51",
        "text": "Be gay"
      },
      {
        "id": "01GP3DFY9XQ1TJMZT5BGAZPXX3",
        "text": "Do crime"
      }
    ],
    "action_taken_comment": null
  }
]`, string(b))

	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=20&min_id=01GP3AWY4CRDVRNZKW0TEAMB5R&target_account_id=01F8MH5ZK5VRH73AKHQM6Y9VNX>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGetResolvedTargetAccount() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]
	resolved := util.Ptr(false)
	targetAccount := suite.testAccounts["local_account_2"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", resolved, "", targetAccount.ID, "", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)

	b, err := json.MarshalIndent(&reports, "", "  ")
	suite.NoError(err)

	suite.Equal(`[]`, string(b))
	suite.Empty(link)
}

func (suite *ReportsGetTestSuite) TestReportsGetNotAdmin() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	reports, _, err := suite.getReports(testAccount, testToken, testUser, http.StatusForbidden, `{"error":"Forbidden: token has insufficient scope permission"}`, nil, "", "", "", "", "", 20)
	suite.NoError(err)
	suite.Empty(reports)
}

func (suite *ReportsGetTestSuite) TestReportsGetZeroLimit() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", nil, "", "", "", "", "", 0)
	suite.NoError(err)
	suite.Len(reports, 2)

	// Limit in Link header should be set to default (20)
	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=20&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=20&min_id=01GP3DFY9XQ1TJMZT5BGAZPXX7>; rel="prev"`, link)
}

func (suite *ReportsGetTestSuite) TestReportsGetHighLimit() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]

	reports, link, err := suite.getReports(testAccount, testToken, testUser, http.StatusOK, "", nil, "", "", "", "", "", 2000)
	suite.NoError(err)
	suite.Len(reports, 2)

	// Limit in Link header should be set to 100
	suite.Equal(`<http://localhost:8080/api/v1/admin/reports?limit=100&max_id=01GP3AWY4CRDVRNZKW0TEAMB5R>; rel="next", <http://localhost:8080/api/v1/admin/reports?limit=100&min_id=01GP3DFY9XQ1TJMZT5BGAZPXX7>; rel="prev"`, link)
}

func TestReportsGetTestSuite(t *testing.T) {
	suite.Run(t, &ReportsGetTestSuite{})
}
