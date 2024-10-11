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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusHistoryTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusHistoryTestSuite) TestGetHistory() {
	var (
		testApplication = suite.testApplications["application_1"]
		testAccount     = suite.testAccounts["local_account_1"]
		testUser        = suite.testUsers["local_account_1"]
		testToken       = oauth.DBTokenToToken(suite.testTokens["local_account_1"])
		targetStatusID  = suite.testStatuses["local_account_1_status_1"].ID
		target          = fmt.Sprintf("http://localhost:8080%s", strings.ReplaceAll(statuses.HistoryPath, ":id", targetStatusID))
	)

	// Setup request.
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("accept", "application/json")
	ctx, _ := testrig.CreateGinTestContext(recorder, request)

	// Set auth + path params.
	ctx.Set(oauth.SessionAuthorizedApplication, testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, testToken)
	ctx.Set(oauth.SessionAuthorizedUser, testUser)
	ctx.Set(oauth.SessionAuthorizedAccount, testAccount)
	ctx.Params = gin.Params{
		gin.Param{
			Key:   statuses.IDKey,
			Value: targetStatusID,
		},
	}

	// Call the handler.
	suite.statusModule.StatusHistoryGETHandler(ctx)

	// Check code.
	if code := recorder.Code; code != http.StatusOK {
		suite.FailNow("", "unexpected http code: %d", code)
	}

	// Read body.
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Indent nicely.
	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`[
  {
    "content": "hello everyone!",
    "spoiler_text": "introduction post",
    "sensitive": true,
    "created_at": "2021-10-20T10:40:37.000Z",
    "account": {
      "id": "01F8MH1H7YV1Z7D2C8K2730QBF",
      "username": "the_mighty_zork",
      "acct": "the_mighty_zork",
      "display_name": "original zork (he/they)",
      "locked": false,
      "discoverable": true,
      "bot": false,
      "created_at": "2022-05-20T11:09:18.000Z",
      "note": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
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
      "last_status_at": "2024-01-10",
      "emojis": [],
      "fields": [],
      "enable_rss": true
    },
    "poll": null,
    "media_attachments": [],
    "emojis": []
  }
]`, dst.String())
}

func TestStatusHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(StatusHistoryTestSuite))
}
