/*
   GoToSocial
   Copyright (C) GoToSocial Authors admin@gotosocial.org
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

package statuses_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusBoostTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusBoostTestSuite) postStatusBoost(
	targetStatusID string,
	app *gtsmodel.Application,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	account *gtsmodel.Account,
) (string, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, app)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedUser, user)
	ctx.Set(oauth.SessionAuthorizedAccount, account)

	const pathBase = "http://localhost:8080/api" + statuses.ReblogPath
	path := strings.ReplaceAll(pathBase, ":"+apiutil.IDKey, targetStatusID)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// Populate target status ID.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   apiutil.IDKey,
			Value: targetStatusID,
		},
	}

	// Trigger handler.
	suite.statusModule.StatusBoostPOSTHandler(ctx)
	return suite.parseStatusResponse(recorder)
}

func (suite *StatusBoostTestSuite) TestPostBoost() {
	var (
		targetStatus = suite.testStatuses["admin_account_status_1"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_1"]
		user         = suite.testUsers["local_account_1"]
		account      = suite.testAccounts["local_account_1"]
	)

	out, recorder := suite.postStatusBoost(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Target status should now
	// be "reblogged" by us.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": true,
  "card": null,
  "content": "",
  "created_at": "right the hell just now babyee",
  "emojis": [],
  "favourited": true,
  "favourites_count": 0,
  "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
  "in_reply_to_account_id": null,
  "in_reply_to_id": null,
  "interaction_policy": {
    "can_favourite": {
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
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  },
  "language": null,
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": {
    "account": "yeah this is my account, what about it punk",
    "application": {
      "name": "superseriousbusiness",
      "website": "https://superserious.business"
    },
    "bookmarked": true,
    "card": null,
    "content": "hello world! #welcome ! first post on the instance :rainbow: !",
    "created_at": "right the hell just now babyee",
    "emojis": [
      {
        "category": "reactions",
        "shortcode": "rainbow",
        "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
        "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
        "visible_in_picker": true
      }
    ],
    "favourited": true,
    "favourites_count": 1,
    "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
    "in_reply_to_account_id": null,
    "in_reply_to_id": null,
    "interaction_policy": {
      "can_favourite": {
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
      },
      "can_reply": {
        "always": [
          "public",
          "me"
        ],
        "with_approval": []
      }
    },
    "language": "en",
    "media_attachments": [
      {
        "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj",
        "description": "Black and white image of some 50's style text saying: Welcome On Board",
        "id": "01F8MH6NEM8D7527KZAECTCR76",
        "meta": {
          "focus": {
            "x": 0,
            "y": 0
          },
          "original": {
            "aspect": 1.9047619,
            "height": 630,
            "size": "1200x630",
            "width": 1200
          },
          "small": {
            "aspect": 1.9104477,
            "height": 268,
            "size": "512x268",
            "width": 512
          }
        },
        "preview_remote_url": null,
        "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.webp",
        "remote_url": null,
        "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
        "type": "image",
        "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg"
      }
    ],
    "mentions": [],
    "muted": false,
    "pinned": false,
    "poll": null,
    "reblog": null,
    "reblogged": true,
    "reblogs_count": 1,
    "replies_count": 1,
    "sensitive": false,
    "spoiler_text": "",
    "tags": [
      {
        "name": "welcome",
        "url": "http://localhost:8080/tags/welcome"
      }
    ],
    "text": "hello world! #welcome ! first post on the instance :rainbow: !",
    "uri": "http://localhost:8080/some/determinate/url",
    "url": "http://localhost:8080/some/determinate/url",
    "visibility": "public"
  },
  "reblogged": true,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusBoostTestSuite) TestPostBoostOwnFollowersOnly() {
	var (
		targetStatus = suite.testStatuses["local_account_1_status_5"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_1"]
		user         = suite.testUsers["local_account_1"]
		account      = suite.testAccounts["local_account_1"]
	)

	out, recorder := suite.postStatusBoost(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Target status should now
	// be "reblogged" by us.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "",
  "created_at": "right the hell just now babyee",
  "emojis": [],
  "favourited": false,
  "favourites_count": 0,
  "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
  "in_reply_to_account_id": null,
  "in_reply_to_id": null,
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "author",
        "followers",
        "mentioned",
        "me"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "author",
        "me"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "author",
        "followers",
        "mentioned",
        "me"
      ],
      "with_approval": []
    }
  },
  "language": null,
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": {
    "account": "yeah this is my account, what about it punk",
    "application": {
      "name": "really cool gts application",
      "website": "https://reallycool.app"
    },
    "bookmarked": false,
    "card": null,
    "content": "hi!",
    "created_at": "right the hell just now babyee",
    "emojis": [],
    "favourited": false,
    "favourites_count": 0,
    "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
    "in_reply_to_account_id": null,
    "in_reply_to_id": null,
    "interaction_policy": {
      "can_favourite": {
        "always": [
          "author",
          "followers",
          "mentioned",
          "me"
        ],
        "with_approval": []
      },
      "can_reblog": {
        "always": [
          "author",
          "me"
        ],
        "with_approval": []
      },
      "can_reply": {
        "always": [
          "author",
          "followers",
          "mentioned",
          "me"
        ],
        "with_approval": []
      }
    },
    "language": "en",
    "media_attachments": [],
    "mentions": [],
    "muted": false,
    "pinned": false,
    "poll": null,
    "reblog": null,
    "reblogged": true,
    "reblogs_count": 1,
    "replies_count": 0,
    "sensitive": false,
    "spoiler_text": "",
    "tags": [],
    "text": "hi!",
    "uri": "http://localhost:8080/some/determinate/url",
    "url": "http://localhost:8080/some/determinate/url",
    "visibility": "private"
  },
  "reblogged": true,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "private"
}`, out)
}

// Try to boost a status that's
// not boostable / visible to us.
func (suite *StatusBoostTestSuite) TestPostUnboostable() {
	var (
		targetStatus = suite.testStatuses["local_account_2_status_4"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_1"]
		user         = suite.testUsers["local_account_1"]
		account      = suite.testAccounts["local_account_1"]
	)

	out, recorder := suite.postStatusBoost(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have 403 from
	// our call to the function.
	suite.Equal(http.StatusForbidden, recorder.Code)

	// We should have a helpful message.
	suite.Equal(`{
  "error": "Forbidden: you do not have permission to boost this status"
}`, out)
}

// Try to boost a status that's not visible to the user.
func (suite *StatusBoostTestSuite) TestPostNotVisible() {
	// Stop local_account_2 following zork.
	err := suite.db.DeleteFollowByID(
		context.Background(),
		suite.testFollows["local_account_2_local_account_1"].ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	var (
		// This is a mutual only status and
		// these accounts aren't mutuals anymore.
		targetStatus = suite.testStatuses["local_account_1_status_3"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_2"]
		user         = suite.testUsers["local_account_2"]
		account      = suite.testAccounts["local_account_2"]
	)

	out, recorder := suite.postStatusBoost(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have 404 from
	// our call to the function.
	suite.Equal(http.StatusNotFound, recorder.Code)

	// We should have a helpful message.
	suite.Equal(`{
  "error": "Not Found: target status not found"
}`, out)
}

// Boost a status that's pending approval by us.
func (suite *StatusBoostTestSuite) TestPostBoostImplicitAccept() {
	var (
		targetStatus = suite.testStatuses["admin_account_status_5"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_2"]
		user         = suite.testUsers["local_account_2"]
		account      = suite.testAccounts["local_account_2"]
	)

	out, recorder := suite.postStatusBoost(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Target status should now
	// be "reblogged" by us.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "",
  "created_at": "right the hell just now babyee",
  "emojis": [],
  "favourited": false,
  "favourites_count": 0,
  "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
  "in_reply_to_account_id": null,
  "in_reply_to_id": null,
  "interaction_policy": {
    "can_favourite": {
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
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  },
  "language": null,
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": {
    "account": "yeah this is my account, what about it punk",
    "application": {
      "name": "superseriousbusiness",
      "website": "https://superserious.business"
    },
    "bookmarked": false,
    "card": null,
    "content": "<p>Hi <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span>, can I reply?</p>",
    "created_at": "right the hell just now babyee",
    "emojis": [],
    "favourited": false,
    "favourites_count": 0,
    "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
    "in_reply_to_account_id": "01F8MH5NBDF2MV7CTC4Q5128HF",
    "in_reply_to_id": "01F8MHC8VWDRBQR0N1BATDDEM5",
    "interaction_policy": {
      "can_favourite": {
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
      },
      "can_reply": {
        "always": [
          "public",
          "me"
        ],
        "with_approval": []
      }
    },
    "language": null,
    "media_attachments": [],
    "mentions": [
      {
        "acct": "1happyturtle",
        "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
        "url": "http://localhost:8080/@1happyturtle",
        "username": "1happyturtle"
      }
    ],
    "muted": false,
    "pinned": false,
    "poll": null,
    "reblog": null,
    "reblogged": true,
    "reblogs_count": 1,
    "replies_count": 0,
    "sensitive": false,
    "spoiler_text": "",
    "tags": [],
    "text": "Hi @1happyturtle, can I reply?",
    "uri": "http://localhost:8080/some/determinate/url",
    "url": "http://localhost:8080/some/determinate/url",
    "visibility": "unlisted"
  },
  "reblogged": true,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "unlisted"
}`, out)

	// Target status should no
	// longer be pending approval.
	dbStatus, err := suite.state.DB.GetStatusByID(
		context.Background(),
		targetStatus.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(*dbStatus.PendingApproval)

	// There should be an Accept
	// stored for the target status.
	intReq, err := suite.state.DB.GetInteractionRequestByInteractionURI(
		context.Background(), targetStatus.URI,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotZero(intReq.AcceptedAt)
	suite.NotEmpty(intReq.URI)
}

func TestStatusBoostTestSuite(t *testing.T) {
	suite.Run(t, new(StatusBoostTestSuite))
}
