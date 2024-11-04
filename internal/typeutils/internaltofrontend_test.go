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

package typeutils_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InternalToFrontendTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontend() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
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
  "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
  "header_description": "A very old-school screenshot of the original team fortress mod for quake",
  "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 8,
  "last_status_at": "2024-01-10",
  "emojis": [],
  "fields": [],
  "enable_rss": true
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendAliasedAndMoved() {
	// Take zork for this test.
	var testAccount = new(gtsmodel.Account)
	*testAccount = *suite.testAccounts["local_account_1"]

	// Update zork to indicate that he's moved to turtle.
	// This is a bit weird but it's just for this test.
	movedTo := suite.testAccounts["local_account_2"]
	testAccount.MovedToURI = movedTo.URI
	testAccount.AlsoKnownAsURIs = []string{movedTo.URI}

	if err := suite.state.DB.UpdateAccount(context.Background(), testAccount, "moved_to_uri"); err != nil {
		suite.FailNow(err.Error())
	}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountSensitive(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	// moved and also_known_as_uris
	// should both be set now.
	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
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
  "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
  "header_description": "A very old-school screenshot of the original team fortress mod for quake",
  "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 8,
  "last_status_at": "2024-01-10",
  "emojis": [],
  "fields": [],
  "source": {
    "privacy": "public",
    "web_visibility": "unlisted",
    "sensitive": false,
    "language": "en",
    "status_content_type": "text/plain",
    "note": "hey yo this is my profile!",
    "fields": [],
    "follow_requests_count": 0,
    "also_known_as_uris": [
      "http://localhost:8080/users/1happyturtle"
    ]
  },
  "enable_rss": true,
  "role": {
    "id": "user",
    "name": "user",
    "color": "",
    "permissions": "0",
    "highlighted": false
  },
  "moved": {
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
    "statuses_count": 8,
    "last_status_at": "2021-07-28",
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
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendWithEmojiStruct() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test
	testEmoji := suite.testEmojis["rainbow"]

	testAccount.Emojis = []*gtsmodel.Emoji{testEmoji}
	testAccount.EmojiIDs = []string{testEmoji.ID}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
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
  "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
  "header_description": "A very old-school screenshot of the original team fortress mod for quake",
  "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 8,
  "last_status_at": "2024-01-10",
  "emojis": [
    {
      "shortcode": "rainbow",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "visible_in_picker": true,
      "category": "reactions"
    }
  ],
  "fields": [],
  "enable_rss": true
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendWithEmojiIDs() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	testEmoji := suite.testEmojis["rainbow"]

	testAccount.EmojiIDs = []string{testEmoji.ID}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
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
  "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
  "header_description": "A very old-school screenshot of the original team fortress mod for quake",
  "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 8,
  "last_status_at": "2024-01-10",
  "emojis": [
    {
      "shortcode": "rainbow",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "visible_in_picker": true,
      "category": "reactions"
    }
  ],
  "fields": [],
  "enable_rss": true
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendSensitive() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	apiAccount, err := suite.typeconverter.AccountToAPIAccountSensitive(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
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
  "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
  "header_description": "A very old-school screenshot of the original team fortress mod for quake",
  "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 8,
  "last_status_at": "2024-01-10",
  "emojis": [],
  "fields": [],
  "source": {
    "privacy": "public",
    "web_visibility": "unlisted",
    "sensitive": false,
    "language": "en",
    "status_content_type": "text/plain",
    "note": "hey yo this is my profile!",
    "fields": [],
    "follow_requests_count": 0
  },
  "enable_rss": true,
  "role": {
    "id": "user",
    "name": "user",
    "color": "",
    "permissions": "0",
    "highlighted": false
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendPublicPunycode() {
	testAccount := suite.testAccounts["remote_account_4"]
	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)

	// Even though account domain is stored in
	// punycode, it should be served in its
	// unicode representation in the 'acct' field.
	suite.Equal(`{
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
  "header_description": "Flat gray background (default header).",
  "followers_count": 0,
  "following_count": 0,
  "statuses_count": 0,
  "last_status_at": null,
  "emojis": [],
  "fields": []
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestLocalInstanceAccountToFrontendPublic() {
	ctx := context.Background()
	testAccount, err := suite.db.GetInstanceAccount(ctx, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(ctx, testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
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
  "header_description": "Flat gray background (default header).",
  "followers_count": 0,
  "following_count": 0,
  "statuses_count": 0,
  "last_status_at": null,
  "emojis": [],
  "fields": []
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestLocalInstanceAccountToFrontendBlocked() {
	ctx := context.Background()
	testAccount, err := suite.db.GetInstanceAccount(ctx, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountBlocked(ctx, testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.MarshalIndent(apiAccount, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01AY6P665V14JJR0AFVRT7311Y",
  "username": "localhost:8080",
  "acct": "localhost:8080",
  "display_name": "",
  "locked": false,
  "discoverable": false,
  "bot": false,
  "created_at": "2020-05-17T13:10:59.000Z",
  "note": "",
  "url": "http://localhost:8080/@localhost:8080",
  "avatar": "",
  "avatar_static": "",
  "header": "http://localhost:8080/assets/default_header.webp",
  "header_static": "http://localhost:8080/assets/default_header.webp",
  "header_description": "Flat gray background (default header).",
  "followers_count": 0,
  "following_count": 0,
  "statuses_count": 0,
  "last_status_at": null,
  "emojis": [],
  "fields": []
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontend() {
	testStatus := suite.testStatuses["admin_account_status_1"]
	requestingAccount := suite.testAccounts["local_account_1"]
	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount, statusfilter.FilterContextNone, nil, nil)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01F8MH75CBF9JFX4ZAD54N0W0R",
  "created_at": "2021-10-20T11:36:45.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": false,
  "spoiler_text": "",
  "visibility": "public",
  "language": "en",
  "uri": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "replies_count": 1,
  "reblogs_count": 0,
  "favourites_count": 1,
  "favourited": true,
  "reblogged": false,
  "muted": false,
  "bookmarked": true,
  "pinned": false,
  "content": "hello world! #welcome ! first post on the instance :rainbow: !",
  "reblog": null,
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
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
    ]
  },
  "media_attachments": [
    {
      "id": "01F8MH6NEM8D7527KZAECTCR76",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.webp",
      "remote_url": null,
      "preview_remote_url": null,
      "meta": {
        "original": {
          "width": 1200,
          "height": 630,
          "size": "1200x630",
          "aspect": 1.9047619
        },
        "small": {
          "width": 512,
          "height": 268,
          "size": "512x268",
          "aspect": 1.9104477
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Black and white image of some 50's style text saying: Welcome On Board",
      "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj"
    }
  ],
  "mentions": [],
  "tags": [
    {
      "name": "welcome",
      "url": "http://localhost:8080/tags/welcome"
    }
  ],
  "emojis": [
    {
      "shortcode": "rainbow",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "visible_in_picker": true,
      "category": "reactions"
    }
  ],
  "card": null,
  "poll": null,
  "text": "hello world! #welcome ! first post on the instance :rainbow: !",
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
}`, string(b))
}

// Modify a fixture status into a status that should be filtered,
// and then filter it, returning the API status or any error from converting it.
func (suite *InternalToFrontendTestSuite) filteredStatusToFrontend(action gtsmodel.FilterAction, boost bool) (*apimodel.Status, error) {
	testStatus := suite.testStatuses["admin_account_status_1"]
	testStatus.Content += " fnord"
	testStatus.Text += " fnord"

	if boost {
		// Modify a fixture boost into a boost of the above status.
		boostStatus := suite.testStatuses["admin_account_status_4"]
		boostStatus.BoostOf = testStatus
		boostStatus.BoostOfID = testStatus.ID
		testStatus = boostStatus
	}

	requestingAccount := suite.testAccounts["local_account_1"]

	expectedMatchingFilter := suite.testFilters["local_account_1_filter_1"]
	expectedMatchingFilter.Action = action

	expectedMatchingFilterKeyword := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"]
	suite.NoError(expectedMatchingFilterKeyword.Compile())
	expectedMatchingFilterKeyword.Filter = expectedMatchingFilter

	expectedMatchingFilter.Keywords = []*gtsmodel.FilterKeyword{expectedMatchingFilterKeyword}

	requestingAccountFilters := []*gtsmodel.Filter{expectedMatchingFilter}

	return suite.typeconverter.StatusToAPIStatus(
		context.Background(),
		testStatus,
		requestingAccount,
		statusfilter.FilterContextHome,
		requestingAccountFilters,
		nil,
	)
}

// Test that a status which is filtered with a warn filter by the requesting user has `filtered` set correctly.
func (suite *InternalToFrontendTestSuite) TestWarnFilteredStatusToFrontend() {
	apiStatus, err := suite.filteredStatusToFrontend(gtsmodel.FilterActionWarn, false)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01F8MH75CBF9JFX4ZAD54N0W0R",
  "created_at": "2021-10-20T11:36:45.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": false,
  "spoiler_text": "",
  "visibility": "public",
  "language": "en",
  "uri": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "replies_count": 1,
  "reblogs_count": 0,
  "favourites_count": 1,
  "favourited": true,
  "reblogged": false,
  "muted": false,
  "bookmarked": true,
  "pinned": false,
  "content": "hello world! #welcome ! first post on the instance :rainbow: ! fnord",
  "reblog": null,
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
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
    ]
  },
  "media_attachments": [
    {
      "id": "01F8MH6NEM8D7527KZAECTCR76",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.webp",
      "remote_url": null,
      "preview_remote_url": null,
      "meta": {
        "original": {
          "width": 1200,
          "height": 630,
          "size": "1200x630",
          "aspect": 1.9047619
        },
        "small": {
          "width": 512,
          "height": 268,
          "size": "512x268",
          "aspect": 1.9104477
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Black and white image of some 50's style text saying: Welcome On Board",
      "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj"
    }
  ],
  "mentions": [],
  "tags": [
    {
      "name": "welcome",
      "url": "http://localhost:8080/tags/welcome"
    }
  ],
  "emojis": [
    {
      "shortcode": "rainbow",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "visible_in_picker": true,
      "category": "reactions"
    }
  ],
  "card": null,
  "poll": null,
  "text": "hello world! #welcome ! first post on the instance :rainbow: ! fnord",
  "filtered": [
    {
      "filter": {
        "id": "01HN26VM6KZTW1ANNRVSBMA461",
        "title": "fnord",
        "context": [
          "home",
          "public"
        ],
        "expires_at": null,
        "filter_action": "warn",
        "keywords": [
          {
            "id": "01HN272TAVWAXX72ZX4M8JZ0PS",
            "keyword": "fnord",
            "whole_word": true
          }
        ],
        "statuses": []
      },
      "keyword_matches": [
        "fnord"
      ],
      "status_matches": []
    }
  ],
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
}`, string(b))
}

// Test that a status which is filtered with a warn filter by the requesting user has `filtered` set correctly when boosted.
func (suite *InternalToFrontendTestSuite) TestWarnFilteredBoostToFrontend() {
	apiStatus, err := suite.filteredStatusToFrontend(gtsmodel.FilterActionWarn, true)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01G36SF3V6Y6V5BF9P4R7PQG7G",
  "created_at": "2021-10-20T10:41:37.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": false,
  "spoiler_text": "",
  "visibility": "public",
  "language": null,
  "uri": "http://localhost:8080/users/admin/statuses/01G36SF3V6Y6V5BF9P4R7PQG7G",
  "url": "http://localhost:8080/@admin/statuses/01G36SF3V6Y6V5BF9P4R7PQG7G",
  "replies_count": 0,
  "reblogs_count": 0,
  "favourites_count": 0,
  "favourited": true,
  "reblogged": false,
  "muted": false,
  "bookmarked": true,
  "pinned": false,
  "content": "",
  "reblog": {
    "id": "01F8MH75CBF9JFX4ZAD54N0W0R",
    "created_at": "2021-10-20T11:36:45.000Z",
    "in_reply_to_id": null,
    "in_reply_to_account_id": null,
    "sensitive": false,
    "spoiler_text": "",
    "visibility": "public",
    "language": "en",
    "uri": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
    "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
    "replies_count": 1,
    "reblogs_count": 0,
    "favourites_count": 1,
    "favourited": true,
    "reblogged": false,
    "muted": false,
    "bookmarked": true,
    "pinned": false,
    "content": "hello world! #welcome ! first post on the instance :rainbow: ! fnord",
    "reblog": null,
    "application": {
      "name": "superseriousbusiness",
      "website": "https://superserious.business"
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
      "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
      "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
      "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
      "header_description": "A very old-school screenshot of the original team fortress mod for quake",
      "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
      "followers_count": 2,
      "following_count": 2,
      "statuses_count": 8,
      "last_status_at": "2024-01-10",
      "emojis": [],
      "fields": [],
      "enable_rss": true
    },
    "media_attachments": [
      {
        "id": "01F8MH6NEM8D7527KZAECTCR76",
        "type": "image",
        "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
        "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
        "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.webp",
        "remote_url": null,
        "preview_remote_url": null,
        "meta": {
          "original": {
            "width": 1200,
            "height": 630,
            "size": "1200x630",
            "aspect": 1.9047619
          },
          "small": {
            "width": 512,
            "height": 268,
            "size": "512x268",
            "aspect": 1.9104477
          },
          "focus": {
            "x": 0,
            "y": 0
          }
        },
        "description": "Black and white image of some 50's style text saying: Welcome On Board",
        "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj"
      }
    ],
    "mentions": [],
    "tags": [
      {
        "name": "welcome",
        "url": "http://localhost:8080/tags/welcome"
      }
    ],
    "emojis": [
      {
        "shortcode": "rainbow",
        "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
        "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
        "visible_in_picker": true,
        "category": "reactions"
      }
    ],
    "card": null,
    "poll": null,
    "text": "hello world! #welcome ! first post on the instance :rainbow: ! fnord",
    "filtered": [
      {
        "filter": {
          "id": "01HN26VM6KZTW1ANNRVSBMA461",
          "title": "fnord",
          "context": [
            "home",
            "public"
          ],
          "expires_at": null,
          "filter_action": "warn",
          "keywords": [
            {
              "id": "01HN272TAVWAXX72ZX4M8JZ0PS",
              "keyword": "fnord",
              "whole_word": true
            }
          ],
          "statuses": []
        },
        "keyword_matches": [
          "fnord"
        ],
        "status_matches": []
      }
    ],
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
  },
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
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
    ]
  },
  "media_attachments": [],
  "mentions": [],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
  "filtered": [
    {
      "filter": {
        "id": "01HN26VM6KZTW1ANNRVSBMA461",
        "title": "fnord",
        "context": [
          "home",
          "public"
        ],
        "expires_at": null,
        "filter_action": "warn",
        "keywords": [
          {
            "id": "01HN272TAVWAXX72ZX4M8JZ0PS",
            "keyword": "fnord",
            "whole_word": true
          }
        ],
        "statuses": []
      },
      "keyword_matches": [
        "fnord"
      ],
      "status_matches": []
    }
  ],
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
}`, string(b))
}

// Test that a status which is filtered with a hide filter by the requesting user results in the ErrHideStatus error.
func (suite *InternalToFrontendTestSuite) TestHideFilteredStatusToFrontend() {
	_, err := suite.filteredStatusToFrontend(gtsmodel.FilterActionHide, false)
	suite.ErrorIs(err, statusfilter.ErrHideStatus)
}

// Test that a status which is filtered with a hide filter by the requesting user results in the ErrHideStatus error for a boost of that status.
func (suite *InternalToFrontendTestSuite) TestHideFilteredBoostToFrontend() {
	_, err := suite.filteredStatusToFrontend(gtsmodel.FilterActionHide, true)
	suite.ErrorIs(err, statusfilter.ErrHideStatus)
}

// Test that a hashtag filter for a hashtag in Mastodon HTML content works the way most users would expect.
func (suite *InternalToFrontendTestSuite) testHashtagFilteredStatusToFrontend(wholeWord bool, boost bool) {
	testStatus := new(gtsmodel.Status)
	*testStatus = *suite.testStatuses["admin_account_status_1"]
	testStatus.Content = `<p>doggo doggin' it</p><p><a href="https://example.test/tags/dogsofmastodon" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>dogsofmastodon</span></a></p>`

	if boost {
		boost, err := suite.typeconverter.StatusToBoost(
			context.Background(),
			testStatus,
			suite.testAccounts["admin_account"],
			"",
		)
		if err != nil {
			suite.FailNow(err.Error())
		}
		testStatus = boost
	}

	requestingAccount := suite.testAccounts["local_account_1"]

	filterKeyword := &gtsmodel.FilterKeyword{
		Keyword:   "#dogsofmastodon",
		WholeWord: &wholeWord,
		Regexp:    nil,
	}
	if err := filterKeyword.Compile(); err != nil {
		suite.FailNow(err.Error())
	}

	filter := &gtsmodel.Filter{
		Action:               gtsmodel.FilterActionWarn,
		Keywords:             []*gtsmodel.FilterKeyword{filterKeyword},
		ContextHome:          util.Ptr(true),
		ContextNotifications: util.Ptr(false),
		ContextPublic:        util.Ptr(false),
		ContextThread:        util.Ptr(false),
		ContextAccount:       util.Ptr(false),
	}

	apiStatus, err := suite.typeconverter.StatusToAPIStatus(
		context.Background(),
		testStatus,
		requestingAccount,
		statusfilter.FilterContextHome,
		[]*gtsmodel.Filter{filter},
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(apiStatus.Filtered)
}

func (suite *InternalToFrontendTestSuite) TestHashtagWholeWordFilteredStatusToFrontend() {
	suite.testHashtagFilteredStatusToFrontend(true, false)
}

func (suite *InternalToFrontendTestSuite) TestHashtagWholeWordFilteredBoostToFrontend() {
	suite.testHashtagFilteredStatusToFrontend(true, true)
}

func (suite *InternalToFrontendTestSuite) TestHashtagAnywhereFilteredStatusToFrontend() {
	suite.testHashtagFilteredStatusToFrontend(false, false)
}

func (suite *InternalToFrontendTestSuite) TestHashtagAnywhereFilteredBoostToFrontend() {
	suite.testHashtagFilteredStatusToFrontend(false, true)
}

// Test that a status from a user muted by the requesting user results in the ErrHideStatus error.
func (suite *InternalToFrontendTestSuite) TestMutedStatusToFrontend() {
	testStatus := suite.testStatuses["admin_account_status_1"]
	requestingAccount := suite.testAccounts["local_account_1"]
	mutes := usermute.NewCompiledUserMuteList([]*gtsmodel.UserMute{
		{
			AccountID:       requestingAccount.ID,
			TargetAccountID: testStatus.AccountID,
			Notifications:   util.Ptr(false),
		},
	})
	_, err := suite.typeconverter.StatusToAPIStatus(
		context.Background(),
		testStatus,
		requestingAccount,
		statusfilter.FilterContextHome,
		nil,
		mutes,
	)
	suite.ErrorIs(err, statusfilter.ErrHideStatus)
}

// Test that a status replying to a user muted by the requesting user results in the ErrHideStatus error.
func (suite *InternalToFrontendTestSuite) TestMutedReplyStatusToFrontend() {
	mutedAccount := suite.testAccounts["local_account_2"]
	testStatus := suite.testStatuses["admin_account_status_1"]
	testStatus.InReplyToID = suite.testStatuses["local_account_2_status_1"].ID
	testStatus.InReplyToAccountID = mutedAccount.ID
	requestingAccount := suite.testAccounts["local_account_1"]
	mutes := usermute.NewCompiledUserMuteList([]*gtsmodel.UserMute{
		{
			AccountID:       requestingAccount.ID,
			TargetAccountID: mutedAccount.ID,
			Notifications:   util.Ptr(false),
		},
	})
	// Populate status so the converter has the account objects it needs for muting.
	err := suite.db.PopulateStatus(context.Background(), testStatus)
	if err != nil {
		suite.FailNow(err.Error())
	}
	// Convert the status to API format, which should fail.
	_, err = suite.typeconverter.StatusToAPIStatus(
		context.Background(),
		testStatus,
		requestingAccount,
		statusfilter.FilterContextHome,
		nil,
		mutes,
	)
	suite.ErrorIs(err, statusfilter.ErrHideStatus)
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontendUnknownAttachments() {
	testStatus := suite.testStatuses["remote_account_2_status_1"]
	requestingAccount := suite.testAccounts["admin_account"]

	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount, statusfilter.FilterContextNone, nil, nil)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01HE7XJ1CG84TBKH5V9XKBVGF5",
  "created_at": "2023-11-02T10:44:25.000Z",
  "in_reply_to_id": "01F8MH75CBF9JFX4ZAD54N0W0R",
  "in_reply_to_account_id": "01F8MH17FWEB39HZJ76B6VXSKF",
  "sensitive": true,
  "spoiler_text": "some unknown media included",
  "visibility": "public",
  "language": "en",
  "uri": "http://example.org/users/Some_User/statuses/01HE7XJ1CG84TBKH5V9XKBVGF5",
  "url": "http://example.org/@Some_User/statuses/01HE7XJ1CG84TBKH5V9XKBVGF5",
  "replies_count": 0,
  "reblogs_count": 0,
  "favourites_count": 0,
  "favourited": false,
  "reblogged": false,
  "muted": false,
  "bookmarked": false,
  "pinned": false,
  "content": "\u003cp\u003ehi \u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003eadmin\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e here's some media for ya\u003c/p\u003e\u003chr\u003e\u003cp\u003e\u003ci lang=\"en\"\u003eℹ️ Note from localhost:8080: 2 attachments in this status were not downloaded. Treat the following external links with care:\u003c/i\u003e\u003c/p\u003e\u003cul\u003e\u003cli\u003e\u003ca href=\"http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7ZGJYTSYMXF927GF9353KR.svg\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e01HE7ZGJYTSYMXF927GF9353KR.svg\u003c/a\u003e [SVG line art of a sloth, public domain]\u003c/li\u003e\u003cli\u003e\u003ca href=\"http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE892Y8ZS68TQCNPX7J888P3.mp3\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e01HE892Y8ZS68TQCNPX7J888P3.mp3\u003c/a\u003e [Jolly salsa song, public domain.]\u003c/li\u003e\u003c/ul\u003e",
  "reblog": null,
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
    "header_description": "Flat gray background (default header).",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 1,
    "last_status_at": "2023-11-02",
    "emojis": [],
    "fields": []
  },
  "media_attachments": [
    {
      "id": "01HE7Y3C432WRSNS10EZM86SA5",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7Y3C432WRSNS10EZM86SA5.jpg",
      "text_url": "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7Y3C432WRSNS10EZM86SA5.jpg",
      "preview_url": "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/small/01HE7Y3C432WRSNS10EZM86SA5.webp",
      "remote_url": "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7Y6G0EMCKST3Q0914WW0MS.jpg",
      "preview_remote_url": null,
      "meta": {
        "original": {
          "width": 3000,
          "height": 2000,
          "size": "3000x2000",
          "aspect": 1.5
        },
        "small": {
          "width": 512,
          "height": 341,
          "size": "512x341",
          "aspect": 1.5014663
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Photograph of a sloth, Public Domain.",
      "blurhash": "LKE3VIw}0KD%a2o{M|t7NFWps:t7"
    }
  ],
  "mentions": [
    {
      "id": "01F8MH17FWEB39HZJ76B6VXSKF",
      "username": "admin",
      "url": "http://localhost:8080/@admin",
      "acct": "admin"
    }
  ],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
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
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToWebStatus() {
	testStatus := suite.testStatuses["remote_account_2_status_1"]

	apiStatus, err := suite.typeconverter.StatusToWebStatus(context.Background(), testStatus)
	suite.NoError(err)

	// MediaAttachments should inherit
	// the status's sensitive flag.
	for _, a := range apiStatus.MediaAttachments {
		if !a.Sensitive {
			suite.FailNow("expected sensitive attachment")
		}
	}

	// We don't really serialize web statuses to JSON
	// ever, but it *is* a nice way of checking it.
	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01HE7XJ1CG84TBKH5V9XKBVGF5",
  "created_at": "2023-11-02T10:44:25.000Z",
  "in_reply_to_id": "01F8MH75CBF9JFX4ZAD54N0W0R",
  "in_reply_to_account_id": "01F8MH17FWEB39HZJ76B6VXSKF",
  "sensitive": true,
  "spoiler_text": "some unknown media included",
  "visibility": "public",
  "language": "en",
  "uri": "http://example.org/users/Some_User/statuses/01HE7XJ1CG84TBKH5V9XKBVGF5",
  "url": "http://example.org/@Some_User/statuses/01HE7XJ1CG84TBKH5V9XKBVGF5",
  "replies_count": 0,
  "reblogs_count": 0,
  "favourites_count": 0,
  "favourited": false,
  "reblogged": false,
  "muted": false,
  "bookmarked": false,
  "pinned": false,
  "content": "\u003cp\u003ehi \u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003eadmin\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e here's some media for ya\u003c/p\u003e",
  "reblog": null,
  "mentions": [
    {
      "id": "01F8MH17FWEB39HZJ76B6VXSKF",
      "username": "admin",
      "url": "http://localhost:8080/@admin",
      "acct": "admin"
    }
  ],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public"
      ],
      "with_approval": []
    }
  },
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
    "header_description": "Flat gray background (default header).",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 1,
    "last_status_at": "2023-11-02",
    "emojis": [],
    "fields": []
  },
  "media_attachments": [
    {
      "id": "01HE7Y3C432WRSNS10EZM86SA5",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7Y3C432WRSNS10EZM86SA5.jpg",
      "text_url": "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7Y3C432WRSNS10EZM86SA5.jpg",
      "preview_url": "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/small/01HE7Y3C432WRSNS10EZM86SA5.webp",
      "remote_url": "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7Y6G0EMCKST3Q0914WW0MS.jpg",
      "preview_remote_url": null,
      "meta": {
        "original": {
          "width": 3000,
          "height": 2000,
          "size": "3000x2000",
          "aspect": 1.5
        },
        "small": {
          "width": 512,
          "height": 341,
          "size": "512x341",
          "aspect": 1.5014663
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Photograph of a sloth, Public Domain.",
      "blurhash": "LKE3VIw}0KD%a2o{M|t7NFWps:t7",
      "Sensitive": true,
      "MIMEType": "image/jpg",
      "PreviewMIMEType": "image/webp"
    },
    {
      "id": "01HE7ZFX9GKA5ZZVD4FACABSS9",
      "type": "unknown",
      "url": null,
      "text_url": null,
      "preview_url": null,
      "remote_url": "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7ZGJYTSYMXF927GF9353KR.svg",
      "preview_remote_url": null,
      "meta": null,
      "description": "SVG line art of a sloth, public domain",
      "blurhash": "L26*j+~qE1RP?wxut7ofRlM{R*of",
      "Sensitive": true,
      "MIMEType": "",
      "PreviewMIMEType": ""
    },
    {
      "id": "01HE88YG74PVAB81PX2XA9F3FG",
      "type": "unknown",
      "url": null,
      "text_url": null,
      "preview_url": null,
      "remote_url": "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE892Y8ZS68TQCNPX7J888P3.mp3",
      "preview_remote_url": null,
      "meta": null,
      "description": "Jolly salsa song, public domain.",
      "blurhash": null,
      "Sensitive": true,
      "MIMEType": "",
      "PreviewMIMEType": ""
    }
  ],
  "LanguageTag": "en",
  "PollOptions": null,
  "Local": false,
  "Indent": 0,
  "ThreadLastMain": false,
  "ThreadContextStatus": false,
  "ThreadFirstReply": false
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontendUnknownLanguage() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["admin_account_status_1"]
	testStatus.Language = ""
	requestingAccount := suite.testAccounts["local_account_1"]
	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount, statusfilter.FilterContextNone, nil, nil)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01F8MH75CBF9JFX4ZAD54N0W0R",
  "created_at": "2021-10-20T11:36:45.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": false,
  "spoiler_text": "",
  "visibility": "public",
  "language": null,
  "uri": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "replies_count": 1,
  "reblogs_count": 0,
  "favourites_count": 1,
  "favourited": true,
  "reblogged": false,
  "muted": false,
  "bookmarked": true,
  "pinned": false,
  "content": "hello world! #welcome ! first post on the instance :rainbow: !",
  "reblog": null,
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
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
    ]
  },
  "media_attachments": [
    {
      "id": "01F8MH6NEM8D7527KZAECTCR76",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.webp",
      "remote_url": null,
      "preview_remote_url": null,
      "meta": {
        "original": {
          "width": 1200,
          "height": 630,
          "size": "1200x630",
          "aspect": 1.9047619
        },
        "small": {
          "width": 512,
          "height": 268,
          "size": "512x268",
          "aspect": 1.9104477
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Black and white image of some 50's style text saying: Welcome On Board",
      "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj"
    }
  ],
  "mentions": [],
  "tags": [
    {
      "name": "welcome",
      "url": "http://localhost:8080/tags/welcome"
    }
  ],
  "emojis": [
    {
      "shortcode": "rainbow",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
      "visible_in_picker": true,
      "category": "reactions"
    }
  ],
  "card": null,
  "poll": null,
  "text": "hello world! #welcome ! first post on the instance :rainbow: !",
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
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontendPartialInteractions() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["local_account_1_status_3"]
	testStatus.Language = ""
	requestingAccount := suite.testAccounts["admin_account"]
	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount, statusfilter.FilterContextNone, nil, nil)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiStatus, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01F8MHBBN8120SYH7D5S050MGK",
  "created_at": "2021-10-20T10:40:37.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": false,
  "spoiler_text": "test: you shouldn't be able to interact with this post in any way",
  "visibility": "private",
  "language": null,
  "uri": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
  "replies_count": 0,
  "reblogs_count": 0,
  "favourites_count": 0,
  "favourited": false,
  "reblogged": false,
  "muted": false,
  "bookmarked": false,
  "pinned": false,
  "content": "this is a very personal post that I don't want anyone to interact with at all, and i only want mutuals to see it",
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
    "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
    "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
    "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
    "header_description": "A very old-school screenshot of the original team fortress mod for quake",
    "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
    "followers_count": 2,
    "following_count": 2,
    "statuses_count": 8,
    "last_status_at": "2024-01-10",
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
  "text": "this is a very personal post that I don't want anyone to interact with at all, and i only want mutuals to see it",
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "author"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "author"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "author"
      ],
      "with_approval": []
    }
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToAPIStatusPendingApproval() {
	var (
		testStatus        = suite.testStatuses["admin_account_status_5"]
		requestingAccount = suite.testAccounts["local_account_2"]
	)

	apiStatus, err := suite.typeconverter.StatusToAPIStatus(
		context.Background(),
		testStatus,
		requestingAccount,
		statusfilter.FilterContextNone,
		nil,
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// We want to see the HTML in
	// the status so don't escape it.
	out := new(bytes.Buffer)
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(apiStatus); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "id": "01J5QVB9VC76NPPRQ207GG4DRZ",
  "created_at": "2024-02-20T10:41:37.000Z",
  "in_reply_to_id": "01F8MHC8VWDRBQR0N1BATDDEM5",
  "in_reply_to_account_id": "01F8MH5NBDF2MV7CTC4Q5128HF",
  "sensitive": false,
  "spoiler_text": "",
  "visibility": "public",
  "language": null,
  "uri": "http://localhost:8080/users/admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
  "url": "http://localhost:8080/@admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
  "replies_count": 0,
  "reblogs_count": 0,
  "favourites_count": 0,
  "favourited": false,
  "reblogged": false,
  "muted": false,
  "bookmarked": false,
  "pinned": false,
  "content": "<p>Hi <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span>, can I reply?</p><hr><p><i lang=\"en\">ℹ️ Note from localhost:8080: This reply is pending your approval. You can quickly accept it by liking, boosting or replying to it. You can also accept or reject it at the following link: <a href=\"http://localhost:8080/settings/user/interaction_requests/01J5QVXCCEATJYSXM9H6MZT4JR\" rel=\"noreferrer noopener nofollow\" target=\"_blank\">http://localhost:8080/settings/user/interaction_requests/01J5QVXCCEATJYSXM9H6MZT4JR</a>.</i></p>",
  "reblog": null,
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
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
    ]
  },
  "media_attachments": [],
  "mentions": [
    {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "url": "http://localhost:8080/@1happyturtle",
      "acct": "1happyturtle"
    }
  ],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
  "text": "Hi @1happyturtle, can I reply?",
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
}
`, out.String())
}

func (suite *InternalToFrontendTestSuite) TestVideoAttachmentToFrontend() {
	testAttachment := suite.testAttachments["local_account_1_status_4_attachment_2"]
	apiAttachment, err := suite.typeconverter.AttachmentToAPIAttachment(context.Background(), testAttachment)
	suite.NoError(err)

	b, err := json.MarshalIndent(apiAttachment, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01CDR64G398ADCHXK08WWTHEZ5",
  "type": "video",
  "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01CDR64G398ADCHXK08WWTHEZ5.mp4",
  "text_url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01CDR64G398ADCHXK08WWTHEZ5.mp4",
  "preview_url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01CDR64G398ADCHXK08WWTHEZ5.webp",
  "remote_url": null,
  "preview_remote_url": null,
  "meta": {
    "original": {
      "width": 720,
      "height": 404,
      "frame_rate": "30/1",
      "duration": 15.034,
      "bitrate": 1209808,
      "size": "720x404",
      "aspect": 1.7821782
    },
    "small": {
      "width": 512,
      "height": 287,
      "size": "512x287",
      "aspect": 1.7839721
    },
    "focus": {
      "x": 0,
      "y": 0
    }
  },
  "description": "A cow adorably licking another cow!",
  "blurhash": "L9B|BBY8yZtS~AxZV@t6,njEjZV@"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestInstanceV1ToFrontend() {
	ctx := context.Background()

	i := &gtsmodel.Instance{}
	if err := suite.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: config.GetHost()}}, i); err != nil {
		suite.FailNow(err.Error())
	}

	instance, err := suite.typeconverter.InstanceToAPIV1Instance(ctx, i)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(instance, "", "  ")
	suite.NoError(err)

	// FIXME: "rules" is empty from the database, because it's not fetched through db.GetInstance
	suite.Equal(`{
  "uri": "localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "description": "\u003cp\u003eHere's a fuller description of the GoToSocial testrig instance.\u003c/p\u003e\u003cp\u003eThis instance is for testing purposes only. It doesn't federate at all. Go check out \u003ca href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003ehttps://github.com/superseriousbusiness/gotosocial/tree/main/testrig\u003c/a\u003e and \u003ca href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003ehttps://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\u003c/a\u003e\u003c/p\u003e\u003cp\u003eUsers on this instance:\u003c/p\u003e\u003cul\u003e\u003cli\u003e\u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003eadmin\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e (admin!).\u003c/li\u003e\u003cli\u003e\u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003e1happyturtle\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e (posts about turtles, we don't know why).\u003c/li\u003e\u003cli\u003e\u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003ethe_mighty_zork\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e (who knows).\u003c/li\u003e\u003c/ul\u003e\u003cp\u003eIf you need to edit the models for the testrig, you can do so at \u003ccode\u003einternal/testmodels.go\u003c/code\u003e.\u003c/p\u003e",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "short_description": "\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e",
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
        "audio/mp4",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 9223372036854775807,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 9223372036854775807,
      "video_matrix_limit": 9223372036854775807
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
    ]
  },
  "max_toot_chars": 5000,
  "rules": [],
  "terms": "\u003cp\u003eThis is where a list of terms and conditions might go.\u003c/p\u003e\u003cp\u003eFor example:\u003c/p\u003e\u003cp\u003eIf you want to sign up on this instance, you oughta know that we:\u003c/p\u003e\u003col\u003e\u003cli\u003eWill sell your data to whoever offers.\u003c/li\u003e\u003cli\u003eSecure the server with password \u003ccode\u003epassword\u003c/code\u003e wherever possible.\u003c/li\u003e\u003c/ol\u003e",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestInstanceV2ToFrontend() {
	ctx := context.Background()

	i := &gtsmodel.Instance{}
	if err := suite.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: config.GetHost()}}, i); err != nil {
		suite.FailNow(err.Error())
	}

	instance, err := suite.typeconverter.InstanceToAPIV2Instance(ctx, i)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(instance, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "domain": "localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "version": "0.0.0-testrig",
  "source_url": "https://github.com/superseriousbusiness/gotosocial",
  "description": "\u003cp\u003eHere's a fuller description of the GoToSocial testrig instance.\u003c/p\u003e\u003cp\u003eThis instance is for testing purposes only. It doesn't federate at all. Go check out \u003ca href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003ehttps://github.com/superseriousbusiness/gotosocial/tree/main/testrig\u003c/a\u003e and \u003ca href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003ehttps://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\u003c/a\u003e\u003c/p\u003e\u003cp\u003eUsers on this instance:\u003c/p\u003e\u003cul\u003e\u003cli\u003e\u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003eadmin\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e (admin!).\u003c/li\u003e\u003cli\u003e\u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003e1happyturtle\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e (posts about turtles, we don't know why).\u003c/li\u003e\u003cli\u003e\u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003ethe_mighty_zork\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e (who knows).\u003c/li\u003e\u003c/ul\u003e\u003cp\u003eIf you need to edit the models for the testrig, you can do so at \u003ccode\u003einternal/testmodels.go\u003c/code\u003e.\u003c/p\u003e",
  "description_text": "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `+"`"+`internal/testmodels.go`+"`"+`.",
  "usage": {
    "users": {
      "active_month": 0
    }
  },
  "thumbnail": {
    "url": "http://localhost:8080/assets/logo.webp"
  },
  "languages": [
    "nl",
    "en-gb"
  ],
  "configuration": {
    "urls": {
      "streaming": "wss://localhost:8080"
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10,
      "max_profile_fields": 6
    },
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
        "audio/mp4",
        "video/mp4",
        "video/quicktime",
        "audio/x-ms-wma",
        "video/x-ms-wmv",
        "video/webm",
        "audio/x-matroska",
        "video/x-matroska"
      ],
      "image_size_limit": 41943040,
      "image_matrix_limit": 9223372036854775807,
      "video_size_limit": 41943040,
      "video_frame_rate_limit": 9223372036854775807,
      "video_matrix_limit": 9223372036854775807
    },
    "polls": {
      "max_options": 6,
      "max_characters_per_option": 50,
      "min_expiration": 300,
      "max_expiration": 2629746
    },
    "translation": {
      "enabled": false
    },
    "emojis": {
      "emoji_size_limit": 51200
    }
  },
  "registrations": {
    "enabled": true,
    "approval_required": true,
    "message": null
  },
  "contact": {
    "email": "admin@example.org",
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
      ]
    }
  },
  "rules": [],
  "terms": "\u003cp\u003eThis is where a list of terms and conditions might go.\u003c/p\u003e\u003cp\u003eFor example:\u003c/p\u003e\u003cp\u003eIf you want to sign up on this instance, you oughta know that we:\u003c/p\u003e\u003col\u003e\u003cli\u003eWill sell your data to whoever offers.\u003c/li\u003e\u003cli\u003eSecure the server with password \u003ccode\u003epassword\u003c/code\u003e wherever possible.\u003c/li\u003e\u003c/ol\u003e",
  "terms_text": "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `+"`"+`password`+"`"+` wherever possible."
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestEmojiToFrontend() {
	emoji, err := suite.typeconverter.EmojiToAPIEmoji(context.Background(), suite.testEmojis["rainbow"])
	suite.NoError(err)

	b, err := json.MarshalIndent(emoji, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "shortcode": "rainbow",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "visible_in_picker": true,
  "category": "reactions"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestEmojiToFrontendAdmin1() {
	emoji, err := suite.typeconverter.EmojiToAdminAPIEmoji(context.Background(), suite.testEmojis["rainbow"])
	suite.NoError(err)

	b, err := json.MarshalIndent(emoji, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "shortcode": "rainbow",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "visible_in_picker": true,
  "category": "reactions",
  "id": "01F8MH9H8E4VG3KDYJR9EGPXCQ",
  "disabled": false,
  "updated_at": "2021-09-20T10:40:37.000Z",
  "total_file_size": 42794,
  "content_type": "image/png",
  "uri": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestEmojiToFrontendAdmin2() {
	emoji, err := suite.typeconverter.EmojiToAdminAPIEmoji(context.Background(), suite.testEmojis["yell"])
	suite.NoError(err)

	b, err := json.MarshalIndent(emoji, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "shortcode": "yell",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
  "visible_in_picker": false,
  "id": "01GD5KP5CQEE1R3X43Y1EHS2CW",
  "disabled": false,
  "domain": "fossbros-anonymous.io",
  "updated_at": "2020-03-18T12:12:00.000Z",
  "total_file_size": 19854,
  "content_type": "image/png",
  "uri": "http://fossbros-anonymous.io/emoji/01GD5KP5CQEE1R3X43Y1EHS2CW"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestReportToFrontend1() {
	report, err := suite.typeconverter.ReportToAPIReport(context.Background(), suite.testReports["local_account_2_report_remote_account_1"])
	suite.NoError(err)

	b, err := json.MarshalIndent(report, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
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
    "statuses_count": 3,
    "last_status_at": "2021-09-11",
    "emojis": [],
    "fields": []
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestReportToFrontend2() {
	report, err := suite.typeconverter.ReportToAPIReport(context.Background(), suite.testReports["remote_account_1_report_local_account_2"])
	suite.NoError(err)

	b, err := json.MarshalIndent(report, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "id": "01GP3DFY9XQ1TJMZT5BGAZPXX7",
  "created_at": "2022-05-15T14:20:12.000Z",
  "action_taken": true,
  "action_taken_at": "2022-05-15T15:01:56.000Z",
  "action_taken_comment": "user was warned not to be a turtle anymore",
  "category": "other",
  "comment": "this is a turtle, not a person, therefore should not be a poster",
  "forwarded": true,
  "status_ids": [],
  "rule_ids": [],
  "target_account": {
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
    "statuses_count": 8,
    "last_status_at": "2021-07-28",
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
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAdminReportToFrontend1() {
	requestingAccount := suite.testAccounts["admin_account"]
	adminReport, err := suite.typeconverter.ReportToAdminAPIReport(context.Background(), suite.testReports["remote_account_1_report_local_account_2"], requestingAccount)
	suite.NoError(err)

	b, err := json.MarshalIndent(adminReport, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
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
      "statuses_count": 3,
      "last_status_at": "2021-09-11",
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
      "statuses_count": 8,
      "last_status_at": "2021-07-28",
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
      ]
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
      ]
    },
    "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
  },
  "statuses": [],
  "rules": [],
  "action_taken_comment": "user was warned not to be a turtle anymore"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAdminReportToFrontend2() {
	requestingAccount := suite.testAccounts["admin_account"]
	adminReport, err := suite.typeconverter.ReportToAdminAPIReport(context.Background(), suite.testReports["local_account_2_report_remote_account_1"], requestingAccount)
	suite.NoError(err)

	b, err := json.MarshalIndent(adminReport, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
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
      "statuses_count": 8,
      "last_status_at": "2021-07-28",
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
      "statuses_count": 3,
      "last_status_at": "2021-09-11",
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
        "statuses_count": 3,
        "last_status_at": "2021-09-11",
        "emojis": [],
        "fields": []
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
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAdminReportToFrontendSuspendedLocalAccount() {
	ctx := context.Background()
	requestingAccount := suite.testAccounts["admin_account"]
	reportedAccount := &gtsmodel.Account{}
	*reportedAccount = *suite.testAccounts["local_account_2"]

	// Suspend/delete the reported account.
	if err := suite.GetProcessor().Account().Delete(ctx, reportedAccount, requestingAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Wait for the delete to process. Stubbifying
	// the account is the last part of the delete,
	// so once it's stubbified we know we're done.
	if !testrig.WaitFor(func() bool {
		dbAccount, err := suite.db.GetAccountByID(ctx, reportedAccount.ID)
		if err != nil {
			suite.FailNow(err.Error())
		}
		return dbAccount.DisplayName == ""
	}) {
		suite.FailNow("timed out waiting for account delete")
	}

	adminReport, err := suite.typeconverter.ReportToAdminAPIReport(context.Background(), suite.testReports["remote_account_1_report_local_account_2"], requestingAccount)
	suite.NoError(err)

	b, err := json.MarshalIndent(adminReport, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
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
      "statuses_count": 3,
      "last_status_at": "2021-09-11",
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
    "ip": "0.0.0.0",
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
    "confirmed": true,
    "approved": true,
    "disabled": false,
    "silenced": false,
    "suspended": true,
    "account": {
      "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
      "username": "1happyturtle",
      "acct": "1happyturtle",
      "display_name": "",
      "locked": true,
      "discoverable": false,
      "bot": false,
      "created_at": "2022-06-04T13:12:00.000Z",
      "note": "",
      "url": "http://localhost:8080/@1happyturtle",
      "avatar": "",
      "avatar_static": "",
      "header": "http://localhost:8080/assets/default_header.webp",
      "header_static": "http://localhost:8080/assets/default_header.webp",
      "header_description": "Flat gray background (default header).",
      "followers_count": 0,
      "following_count": 0,
      "statuses_count": 0,
      "last_status_at": null,
      "emojis": [],
      "fields": [],
      "suspended": true,
      "hide_collections": true
    }
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
      ]
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
      ]
    },
    "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
  },
  "statuses": [],
  "rules": [],
  "action_taken_comment": "user was warned not to be a turtle anymore"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestRelationshipFollowRequested() {
	var (
		ctx      = context.Background()
		account1 = suite.testAccounts["admin_account"]
		account2 = suite.testAccounts["local_account_2"]
	)

	// Put a follow request in the db from
	// admin account targeting local_account_2.
	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       account1.ID,
		TargetAccountID: account2.ID,
	}
	if err := suite.db.PutFollowRequest(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch the relationship from the database.
	dbRelationship, err := suite.state.DB.GetRelationship(ctx, account1.ID, account2.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check API model is set appropriately.
	relationship, err := suite.typeconverter.RelationshipToAPIRelationship(ctx, dbRelationship)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(relationship, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
  "following": false,
  "showing_reblogs": false,
  "notifying": false,
  "followed_by": false,
  "blocking": false,
  "blocked_by": false,
  "muting": false,
  "muting_notifications": false,
  "requested": true,
  "requested_by": false,
  "domain_blocking": false,
  "endorsed": false,
  "note": ""
}`, string(b))

	// Check relationship from the other side too.
	dbRelationship, err = suite.state.DB.GetRelationship(ctx, account2.ID, account1.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check API model is set appropriately.
	relationship, err = suite.typeconverter.RelationshipToAPIRelationship(ctx, dbRelationship)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err = json.MarshalIndent(relationship, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "id": "01F8MH17FWEB39HZJ76B6VXSKF",
  "following": false,
  "showing_reblogs": false,
  "notifying": false,
  "followed_by": false,
  "blocking": false,
  "blocked_by": false,
  "muting": false,
  "muting_notifications": false,
  "requested": false,
  "requested_by": true,
  "domain_blocking": false,
  "endorsed": false,
  "note": ""
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestIntReqToAPI() {
	requestingAccount := suite.testAccounts["local_account_2"]
	adminReport, err := suite.typeconverter.InteractionReqToAPIInteractionReq(
		context.Background(),
		suite.testInteractionRequests["admin_account_reply_turtle"],
		requestingAccount,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(adminReport, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "id": "01J5QVXCCEATJYSXM9H6MZT4JR",
  "type": "reply",
  "created_at": "2024-02-20T10:41:37.000Z",
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
    ]
  },
  "status": {
    "id": "01F8MHC8VWDRBQR0N1BATDDEM5",
    "created_at": "2021-10-20T10:40:37.000Z",
    "in_reply_to_id": null,
    "in_reply_to_account_id": null,
    "sensitive": true,
    "spoiler_text": "you won't be able to reply to this without my approval",
    "visibility": "public",
    "language": "en",
    "uri": "http://localhost:8080/users/1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
    "url": "http://localhost:8080/@1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
    "replies_count": 1,
    "reblogs_count": 0,
    "favourites_count": 0,
    "favourited": false,
    "reblogged": false,
    "muted": false,
    "bookmarked": false,
    "pinned": false,
    "content": "🐢 i don't mind people sharing and liking this one but I want to moderate replies to it 🐢",
    "reblog": null,
    "application": {
      "name": "kindaweird",
      "website": "https://kindaweird.app"
    },
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
      "statuses_count": 8,
      "last_status_at": "2021-07-28",
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
    "media_attachments": [],
    "mentions": [],
    "tags": [],
    "emojis": [],
    "card": null,
    "poll": null,
    "text": "🐢 i don't mind people sharing and liking this one but I want to moderate replies to it 🐢",
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
          "author",
          "me"
        ],
        "with_approval": [
          "public"
        ]
      },
      "can_reblog": {
        "always": [
          "public",
          "me"
        ],
        "with_approval": []
      }
    }
  },
  "reply": {
    "id": "01J5QVB9VC76NPPRQ207GG4DRZ",
    "created_at": "2024-02-20T10:41:37.000Z",
    "in_reply_to_id": "01F8MHC8VWDRBQR0N1BATDDEM5",
    "in_reply_to_account_id": "01F8MH5NBDF2MV7CTC4Q5128HF",
    "sensitive": false,
    "spoiler_text": "",
    "visibility": "public",
    "language": null,
    "uri": "http://localhost:8080/users/admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
    "url": "http://localhost:8080/@admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
    "replies_count": 0,
    "reblogs_count": 0,
    "favourites_count": 0,
    "favourited": false,
    "reblogged": false,
    "muted": false,
    "bookmarked": false,
    "pinned": false,
    "content": "\u003cp\u003eHi \u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003e1happyturtle\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e, can I reply?\u003c/p\u003e",
    "reblog": null,
    "application": {
      "name": "superseriousbusiness",
      "website": "https://superserious.business"
    },
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
      ]
    },
    "media_attachments": [],
    "mentions": [
      {
        "id": "01F8MH5NBDF2MV7CTC4Q5128HF",
        "username": "1happyturtle",
        "url": "http://localhost:8080/@1happyturtle",
        "acct": "1happyturtle"
      }
    ],
    "tags": [],
    "emojis": [],
    "card": null,
    "poll": null,
    "text": "Hi @1happyturtle, can I reply?",
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
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestConversationToAPISelfConvo() {
	var (
		ctx                                       = context.Background()
		requester                                 = suite.testAccounts["local_account_1"]
		lastStatus                                = suite.testStatuses["local_account_1_status_1"]
		filters    []*gtsmodel.Filter             = nil
		mutes      *usermute.CompiledUserMuteList = nil
	)

	convo := &gtsmodel.Conversation{
		ID:            "01J9C6K86PKZ5GY5WXV94DGH6R",
		CreatedAt:     testrig.TimeMustParse("2022-06-10T15:22:08Z"),
		UpdatedAt:     testrig.TimeMustParse("2022-06-10T15:22:08Z"),
		AccountID:     requester.ID,
		Account:       requester,
		OtherAccounts: nil,
		LastStatus:    lastStatus,
		Read:          util.Ptr(true),
	}

	apiConvo, err := suite.typeconverter.ConversationToAPIConversation(
		ctx,
		convo,
		requester,
		filters,
		mutes,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(apiConvo, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	// No other accounts involved, so we should only
	// have our own account in the "accounts" field.
	suite.Equal(`{
  "id": "01J9C6K86PKZ5GY5WXV94DGH6R",
  "unread": false,
  "accounts": [
    {
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
      "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
      "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
      "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
      "header_description": "A very old-school screenshot of the original team fortress mod for quake",
      "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
      "followers_count": 2,
      "following_count": 2,
      "statuses_count": 8,
      "last_status_at": "2024-01-10",
      "emojis": [],
      "fields": [],
      "enable_rss": true
    }
  ],
  "last_status": {
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
      "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
      "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
      "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
      "header_description": "A very old-school screenshot of the original team fortress mod for quake",
      "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
      "followers_count": 2,
      "following_count": 2,
      "statuses_count": 8,
      "last_status_at": "2024-01-10",
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
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestConversationToAPI() {
	var (
		ctx                                       = context.Background()
		requester                                 = suite.testAccounts["local_account_1"]
		lastStatus                                = suite.testStatuses["local_account_1_status_1"]
		filters    []*gtsmodel.Filter             = nil
		mutes      *usermute.CompiledUserMuteList = nil
	)

	convo := &gtsmodel.Conversation{
		ID:        "01J9C6K86PKZ5GY5WXV94DGH6R",
		CreatedAt: testrig.TimeMustParse("2022-06-10T15:22:08Z"),
		UpdatedAt: testrig.TimeMustParse("2022-06-10T15:22:08Z"),
		AccountID: requester.ID,
		Account:   requester,
		OtherAccounts: []*gtsmodel.Account{
			suite.testAccounts["local_account_2"],
		},
		LastStatus: lastStatus,
		Read:       util.Ptr(false),
	}

	apiConvo, err := suite.typeconverter.ConversationToAPIConversation(
		ctx,
		convo,
		requester,
		filters,
		mutes,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(apiConvo, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	// One other account is involved, so they
	// should in the "accounts" field and not us.
	suite.Equal(`{
  "id": "01J9C6K86PKZ5GY5WXV94DGH6R",
  "unread": true,
  "accounts": [
    {
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
      "statuses_count": 8,
      "last_status_at": "2021-07-28",
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
    }
  ],
  "last_status": {
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
      "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
      "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
      "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
      "header_description": "A very old-school screenshot of the original team fortress mod for quake",
      "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
      "followers_count": 2,
      "following_count": 2,
      "statuses_count": 8,
      "last_status_at": "2024-01-10",
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
  }
}`, string(b))
}

func TestInternalToFrontendTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToFrontendTestSuite))
}
