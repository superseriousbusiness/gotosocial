/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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

	b, err := json.Marshal(apiAccount)
	suite.NoError(err)
	suite.Equal(`{"id":"01F8MH1H7YV1Z7D2C8K2730QBF","username":"the_mighty_zork","acct":"the_mighty_zork","display_name":"original zork (he/they)","locked":false,"bot":false,"created_at":"2022-05-20T11:09:18.000Z","note":"\u003cp\u003ehey yo this is my profile!\u003c/p\u003e","url":"http://localhost:8080/@the_mighty_zork","avatar":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg","avatar_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg","header":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","header_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","followers_count":2,"following_count":2,"statuses_count":5,"last_status_at":"2022-05-20T11:37:55.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"user"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendWithEmojiStruct() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	testEmoji := suite.testEmojis["rainbow"]

	testAccount.Emojis = []*gtsmodel.Emoji{testEmoji}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.Marshal(apiAccount)
	suite.NoError(err)
	suite.Equal(`{"id":"01F8MH1H7YV1Z7D2C8K2730QBF","username":"the_mighty_zork","acct":"the_mighty_zork","display_name":"original zork (he/they)","locked":false,"bot":false,"created_at":"2022-05-20T11:09:18.000Z","note":"\u003cp\u003ehey yo this is my profile!\u003c/p\u003e","url":"http://localhost:8080/@the_mighty_zork","avatar":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg","avatar_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg","header":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","header_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","followers_count":2,"following_count":2,"statuses_count":5,"last_status_at":"2022-05-20T11:37:55.000Z","emojis":[{"shortcode":"rainbow","url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","static_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","visible_in_picker":true,"category":"reactions"}],"fields":[],"enable_rss":true,"role":"user"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendWithEmojiIDs() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	testEmoji := suite.testEmojis["rainbow"]

	testAccount.EmojiIDs = []string{testEmoji.ID}

	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.Marshal(apiAccount)
	suite.NoError(err)
	suite.Equal(`{"id":"01F8MH1H7YV1Z7D2C8K2730QBF","username":"the_mighty_zork","acct":"the_mighty_zork","display_name":"original zork (he/they)","locked":false,"bot":false,"created_at":"2022-05-20T11:09:18.000Z","note":"\u003cp\u003ehey yo this is my profile!\u003c/p\u003e","url":"http://localhost:8080/@the_mighty_zork","avatar":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg","avatar_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg","header":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","header_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","followers_count":2,"following_count":2,"statuses_count":5,"last_status_at":"2022-05-20T11:37:55.000Z","emojis":[{"shortcode":"rainbow","url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","static_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","visible_in_picker":true,"category":"reactions"}],"fields":[],"enable_rss":true,"role":"user"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontendSensitive() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	apiAccount, err := suite.typeconverter.AccountToAPIAccountSensitive(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.Marshal(apiAccount)
	suite.NoError(err)
	suite.Equal(`{"id":"01F8MH1H7YV1Z7D2C8K2730QBF","username":"the_mighty_zork","acct":"the_mighty_zork","display_name":"original zork (he/they)","locked":false,"bot":false,"created_at":"2022-05-20T11:09:18.000Z","note":"\u003cp\u003ehey yo this is my profile!\u003c/p\u003e","url":"http://localhost:8080/@the_mighty_zork","avatar":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg","avatar_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg","header":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","header_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","followers_count":2,"following_count":2,"statuses_count":5,"last_status_at":"2022-05-20T11:37:55.000Z","emojis":[],"fields":[],"source":{"privacy":"public","language":"en","status_format":"plain","note":"hey yo this is my profile!","fields":[]},"enable_rss":true,"role":"user"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestStatusToFrontend() {
	testStatus := suite.testStatuses["admin_account_status_1"]
	requestingAccount := suite.testAccounts["local_account_1"]
	apiStatus, err := suite.typeconverter.StatusToAPIStatus(context.Background(), testStatus, requestingAccount)
	suite.NoError(err)

	b, err := json.Marshal(apiStatus)
	suite.NoError(err)

	suite.Equal(`{"id":"01F8MH75CBF9JFX4ZAD54N0W0R","created_at":"2021-10-20T11:36:45.000Z","in_reply_to_id":null,"in_reply_to_account_id":null,"sensitive":false,"spoiler_text":"","visibility":"public","language":"en","uri":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R","url":"http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R","replies_count":0,"reblogs_count":0,"favourites_count":1,"favourited":true,"reblogged":false,"muted":false,"bookmarked":false,"pinned":false,"content":"hello world! #welcome ! first post on the instance :rainbow: !","reblog":null,"application":{"name":"superseriousbusiness","website":"https://superserious.business"},"account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[],"enable_rss":true,"role":"admin"},"media_attachments":[{"id":"01F8MH6NEM8D7527KZAECTCR76","type":"image","url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg","text_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg","preview_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpeg","remote_url":null,"preview_remote_url":null,"meta":{"original":{"width":1200,"height":630,"size":"1200x630","aspect":1.9047619},"small":{"width":256,"height":134,"size":"256x134","aspect":1.9104477},"focus":{"x":0,"y":0}},"description":"Black and white image of some 50's style text saying: Welcome On Board","blurhash":"LNJRdVM{00Rj%Mayt7j[4nWBofRj"}],"mentions":[],"tags":[{"name":"welcome","url":"http://localhost:8080/tags/welcome"}],"emojis":[{"shortcode":"rainbow","url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","static_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","visible_in_picker":true,"category":"reactions"}],"card":null,"poll":null,"text":"hello world! #welcome ! first post on the instance :rainbow: !"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestInstanceToFrontend() {
	testInstance := &gtsmodel.Instance{
		CreatedAt:        testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		UpdatedAt:        testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		Domain:           "example.org",
		Title:            "example instance",
		URI:              "https://example.org",
		ShortDescription: "a little description",
		Description:      "a much longer description",
		ContactEmail:     "someone@example.org",
		Version:          "software-from-hell 0.666",
	}

	apiInstance, err := suite.typeconverter.InstanceToAPIInstance(context.Background(), testInstance)
	suite.NoError(err)

	b, err := json.Marshal(apiInstance)
	suite.NoError(err)

	suite.Equal(`{"uri":"https://example.org","title":"example instance","description":"a much longer description","short_description":"a little description","email":"someone@example.org","version":"software-from-hell 0.666","registrations":false,"approval_required":false,"invites_enabled":false,"thumbnail":"","max_toot_chars":0}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestInstanceToFrontendWithAdminAccount() {
	testInstance := &gtsmodel.Instance{
		CreatedAt:        testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		UpdatedAt:        testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		Domain:           "example.org",
		Title:            "example instance",
		URI:              "https://example.org",
		ShortDescription: "a little description",
		Description:      "a much longer description",
		ContactEmail:     "someone@example.org",
		ContactAccountID: suite.testAccounts["remote_account_2"].ID,
		Version:          "software-from-hell 0.666",
	}

	apiInstance, err := suite.typeconverter.InstanceToAPIInstance(context.Background(), testInstance)
	suite.NoError(err)

	b, err := json.Marshal(apiInstance)
	suite.NoError(err)

	suite.Equal(`{"uri":"https://example.org","title":"example instance","description":"a much longer description","short_description":"a little description","email":"someone@example.org","version":"software-from-hell 0.666","registrations":false,"approval_required":false,"invites_enabled":false,"thumbnail":"","contact_account":{"id":"01FHMQX3GAABWSM0S2VZEC2SWC","username":"some_user","acct":"some_user@example.org","display_name":"some user","locked":true,"bot":false,"created_at":"2020-08-10T12:13:28.000Z","note":"i'm a real son of a gun","url":"http://example.org/@some_user","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":0,"following_count":0,"statuses_count":0,"last_status_at":null,"emojis":[],"fields":[]},"max_toot_chars":0}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestEmojiToFrontend() {
	emoji, err := suite.typeconverter.EmojiToAPIEmoji(context.Background(), suite.testEmojis["rainbow"])
	suite.NoError(err)

	b, err := json.Marshal(emoji)
	suite.NoError(err)

	suite.Equal(`{"shortcode":"rainbow","url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","static_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","visible_in_picker":true,"category":"reactions"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestEmojiToFrontendAdmin1() {
	emoji, err := suite.typeconverter.EmojiToAdminAPIEmoji(context.Background(), suite.testEmojis["rainbow"])
	suite.NoError(err)

	b, err := json.Marshal(emoji)
	suite.NoError(err)

	suite.Equal(`{"shortcode":"rainbow","url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","static_url":"http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png","visible_in_picker":true,"category":"reactions","id":"01F8MH9H8E4VG3KDYJR9EGPXCQ","disabled":false,"updated_at":"2021-09-20T10:40:37.000Z","total_file_size":47115,"content_type":"image/png","uri":"http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ"}`, string(b))
}

func (suite *InternalToFrontendTestSuite) TestEmojiToFrontendAdmin2() {
	emoji, err := suite.typeconverter.EmojiToAdminAPIEmoji(context.Background(), suite.testEmojis["yell"])
	suite.NoError(err)

	b, err := json.Marshal(emoji)
	suite.NoError(err)

	suite.Equal(`{"shortcode":"yell","url":"http://localhost:8080/fileserver/01GD5KR15NHTY8FZ01CD4D08XP/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png","static_url":"http://localhost:8080/fileserver/01GD5KR15NHTY8FZ01CD4D08XP/emoji/static/01GD5KP5CQEE1R3X43Y1EHS2CW.png","visible_in_picker":false,"id":"01GD5KP5CQEE1R3X43Y1EHS2CW","disabled":false,"domain":"fossbros-anonymous.io","updated_at":"2020-03-18T12:12:00.000Z","total_file_size":21697,"content_type":"image/png","uri":"http://fossbros-anonymous.io/emoji/01GD5KP5CQEE1R3X43Y1EHS2CW"}`, string(b))
}

func TestInternalToFrontendTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToFrontendTestSuite))
}
