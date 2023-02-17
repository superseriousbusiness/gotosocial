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

package typeutils_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
  "avatar_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpg",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 5,
  "last_status_at": "2022-05-20T11:37:55.000Z",
  "emojis": [],
  "fields": [],
  "enable_rss": true,
  "role": {
    "name": "user"
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendWithEmojiStruct() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	testEmoji := suite.testEmojis["rainbow"]

	testAccount.Emojis = []*gtsmodel.Emoji{testEmoji}

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
  "avatar_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpg",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 5,
  "last_status_at": "2022-05-20T11:37:55.000Z",
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
  "enable_rss": true,
  "role": {
    "name": "user"
  }
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
  "avatar_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpg",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 5,
  "last_status_at": "2022-05-20T11:37:55.000Z",
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
  "enable_rss": true,
  "role": {
    "name": "user"
  }
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
  "avatar_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpg",
  "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 5,
  "last_status_at": "2022-05-20T11:37:55.000Z",
  "emojis": [],
  "fields": [],
  "source": {
    "privacy": "public",
    "sensitive": false,
    "language": "en",
    "status_format": "plain",
    "note": "hey yo this is my profile!",
    "fields": [],
    "follow_requests_count": 0
  },
  "enable_rss": true,
  "role": {
    "name": "user"
  }
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontend() {
	testStatus := suite.testStatuses["admin_account_status_1"]
	requestingAccount := suite.testAccounts["local_account_1"]
	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount)
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
  "replies_count": 0,
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
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 1,
    "following_count": 1,
    "statuses_count": 4,
    "last_status_at": "2021-10-20T10:41:37.000Z",
    "emojis": [],
    "fields": [],
    "enable_rss": true,
    "role": {
      "name": "admin"
    }
  },
  "media_attachments": [
    {
      "id": "01F8MH6NEM8D7527KZAECTCR76",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpg",
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
          "width": 256,
          "height": 134,
          "size": "256x134",
          "aspect": 1.9104477
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Black and white image of some 50's style text saying: Welcome On Board",
      "blurhash": "LNJRdVM{00Rj%Mayt7j[4nWBofRj"
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
  "text": "hello world! #welcome ! first post on the instance :rainbow: !"
}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontendUnknownLanguage() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["admin_account_status_1"]
	testStatus.Language = ""
	requestingAccount := suite.testAccounts["local_account_1"]
	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount)
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
  "replies_count": 0,
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
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 1,
    "following_count": 1,
    "statuses_count": 4,
    "last_status_at": "2021-10-20T10:41:37.000Z",
    "emojis": [],
    "fields": [],
    "enable_rss": true,
    "role": {
      "name": "admin"
    }
  },
  "media_attachments": [
    {
      "id": "01F8MH6NEM8D7527KZAECTCR76",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "text_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
      "preview_url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpg",
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
          "width": 256,
          "height": 134,
          "size": "256x134",
          "aspect": 1.9104477
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "Black and white image of some 50's style text saying: Welcome On Board",
      "blurhash": "LNJRdVM{00Rj%Mayt7j[4nWBofRj"
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
  "text": "hello world! #welcome ! first post on the instance :rainbow: !"
}`, string(b))
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
  "preview_url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01CDR64G398ADCHXK08WWTHEZ5.jpg",
  "remote_url": null,
  "preview_remote_url": null,
  "meta": {
    "original": {
      "width": 720,
      "height": 404,
      "frame_rate": "30/1",
      "duration": 15.033334,
      "bitrate": 1206522
    },
    "small": {
      "width": 720,
      "height": 404,
      "size": "720x404",
      "aspect": 1.7821782
    }
  },
  "description": "A cow adorably licking another cow!"
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

	suite.Equal(`{
  "uri": "http://localhost:8080",
  "account_domain": "localhost:8080",
  "title": "GoToSocial Testrig Instance",
  "description": "\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e",
  "short_description": "\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e",
  "email": "admin@example.org",
  "version": "0.0.0-testrig",
  "registrations": true,
  "approval_required": true,
  "invites_enabled": false,
  "configuration": {
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/png",
        "image/webp",
        "video/mp4"
      ],
      "image_size_limit": 10485760,
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
      "max_featured_tags": 10
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
    "status_count": 16,
    "user_count": 4
  },
  "thumbnail": "http://localhost:8080/assets/logo.png",
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
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 1,
    "following_count": 1,
    "statuses_count": 4,
    "last_status_at": "2021-10-20T10:41:37.000Z",
    "emojis": [],
    "fields": [],
    "enable_rss": true,
    "role": {
      "name": "admin"
    }
  },
  "max_toot_chars": 5000
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
  "description": "\u003cp\u003eThis is the GoToSocial testrig. It doesn't federate or anything.\u003c/p\u003e\u003cp\u003eWhen the testrig is shut down, all data on it will be deleted.\u003c/p\u003e\u003cp\u003eDon't use this in production!\u003c/p\u003e",
  "usage": {
    "users": {
      "active_month": 0
    }
  },
  "thumbnail": {
    "url": "http://localhost:8080/assets/logo.png"
  },
  "languages": [],
  "configuration": {
    "urls": {
      "streaming": "wss://localhost:8080"
    },
    "accounts": {
      "allow_custom_css": true,
      "max_featured_tags": 10
    },
    "statuses": {
      "max_characters": 5000,
      "max_media_attachments": 6,
      "characters_reserved_per_url": 25
    },
    "media_attachments": {
      "supported_mime_types": [
        "image/jpeg",
        "image/gif",
        "image/png",
        "image/webp",
        "video/mp4"
      ],
      "image_size_limit": 10485760,
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
      "header": "http://localhost:8080/assets/default_header.png",
      "header_static": "http://localhost:8080/assets/default_header.png",
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 4,
      "last_status_at": "2021-10-20T10:41:37.000Z",
      "emojis": [],
      "fields": [],
      "enable_rss": true,
      "role": {
        "name": "admin"
      }
    }
  },
  "rules": []
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
  "total_file_size": 47115,
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
  "total_file_size": 21697,
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
  "rule_ids": [],
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
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 1,
    "last_status_at": "2021-09-20T10:40:37.000Z",
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
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 1,
    "following_count": 1,
    "statuses_count": 7,
    "last_status_at": "2021-10-20T10:40:37.000Z",
    "emojis": [],
    "fields": [],
    "role": {
      "name": "user"
    }
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
      "name": "user"
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
    "role": {
      "name": "user"
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
      "header": "http://localhost:8080/assets/default_header.png",
      "header_static": "http://localhost:8080/assets/default_header.png",
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 7,
      "last_status_at": "2021-10-20T10:40:37.000Z",
      "emojis": [],
      "fields": [],
      "role": {
        "name": "user"
      }
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
    "role": {
      "name": "admin"
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
      "header": "http://localhost:8080/assets/default_header.png",
      "header_static": "http://localhost:8080/assets/default_header.png",
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 4,
      "last_status_at": "2021-10-20T10:41:37.000Z",
      "emojis": [],
      "fields": [],
      "enable_rss": true,
      "role": {
        "name": "admin"
      }
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
    "role": {
      "name": "admin"
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
      "header": "http://localhost:8080/assets/default_header.png",
      "header_static": "http://localhost:8080/assets/default_header.png",
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 4,
      "last_status_at": "2021-10-20T10:41:37.000Z",
      "emojis": [],
      "fields": [],
      "enable_rss": true,
      "role": {
        "name": "admin"
      }
    },
    "created_by_application_id": "01F8MGXQRHYF5QPMTMXP78QC2F"
  },
  "statuses": [],
  "rule_ids": [],
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
    "ip": "118.44.18.196",
    "ips": [],
    "locale": "en",
    "invite_request": "",
    "role": {
      "name": "user"
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
      "header": "http://localhost:8080/assets/default_header.png",
      "header_static": "http://localhost:8080/assets/default_header.png",
      "followers_count": 1,
      "following_count": 1,
      "statuses_count": 7,
      "last_status_at": "2021-10-20T10:40:37.000Z",
      "emojis": [],
      "fields": [],
      "role": {
        "name": "user"
      }
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
      "name": "user"
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
        "discoverable": true,
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
}`, string(b))
}

func TestInternalToFrontendTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToFrontendTestSuite))
}
