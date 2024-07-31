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
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusMuteTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusMuteTestSuite) post(path string, handler func(*gin.Context), targetStatusID string) (int, string) {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, path, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   apiutil.IDKey,
			Value: targetStatusID,
		},
	}

	handler(ctx)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	indented := bytes.Buffer{}
	if err := json.Indent(&indented, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	return recorder.Code, indented.String()
}

func (suite *StatusMuteTestSuite) TestMuteUnmuteStatus() {
	var (
		targetStatus = suite.testStatuses["local_account_1_status_1"]
		path         = fmt.Sprintf("http://localhost:8080/api%s", strings.ReplaceAll(statuses.MutePath, ":id", targetStatus.ID))
	)

	// Mute the status, ensure `muted` is `true`.
	code, muted := suite.post(path, suite.statusModule.StatusMutePOSTHandler, targetStatus.ID)
	suite.Equal(http.StatusOK, code)
	suite.Equal(`{
  "id": "01F8MHAMCHF6Y650WCRSCP4WMY",
  "created_at": "2021-10-20T10:40:37.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": true,
  "spoiler_text": "introduction post",
  "visibility": "public",
  "language": "en",
  "uri": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "replies_count": 2,
  "reblogs_count": 1,
  "favourites_count": 1,
  "favourited": false,
  "reblogged": false,
  "muted": true,
  "bookmarked": false,
  "pinned": false,
  "content": "hello everyone!",
  "reblog": null,
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
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
    "last_status_at": "2024-01-10T09:24:00.000Z",
    "emojis": [],
    "fields": [],
    "enable_rss": true
  },
  "media_attachments": [],
  "mentions": [],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
  "text": "hello everyone!",
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  }
}`, muted)

	// Unmute the status, ensure `muted` is `false`.
	code, unmuted := suite.post(path, suite.statusModule.StatusUnmutePOSTHandler, targetStatus.ID)
	suite.Equal(http.StatusOK, code)
	suite.Equal(`{
  "id": "01F8MHAMCHF6Y650WCRSCP4WMY",
  "created_at": "2021-10-20T10:40:37.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": true,
  "spoiler_text": "introduction post",
  "visibility": "public",
  "language": "en",
  "uri": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "replies_count": 2,
  "reblogs_count": 1,
  "favourites_count": 1,
  "favourited": false,
  "reblogged": false,
  "muted": false,
  "bookmarked": false,
  "pinned": false,
  "content": "hello everyone!",
  "reblog": null,
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
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
    "last_status_at": "2024-01-10T09:24:00.000Z",
    "emojis": [],
    "fields": [],
    "enable_rss": true
  },
  "media_attachments": [],
  "mentions": [],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
  "text": "hello everyone!",
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  }
}`, unmuted)
}

func TestStatusMuteTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMuteTestSuite))
}
