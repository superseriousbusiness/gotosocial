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
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type StatusCreateTestSuite struct {
	StatusStandardTestSuite
}

const (
	statusWithLinksAndTags = "#test alright, should be able to post #links with fragments in them now, let's see........\n\nhttps://docs.gotosocial.org/en/latest/user_guide/posts/#links\n\n#gotosocial\n\n(tobi remember to pull the docker image challenge)"
	statusMarkdown         = "# Title\n\n## Smaller title\n\nThis is a post written in [markdown](https://www.markdownguide.org/)\n\n<img src=\"https://d33wubrfki0l68.cloudfront.net/f1f475a6fda1c2c4be4cac04033db5c3293032b4/513a4/assets/images/markdown-mark-white.svg\"/>"
)

// Post a status.
func (suite *StatusCreateTestSuite) postStatusCore(
	formData map[string][]string,
	jsonData string,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	if formData != nil {
		buf, w, err := testrig.CreateMultipartFormData(nil, formData)
		if err != nil {
			suite.FailNow(err.Error())
		}

		ctx.Request = httptest.NewRequest(
			http.MethodPost,
			"http://localhost:8080"+statuses.BasePath,
			bytes.NewReader(buf.Bytes()),
		)
		ctx.Request.Header.Set("content-type", w.FormDataContentType())
	} else {
		ctx.Request = httptest.NewRequest(
			http.MethodPost,
			"http://localhost:8080"+statuses.BasePath,
			bytes.NewReader([]byte(jsonData)),
		)
		ctx.Request.Header.Set("content-type", "application/json")
	}

	ctx.Request.Header.Set("accept", "application/json")

	// Trigger handler.
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	return recorder
}

// Post a status and return the result as deterministic JSON.
func (suite *StatusCreateTestSuite) postStatus(
	formData map[string][]string,
	jsonData string,
) (string, *httptest.ResponseRecorder) {
	recorder := suite.postStatusCore(formData, jsonData)
	return suite.parseStatusResponse(recorder)
}

// Post a status and return the result as a non-deterministic API structure.
func (suite *StatusCreateTestSuite) postStatusStruct(
	formData map[string][]string,
	jsonData string,
) (*apimodel.Status, *httptest.ResponseRecorder) {
	recorder := suite.postStatusCore(formData, jsonData)

	result := recorder.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	apiStatus := apimodel.Status{}
	if err := json.Unmarshal(data, &apiStatus); err != nil {
		suite.FailNow(err.Error())
	}

	return &apiStatus, recorder
}

// Post a new status with some custom visibility settings
func (suite *StatusCreateTestSuite) TestPostNewStatus() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":       {"this is a brand new status! #helloworld"},
		"spoiler_text": {"hello hello"},
		"sensitive":    {"true"},
		"visibility":   {string(apimodel.VisibilityMutualsOnly)},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>this is a brand new status! <a href=\"http://localhost:8080/tags/helloworld\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>helloworld</span></a></p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "author",
        "followers",
        "mentioned",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "author",
        "me"
      ],
      "automatic_approval": [
        "author",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "author",
        "followers",
        "mentioned",
        "me"
      ],
      "automatic_approval": [
        "author",
        "followers",
        "mentioned",
        "me"
      ],
      "manual_approval": [],
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
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": true,
  "spoiler_text": "hello hello",
  "tags": [
    {
      "name": "helloworld",
      "url": "http://localhost:8080/tags/helloworld"
    }
  ],
  "text": "this is a brand new status! #helloworld",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "private"
}`, out)
}

// Post a new status with some custom visibility settings
func (suite *StatusCreateTestSuite) TestPostNewStatusIntPolicy() {
	out, recorder := suite.postStatus(map[string][]string{
		"status": {"this is a brand new status! #helloworld"},
		"interaction_policy[can_reply][always][0]":        {"author"},
		"interaction_policy[can_reply][always][1]":        {"followers"},
		"interaction_policy[can_reply][always][2]":        {"following"},
		"interaction_policy[can_reply][with_approval][0]": {"public"},
		"interaction_policy[can_announce][always][0]":     {""},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Custom interaction policies
	// should be set on the status.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>this is a brand new status! <a href=\"http://localhost:8080/tags/helloworld\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>helloworld</span></a></p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
        "me"
      ],
      "automatic_approval": [
        "author",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "author",
        "me"
      ],
      "automatic_approval": [
        "author",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "author",
        "followers",
        "following",
        "mentioned",
        "me"
      ],
      "automatic_approval": [
        "author",
        "followers",
        "following",
        "mentioned",
        "me"
      ],
      "manual_approval": [
        "public"
      ],
      "with_approval": [
        "public"
      ]
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [
    {
      "name": "helloworld",
      "url": "http://localhost:8080/tags/helloworld"
    }
  ],
  "text": "this is a brand new status! #helloworld",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusIntPolicyJSON() {
	out, recorder := suite.postStatus(nil, `{
  "status": "this is a brand new status! #helloworld",
  "interaction_policy": {
    "can_reply": {
      "always": [
        "author",
        "followers",
        "following"
      ],
      "with_approval": [
        "public"
      ]
    },
    "can_announce": {
      "always": []
    }
  }
}`)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Custom interaction policies
	// should be set on the status.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>this is a brand new status! <a href=\"http://localhost:8080/tags/helloworld\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>helloworld</span></a></p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
        "me"
      ],
      "automatic_approval": [
        "author",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "author",
        "me"
      ],
      "automatic_approval": [
        "author",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "author",
        "followers",
        "following",
        "mentioned",
        "me"
      ],
      "automatic_approval": [
        "author",
        "followers",
        "following",
        "mentioned",
        "me"
      ],
      "manual_approval": [
        "public"
      ],
      "with_approval": [
        "public"
      ]
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [
    {
      "name": "helloworld",
      "url": "http://localhost:8080/tags/helloworld"
    }
  ],
  "text": "this is a brand new status! #helloworld",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusMessedUpIntPolicy() {
	out, recorder := suite.postStatus(nil, `{
  "status": "this is a brand new status! #helloworld",
  "visibility": "private",
  "interaction_policy": {
    "can_reply": {
      "always": [
        "public"
      ]
    }
  }
}`)

	// We should have 400 from
	// our call to the function.
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// We should have a helpful error
	// message telling us how we screwed up.
	suite.Equal(`{
  "error": "Bad Request: error converting private.can_reply.always: policyURI public is not feasible for visibility private"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewScheduledStatus() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":       {"this is a brand new status! #helloworld"},
		"spoiler_text": {"hello hello"},
		"sensitive":    {"true"},
		"visibility":   {string(apimodel.VisibilityMutualsOnly)},
		"scheduled_at": {"2080-10-04T15:32:02.018Z"},
	}, "")

	// We should have 501 from
	// our call to the function.
	suite.Equal(http.StatusNotImplemented, recorder.Code)

	// We should have a helpful error message.
	suite.Equal(`{
  "error": "Not Implemented: scheduled statuses are not yet supported"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewBackfilledStatus() {
	// A time in the past.
	scheduledAtStr := "2020-10-04T15:32:02.018Z"
	scheduledAt, err := time.Parse(time.RFC3339Nano, scheduledAtStr)
	if err != nil {
		suite.FailNow(err.Error())
	}

	status, recorder := suite.postStatusStruct(map[string][]string{
		"status":       {"this is a recycled status from the past!"},
		"scheduled_at": {scheduledAtStr},
	}, "")

	// Creating a status in the past should succeed.
	suite.Equal(http.StatusOK, recorder.Code)

	// The status should be backdated.
	createdAt, err := time.Parse(time.RFC3339Nano, status.CreatedAt)
	if err != nil {
		suite.FailNow(err.Error())
		return
	}
	suite.Equal(scheduledAt, createdAt.UTC())

	// The status's ULID should be backdated.
	timeFromULID, err := id.TimeFromULID(status.ID)
	if err != nil {
		suite.FailNow(err.Error())
		return
	}
	suite.Equal(scheduledAt, timeFromULID.UTC())
}

func (suite *StatusCreateTestSuite) TestPostNewBackfilledStatusWithSelfMention() {
	_, recorder := suite.postStatus(map[string][]string{
		"status":       {"@the_mighty_zork this is a recycled mention from the past!"},
		"scheduled_at": {"2020-10-04T15:32:02.018Z"},
	}, "")

	// Mentioning yourself is allowed in backfilled statuses.
	suite.Equal(http.StatusOK, recorder.Code)
}

func (suite *StatusCreateTestSuite) TestPostNewBackfilledStatusWithMention() {
	_, recorder := suite.postStatus(map[string][]string{
		"status":       {"@admin this is a recycled mention from the past!"},
		"scheduled_at": {"2020-10-04T15:32:02.018Z"},
	}, "")

	// Mentioning others is forbidden in backfilled statuses.
	suite.Equal(http.StatusForbidden, recorder.Code)
}

func (suite *StatusCreateTestSuite) TestPostNewBackfilledStatusWithSelfReply() {
	_, recorder := suite.postStatus(map[string][]string{
		"status":         {"this is a recycled reply from the past!"},
		"scheduled_at":   {"2020-10-04T15:32:02.018Z"},
		"in_reply_to_id": {suite.testStatuses["local_account_1_status_1"].ID},
	}, "")

	// Replying to yourself is allowed in backfilled statuses.
	suite.Equal(http.StatusOK, recorder.Code)
}

func (suite *StatusCreateTestSuite) TestPostNewBackfilledStatusWithReply() {
	_, recorder := suite.postStatus(map[string][]string{
		"status":         {"this is a recycled reply from the past!"},
		"scheduled_at":   {"2020-10-04T15:32:02.018Z"},
		"in_reply_to_id": {suite.testStatuses["admin_account_status_1"].ID},
	}, "")

	// Replying to others is forbidden in backfilled statuses.
	suite.Equal(http.StatusForbidden, recorder.Code)
}

func (suite *StatusCreateTestSuite) TestPostNewBackfilledStatusWithPoll() {
	_, recorder := suite.postStatus(map[string][]string{
		"status":           {"this is a recycled poll from the past!"},
		"scheduled_at":     {"2020-10-04T15:32:02.018Z"},
		"poll[options][]":  {"first option", "second option"},
		"poll[expires_in]": {"3600"},
		"poll[multiple]":   {"true"},
	}, "")

	// Polls are forbidden in backfilled statuses.
	suite.Equal(http.StatusForbidden, recorder.Code)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusMarkdown() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":       {statusMarkdown},
		"visibility":   {string(apimodel.VisibilityPublic)},
		"content_type": {"text/markdown"},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// The content field should have
	// all the nicely parsed markdown stuff.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<h1>Title</h1><h2>Smaller title</h2><p>This is a post written in <a href=\"https://www.markdownguide.org/\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">markdown</a></p>",
  "content_type": "text/markdown",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
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
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "# Title\n\n## Smaller title\n\nThis is a post written in [markdown](https://www.markdownguide.org/)\n\n<img src=\"https://d33wubrfki0l68.cloudfront.net/f1f475a6fda1c2c4be4cac04033db5c3293032b4/513a4/assets/images/markdown-mark-white.svg\"/>",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

// Mention an account that is not yet known to the
// instance -- it should be looked up and put in the db.
func (suite *StatusCreateTestSuite) TestMentionUnknownAccount() {
	// First remove remote account 1 from the database
	// so it gets looked up again when we mention it.
	remoteAccount := suite.testAccounts["remote_account_1"]
	if err := suite.db.DeleteAccount(
		suite.T().Context(),
		remoteAccount.ID,
	); err != nil {
		suite.FailNow(err.Error())
	}

	out, recorder := suite.postStatus(map[string][]string{
		"status":     {"hello @brand_new_person@unknown-instance.com"},
		"visibility": {string(apimodel.VisibilityPublic)},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Status should have a mention of
	// the now-freshly-looked-up account.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>hello <span class=\"h-card\"><a href=\"https://unknown-instance.com/@brand_new_person\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>brand_new_person</span></a></span></p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [
    {
      "acct": "brand_new_person@unknown-instance.com",
      "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
      "url": "https://unknown-instance.com/@brand_new_person",
      "username": "brand_new_person"
    }
  ],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "hello @brand_new_person@unknown-instance.com",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostStatusWithLinksAndTags() {
	out, recorder := suite.postStatus(map[string][]string{
		"status": {statusWithLinksAndTags},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Status should have proper
	// tags + formatted links.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p><a href=\"http://localhost:8080/tags/test\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>test</span></a> alright, should be able to post <a href=\"http://localhost:8080/tags/links\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>links</span></a> with fragments in them now, let's see........<br><br><a href=\"https://docs.gotosocial.org/en/latest/user_guide/posts/#links\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://docs.gotosocial.org/en/latest/user_guide/posts/#links</a><br><br><a href=\"http://localhost:8080/tags/gotosocial\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>gotosocial</span></a><br><br>(tobi remember to pull the docker image challenge)</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
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
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [
    {
      "name": "test",
      "url": "http://localhost:8080/tags/test"
    },
    {
      "name": "links",
      "url": "http://localhost:8080/tags/links"
    },
    {
      "name": "gotosocial",
      "url": "http://localhost:8080/tags/gotosocial"
    }
  ],
  "text": "#test alright, should be able to post #links with fragments in them now, let's see........\n\nhttps://docs.gotosocial.org/en/latest/user_guide/posts/#links\n\n#gotosocial\n\n(tobi remember to pull the docker image challenge)",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusWithEmoji() {
	out, recorder := suite.postStatus(map[string][]string{
		"status": {"here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: \n here's an emoji that isn't in the db: :test_emoji: "},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Emojis array should be
	// populated on returned status.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow:<br>here's an emoji that isn't in the db: :test_emoji:</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
  "emojis": [
    {
      "category": "reactions",
      "shortcode": "rainbow",
      "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "visible_in_picker": true
    }
  ],
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
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
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: \n here's an emoji that isn't in the db: :test_emoji: ",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

// Try to reply to a status that doesn't exist
func (suite *StatusCreateTestSuite) TestReplyToNonexistentStatus() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":         {"this is a reply to a status that doesn't exist"},
		"spoiler_text":   {"don't open cuz it won't work"},
		"in_reply_to_id": {"3759e7ef-8ee1-4c0c-86f6-8b70b9ad3d50"},
	}, "")

	// We should have 404 from
	// our call to the function.
	suite.Equal(http.StatusNotFound, recorder.Code)
	suite.Equal(`{
  "error": "Not Found: target status not found"
}`, out)
}

// Post a reply to the status of
// a local user that allows replies.
func (suite *StatusCreateTestSuite) TestReplyToLocalStatus() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":         {fmt.Sprintf("hello @%s this reply should work!", testrig.NewTestAccounts()["local_account_2"].Username)},
		"in_reply_to_id": {testrig.NewTestStatuses()["local_account_2_status_1"].ID},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// in_reply_to_x
	// fields should be set.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>hello <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> this reply should work!</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
  "emojis": [],
  "favourited": false,
  "favourites_count": 0,
  "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
  "in_reply_to_account_id": "01F8MH5NBDF2MV7CTC4Q5128HF",
  "in_reply_to_id": "01F8MHBQCBTDKN6X5VHGMMN4MA",
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    }
  },
  "language": "en",
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
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "hello @1happyturtle this reply should work!",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

// Take a media file which is currently not associated
// with a status, and attach it to a new status.
func (suite *StatusCreateTestSuite) TestAttachNewMediaSuccess() {
	attachment := suite.testAttachments["local_account_1_unattached_1"]

	out, recorder := suite.postStatus(map[string][]string{
		"status":      {"here's an image attachment"},
		"media_ids[]": {attachment.ID},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Status should have
	// media attached.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>here's an image attachment</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    }
  },
  "language": "en",
  "media_attachments": [
    {
      "blurhash": "LNABP8o#Dge,S6M}axxVEQjYxWbH",
      "description": "the oh you meme",
      "id": "01F8MH8RMYQ6MSNY3JM2XT1CQ5",
      "meta": {
        "focus": {
          "x": 0,
          "y": 0
        },
        "original": {
          "aspect": 1.7777778,
          "height": 450,
          "size": "800x450",
          "width": 800
        },
        "small": {
          "aspect": 1.7777778,
          "height": 288,
          "size": "512x288",
          "width": 512
        }
      },
      "preview_remote_url": null,
      "preview_url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.webp",
      "remote_url": null,
      "text_url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpg",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpg"
    }
  ],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "here's an image attachment",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

// Post a new status with a language
// tag that is not in canonical format.
func (suite *StatusCreateTestSuite) TestPostNewStatusWithNoncanonicalLanguageTag() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":   {"English? what's English? i speak American"},
		"language": {"en-us"},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// The returned language tag should
	// use its canonicalized version rather
	// than the format we submitted.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>English? what's English? i speak American</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    }
  },
  "language": "en-US",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "English? what's English? i speak American",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusWithPollForm() {
	out, recorder := suite.postStatus(map[string][]string{
		"status":           {"this is a status with a poll!"},
		"visibility":       {"public"},
		"poll[options][]":  {"first option", "second option"},
		"poll[expires_in]": {"3600"},
		"poll[multiple]":   {"true"},
	}, "")

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Status poll should
	// be as expected.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>this is a status with a poll!</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": {
    "emojis": [],
    "expired": false,
    "expires_at": "ah like you know whatever dude it's chill",
    "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
    "multiple": true,
    "options": [
      {
        "title": "first option",
        "votes_count": 0
      },
      {
        "title": "second option",
        "votes_count": 0
      }
    ],
    "own_votes": [],
    "voted": true,
    "voters_count": 0,
    "votes_count": 0
  },
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "this is a status with a poll!",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusWithPollJSON() {
	out, recorder := suite.postStatus(nil, `{
  "status": "this is a status with a poll!",
  "visibility": "public",
  "poll": {
    "options": ["first option", "second option"],
    "expires_in": 3600,
    "multiple": true
  }
}`)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Status poll should
	// be as expected.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>this is a status with a poll!</p>",
  "content_type": "text/plain",
  "created_at": "right the hell just now babyee",
  "edited_at": null,
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
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": [],
      "with_approval": []
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": {
    "emojis": [],
    "expired": false,
    "expires_at": "ah like you know whatever dude it's chill",
    "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
    "multiple": true,
    "options": [
      {
        "title": "first option",
        "votes_count": 0
      },
      {
        "title": "second option",
        "votes_count": 0
      }
    ],
    "own_votes": [],
    "voted": true,
    "voters_count": 0,
    "votes_count": 0
  },
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "this is a status with a poll!",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}
