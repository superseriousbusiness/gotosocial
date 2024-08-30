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

package instance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InstancePatchTestSuite struct {
	InstanceStandardTestSuite
}

func (suite *InstancePatchTestSuite) instancePatch(fieldName string, fileName string, extraFields map[string][]string) (code int, body []byte) {
	var dataF testrig.DataF
	if fieldName != "" && fileName != "" {
		dataF = testrig.FileToDataF(fieldName, fileName)
	}

	requestBody, w, err := testrig.CreateMultipartFormData(dataF, extraFields)
	if err != nil {
		suite.FailNow(err.Error())
	}

	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPathV1, requestBody.Bytes(), w.FormDataContentType(), true)

	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return recorder.Code, b
}

func (suite *InstancePatchTestSuite) TestInstancePatch1() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"title":            {"Example Instance"},
		"contact_username": {"admin"},
		"contact_email":    {"someone@example.org"},
	})

	if expectedCode := http.StatusOK; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "Example Instance",
  "description": "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "<p>This is the GoToSocial testrig. It doesn't federate or anything.</p><p>When the testrig is shut down, all data on it will be deleted.</p><p>Don't use this in production!</p>",
  "short_description_text": "This is the GoToSocial testrig. It doesn't federate or anything.\n\nWhen the testrig is shut down, all data on it will be deleted.\n\nDon't use this in production!",
  "email": "someone@example.org",
  "version": "0.0.0-testrig",
  "languages": [
    "nl",
    "en-gb"
  ],
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25,
      "supported_mime_types": [
        "text/plain",
        "text/markdown"
      ]
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/webp",
        "audio/mp2",
        "audio/mp3",
        "video/x-msvideo",
        "audio/flac",
        "audio/x-flac",
        "image/png",
        "image/apng",
        "audio/ogg",
        "video/ogg",
        "audio/x-m4a",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 16777216,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 60,
      "video_matrix_limit": 16777216
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "urls": {
    "streaming_api": "wss://localhost:8080"
  },
  "stats": {
    "domain_count": 2,
    "status_count": 20,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/assets/logo.webp",
  "contact_account": {
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
  "max_toot_chars": 5000,
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
  "terms": "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, dst.String())
}

func (suite *InstancePatchTestSuite) TestInstancePatch2() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"title": {"<p>Geoff's Instance</p>"},
	})

	if expectedCode := http.StatusOK; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "Geoff's Instance",
  "description": "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "<p>This is the GoToSocial testrig. It doesn't federate or anything.</p><p>When the testrig is shut down, all data on it will be deleted.</p><p>Don't use this in production!</p>",
  "short_description_text": "This is the GoToSocial testrig. It doesn't federate or anything.\n\nWhen the testrig is shut down, all data on it will be deleted.\n\nDon't use this in production!",
  "email": "admin@example.org",
  "version": "0.0.0-testrig",
  "languages": [
    "nl",
    "en-gb"
  ],
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25,
      "supported_mime_types": [
        "text/plain",
        "text/markdown"
      ]
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/webp",
        "audio/mp2",
        "audio/mp3",
        "video/x-msvideo",
        "audio/flac",
        "audio/x-flac",
        "image/png",
        "image/apng",
        "audio/ogg",
        "video/ogg",
        "audio/x-m4a",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 16777216,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 60,
      "video_matrix_limit": 16777216
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "urls": {
    "streaming_api": "wss://localhost:8080"
  },
  "stats": {
    "domain_count": 2,
    "status_count": 20,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/assets/logo.webp",
  "contact_account": {
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
  "max_toot_chars": 5000,
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
  "terms": "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, dst.String())
}

func (suite *InstancePatchTestSuite) TestInstancePatch3() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"short_description": {"This is some html, which is <em>allowed</em> in short descriptions."},
	})

	if expectedCode := http.StatusOK; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "description": "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "<p>This is some html, which is <em>allowed</em> in short descriptions.</p>",
  "short_description_text": "This is some html, which is <em>allowed</em> in short descriptions.",
  "email": "admin@example.org",
  "version": "0.0.0-testrig",
  "languages": [
    "nl",
    "en-gb"
  ],
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25,
      "supported_mime_types": [
        "text/plain",
        "text/markdown"
      ]
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/webp",
        "audio/mp2",
        "audio/mp3",
        "video/x-msvideo",
        "audio/flac",
        "audio/x-flac",
        "image/png",
        "image/apng",
        "audio/ogg",
        "video/ogg",
        "audio/x-m4a",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 16777216,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 60,
      "video_matrix_limit": 16777216
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "urls": {
    "streaming_api": "wss://localhost:8080"
  },
  "stats": {
    "domain_count": 2,
    "status_count": 20,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/assets/logo.webp",
  "contact_account": {
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
  "max_toot_chars": 5000,
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
  "terms": "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, dst.String())
}

func (suite *InstancePatchTestSuite) TestInstancePatch4() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"": {""},
	})

	if expectedCode := http.StatusBadRequest; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{"error":"Bad Request: empty form submitted"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch5() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"short_description": {"<p>This is some html, which is <em>allowed</em> in short descriptions.</p>"},
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, instance.InstanceInformationPathV1, bodyBytes, w.FormDataContentType(), true)

	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	suite.Equal(http.StatusForbidden, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Forbidden: user is not an admin so cannot update instance settings"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch6() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"contact_email": {""},
	})

	if expectedCode := http.StatusOK; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "description": "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "<p>This is the GoToSocial testrig. It doesn't federate or anything.</p><p>When the testrig is shut down, all data on it will be deleted.</p><p>Don't use this in production!</p>",
  "short_description_text": "This is the GoToSocial testrig. It doesn't federate or anything.\n\nWhen the testrig is shut down, all data on it will be deleted.\n\nDon't use this in production!",
  "email": "",
  "version": "0.0.0-testrig",
  "languages": [
    "nl",
    "en-gb"
  ],
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25,
      "supported_mime_types": [
        "text/plain",
        "text/markdown"
      ]
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/webp",
        "audio/mp2",
        "audio/mp3",
        "video/x-msvideo",
        "audio/flac",
        "audio/x-flac",
        "image/png",
        "image/apng",
        "audio/ogg",
        "video/ogg",
        "audio/x-m4a",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 16777216,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 60,
      "video_matrix_limit": 16777216
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "urls": {
    "streaming_api": "wss://localhost:8080"
  },
  "stats": {
    "domain_count": 2,
    "status_count": 20,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/assets/logo.webp",
  "contact_account": {
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
  "max_toot_chars": 5000,
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
  "terms": "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, dst.String())
}

func (suite *InstancePatchTestSuite) TestInstancePatch7() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"contact_email": {"not.an.email.address"},
	})

	if expectedCode := http.StatusBadRequest; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{"error":"Bad Request: mail: missing '@' or angle-addr"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch8() {
	code, b := suite.instancePatch("thumbnail", "../../../../testrig/media/peglin.gif", map[string][]string{
		"thumbnail_description": {"A bouncing little green peglin."},
	})

	if expectedCode := http.StatusOK; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	instanceAccount, err := suite.db.GetInstanceAccount(context.Background(), "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "description": "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "<p>This is the GoToSocial testrig. It doesn't federate or anything.</p><p>When the testrig is shut down, all data on it will be deleted.</p><p>Don't use this in production!</p>",
  "short_description_text": "This is the GoToSocial testrig. It doesn't federate or anything.\n\nWhen the testrig is shut down, all data on it will be deleted.\n\nDon't use this in production!",
  "email": "admin@example.org",
  "version": "0.0.0-testrig",
  "languages": [
    "nl",
    "en-gb"
  ],
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25,
      "supported_mime_types": [
        "text/plain",
        "text/markdown"
      ]
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/webp",
        "audio/mp2",
        "audio/mp3",
        "video/x-msvideo",
        "audio/flac",
        "audio/x-flac",
        "image/png",
        "image/apng",
        "audio/ogg",
        "video/ogg",
        "audio/x-m4a",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 16777216,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 60,
      "video_matrix_limit": 16777216
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "urls": {
    "streaming_api": "wss://localhost:8080"
  },
  "stats": {
    "domain_count": 2,
    "status_count": 20,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/attachment/original/`+instanceAccount.AvatarMediaAttachment.ID+`.gif",`+`
  "thumbnail_type": "image/gif",
  "thumbnail_static": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/attachment/small/`+instanceAccount.AvatarMediaAttachment.ID+`.webp",`+`
  "thumbnail_static_type": "image/webp",
  "thumbnail_description": "A bouncing little green peglin.",
  "contact_account": {
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
  "max_toot_chars": 5000,
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
  "terms": "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, dst.String())

	// extra bonus: check the v2 model thumbnail after the patch
	instanceV2, err := suite.processor.InstanceGetV2(context.Background())
	if err != nil {
		suite.FailNow(err.Error())
	}

	instanceV2ThumbnailJson, err := json.MarshalIndent(instanceV2.Thumbnail, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/attachment/original/`+instanceAccount.AvatarMediaAttachment.ID+`.gif",`+`
  "thumbnail_type": "image/gif",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/attachment/small/`+instanceAccount.AvatarMediaAttachment.ID+`.webp",`+`
  "thumbnail_static_type": "image/webp",
  "thumbnail_description": "A bouncing little green peglin.",
  "blurhash": "LE9801Rl4Yt5%dWCV]t5Dmoex?WC"
}`, string(instanceV2ThumbnailJson))

	// double extra special bonus: now update the image description without changing the image
	code2, b2 := suite.instancePatch("", "", map[string][]string{
		"thumbnail_description": {"updating the thumbnail description without changing anything else!"},
	})

	if expectedCode := http.StatusOK; code2 != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code2)
	}

	// just extract the value we wanna check, no need to print the whole thing again
	i := make(map[string]interface{})
	if err := json.Unmarshal(b2, &i); err != nil {
		suite.FailNow(err.Error())
	}

	suite.EqualValues("updating the thumbnail description without changing anything else!", i["thumbnail_description"])
}

func (suite *InstancePatchTestSuite) TestInstancePatch9() {
	code, b := suite.instancePatch("", "", map[string][]string{
		"thumbnail_description": {"setting a new description without having a custom image set; this should change nothing!"},
	})

	if expectedCode := http.StatusOK; code != expectedCode {
		suite.FailNowf("wrong status code", "expected %d but got %d", expectedCode, code)
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "description": "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "<p>This is the GoToSocial testrig. It doesn't federate or anything.</p><p>When the testrig is shut down, all data on it will be deleted.</p><p>Don't use this in production!</p>",
  "short_description_text": "This is the GoToSocial testrig. It doesn't federate or anything.\n\nWhen the testrig is shut down, all data on it will be deleted.\n\nDon't use this in production!",
  "email": "admin@example.org",
  "version": "0.0.0-testrig",
  "languages": [
    "nl",
    "en-gb"
  ],
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25,
      "supported_mime_types": [
        "text/plain",
        "text/markdown"
      ]
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/webp",
        "audio/mp2",
        "audio/mp3",
        "video/x-msvideo",
        "audio/flac",
        "audio/x-flac",
        "image/png",
        "image/apng",
        "audio/ogg",
        "video/ogg",
        "audio/x-m4a",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 16777216,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 60,
      "video_matrix_limit": 16777216
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "urls": {
    "streaming_api": "wss://localhost:8080"
  },
  "stats": {
    "domain_count": 2,
    "status_count": 20,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/assets/logo.webp",
  "contact_account": {
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
  "max_toot_chars": 5000,
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
  "terms": "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, dst.String())
}

func TestInstancePatchTestSuite(t *testing.T) {
	suite.Run(t, &InstancePatchTestSuite{})
}
