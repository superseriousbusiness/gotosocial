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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
)

type AccountsGetTestSuite struct {
	AdminStandardTestSuite
}

func (suite *AccountsGetTestSuite) TestAccountsGetFromTop() {
	recorder := httptest.NewRecorder()

	path := admin.AccountsV2Path
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")

	suite.adminModule.AccountsGETV2Handler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(b)

	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	link := recorder.Header().Get("Link")
	suite.Equal(`<http://localhost:8080/api/v2/admin/accounts?limit=50&max_id=xn--xample-ova.org%2F%40%C3%BCser>; rel="next", <http://localhost:8080/api/v2/admin/accounts?limit=50&min_id=%2F%401happyturtle>; rel="prev"`, link)

	suite.Equal(`[
  {
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
      "note": "<p>i post about things that concern me</p>",
      "url": "http://localhost:8080/@1happyturtle",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 8,
      "last_status_at": "2021-07-28T08:40:37.000Z",
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
      "hide_collections": true
    },
    "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
  },
  {
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
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 4,
      "last_status_at": "2021-10-20T10:41:37.000Z",
      "emojis": [],
      "fields": [],
      "enable_rss": true,
      "roles": [
        {
          "id": "admin",
          "name": "admin",
          "color": ""
        }
      ]
    },
    "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
  },
  {
    "id": "01AY6P665V14JJR0AFVRT7311Y",
    "username": "localhost:8080",
    "domain": null,
    "created_at": "2020-05-17T13:10:59.000Z",
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
      "id": "01AY6P665V14JJR0AFVRT7311Y",
      "username": "localhost:8080",
      "acct": "localhost:8080",
      "display_name": "",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2020-05-17T13:10:59.000Z",
      "note": "",
      "url": "http://localhost:8080/@localhost:8080",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 0,
      "last_status_at": null,
      "emojis": [],
      "fields": []
    }
  },
  {
    "id": "01F8MH1H7YV1Z7D2C8K2730QBF",
    "username": "the_mighty_zork",
    "domain": null,
    "created_at": "2022-05-20T11:09:18.000Z",
    "email": "zork@example.org",
    "ip": null,
    "ips": [],
    "locale": "en",
    "invite_request": "I wanna be on this damned webbed site so bad! Please! Wow",
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
      "id": "01F8MH1H7YV1Z7D2C8K2730QBF",
      "username": "the_mighty_zork",
      "acct": "the_mighty_zork",
      "display_name": "original zork (he/they)",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2022-05-20T11:09:18.000Z",
      "note": "<p>hey yo this is my profile!</p>",
      "url": "http://localhost:8080/@the_mighty_zork",
      "avatar": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg",
      "avatar_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp",
      "avatar_description": "a green goblin looking nasty",
      "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
      "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
      "header_description": "A very old-school screenshot of the original team fortress mod for quake",
      "followers_count": 2,
      "following_count": 2,
      "statuses_count": 8,
      "last_status_at": "2024-01-10T09:24:00.000Z",
      "emojis": [],
      "fields": [],
      "enable_rss": true
    },
    "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
  },
  {
    "id": "01F8MH0BBE4FHXPH513MBVFHB0",
    "username": "weed_lord420",
    "domain": null,
    "created_at": "2022-06-04T13:12:00.000Z",
    "email": "weed_lord420@example.org",
    "ip": "199.222.111.89",
    "ips": [],
    "locale": "en",
    "invite_request": "hi, please let me in! I'm looking for somewhere neato bombeato to hang out.",
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
      "id": "01F8MH0BBE4FHXPH513MBVFHB0",
      "username": "weed_lord420",
      "acct": "weed_lord420",
      "display_name": "",
      "locked": false,
      "discoverable": false,
      "bot": false,
      "created_at": "2022-06-04T13:12:00.000Z",
      "note": "",
      "url": "http://localhost:8080/@weed_lord420",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 0,
      "last_status_at": null,
      "emojis": [],
      "fields": []
    },
    "created_by_application_id": "01F8MGY43H3N2C8EWPR2FPYEXG"
  },
  {
    "id": "01FHMQX3GAABWSM0S2VZEC2SWC",
    "username": "Some_User",
    "domain": "example.org",
    "created_at": "2020-08-10T12:13:28.000Z",
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
      "id": "01FHMQX3GAABWSM0S2VZEC2SWC",
      "username": "Some_User",
      "acct": "Some_User@example.org",
      "display_name": "some user",
      "locked": true,
      "discoverable": true,
      "bot": false,
      "created_at": "2020-08-10T12:13:28.000Z",
      "note": "i'm a real son of a gun",
      "url": "http://example.org/@Some_User",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 1,
      "last_status_at": "2023-11-02T10:44:25.000Z",
      "emojis": [],
      "fields": []
    }
  },
  {
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
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 3,
      "last_status_at": "2021-09-11T09:40:37.000Z",
      "emojis": [],
      "fields": []
    }
  },
  {
    "id": "062G5WYKY35KKD12EMSM3F8PJ8",
    "username": "her_fuckin_maj",
    "domain": "thequeenisstillalive.technology",
    "created_at": "2020-08-10T12:13:28.000Z",
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
      "id": "062G5WYKY35KKD12EMSM3F8PJ8",
      "username": "her_fuckin_maj",
      "acct": "her_fuckin_maj@thequeenisstillalive.technology",
      "display_name": "lizzzieeeeeeeeeeee",
      "locked": true,
      "discoverable": true,
      "bot": false,
      "created_at": "2020-08-10T12:13:28.000Z",
      "note": "if i die blame charles don't let that fuck become king",
      "url": "http://thequeenisstillalive.technology/@her_fuckin_maj",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/fileserver/062G5WYKY35KKD12EMSM3F8PJ8/header/original/01PFPMWK2FF0D9WMHEJHR07C3R.jpg",
      "header_static": "http://localhost:8080/fileserver/062G5WYKY35KKD12EMSM3F8PJ8/header/small/01PFPMWK2FF0D9WMHEJHR07C3R.webp",
      "header_description": "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 0,
      "last_status_at": null,
      "emojis": [],
      "fields": []
    }
  },
  {
    "id": "07GZRBAEMBNKGZ8Z9VSKSXKR98",
    "username": "üser",
    "domain": "ëxample.org",
    "created_at": "2020-08-10T12:13:28.000Z",
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
      "id": "07GZRBAEMBNKGZ8Z9VSKSXKR98",
      "username": "üser",
      "acct": "üser@ëxample.org",
      "display_name": "",
      "locked": false,
      "discoverable": false,
      "bot": false,
      "created_at": "2020-08-10T12:13:28.000Z",
      "note": "",
      "url": "https://xn--xample-ova.org/users/@%C3%BCser",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 0,
      "last_status_at": null,
      "emojis": [],
      "fields": []
    }
  }
]`, dst.String())
}

func (suite *AccountsGetTestSuite) TestAccountsMinID() {
	recorder := httptest.NewRecorder()

	path := admin.AccountsV2Path + "?limit=1&min_id=/@the_mighty_zork"
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")

	ctx.Params = gin.Params{
		{
			Key:   "min_id",
			Value: "/@the_mighty_zork",
		},
		{
			Key:   "limit",
			Value: "1",
		},
	}

	suite.adminModule.AccountsGETV2Handler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(b)

	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	link := recorder.Header().Get("Link")
	suite.Equal(`<http://localhost:8080/api/v2/admin/accounts?limit=1&max_id=%2F%40localhost%3A8080>; rel="next", <http://localhost:8080/api/v2/admin/accounts?limit=1&min_id=%2F%40localhost%3A8080>; rel="prev"`, link)

	suite.Equal(`[
  {
    "id": "01AY6P665V14JJR0AFVRT7311Y",
    "username": "localhost:8080",
    "domain": null,
    "created_at": "2020-05-17T13:10:59.000Z",
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
      "id": "01AY6P665V14JJR0AFVRT7311Y",
      "username": "localhost:8080",
      "acct": "localhost:8080",
      "display_name": "",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2020-05-17T13:10:59.000Z",
      "note": "",
      "url": "http://localhost:8080/@localhost:8080",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 0,
      "last_status_at": null,
      "emojis": [],
      "fields": []
    }
  }
]`, dst.String())
}

func TestAccountsGetTestSuite(t *testing.T) {
	suite.Run(t, &AccountsGetTestSuite{})
}
