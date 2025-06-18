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
	"encoding/json"
	"errors"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type InternalToASTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToASTestSuite) TestAccountToAS() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "discoverable": true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "name": "a green goblin looking nasty",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "name": "A very old-school screenshot of the original team fortress mod for quake",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqtQQjwFLHPez+7uF9AX7\nuvLFHm3SyNIozhhVmGhxHIs0xdgRnZKmzmZkFdrFuXddBTAglU4C2u3dw10jJO1a\nWIFQF8bGkRHZG7Pd25/XmWWBRPmOJxNLeWBqpj0G+2zTMgnAV72hALSDFY2/QDsx\nUthenKw0Srpj1LUwvRbyVQQ8fGu4v0HACFnlOX2hCQwhfAnGrb0V70Y2IJu++MP7\n6i49md0vR0Mv3WbsEJUNp1fTIUzkgWB31icvfrNmaaAxw5FkAE+KfkkylhRxi5i5\nRR1XQUINWc2Kj2Kro+CJarKG+9zasMyN7+D230gpESi8rXv1SwRu865FR3gANdDS\nMwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-05-20T11:09:18Z",
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestAccountToASBot() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test

	// Update zork to be a bot.
	testAccount.ActorType = gtsmodel.AccountActorTypeApplication
	if err := suite.state.DB.UpdateAccount(suite.T().Context(), testAccount); err != nil {
		suite.FailNow(err.Error())
	}

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "discoverable": true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "name": "a green goblin looking nasty",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "name": "A very old-school screenshot of the original team fortress mod for quake",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqtQQjwFLHPez+7uF9AX7\nuvLFHm3SyNIozhhVmGhxHIs0xdgRnZKmzmZkFdrFuXddBTAglU4C2u3dw10jJO1a\nWIFQF8bGkRHZG7Pd25/XmWWBRPmOJxNLeWBqpj0G+2zTMgnAV72hALSDFY2/QDsx\nUthenKw0Srpj1LUwvRbyVQQ8fGu4v0HACFnlOX2hCQwhfAnGrb0V70Y2IJu++MP7\n6i49md0vR0Mv3WbsEJUNp1fTIUzkgWB31icvfrNmaaAxw5FkAE+KfkkylhRxi5i5\nRR1XQUINWc2Kj2Kro+CJarKG+9zasMyN7+D230gpESi8rXv1SwRu865FR3gANdDS\nMwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-05-20T11:09:18Z",
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Application",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestAccountToASWithFields() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_2"]

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "PropertyValue": "schema:PropertyValue",
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "schema": "http://schema.org#",
      "toot": "http://joinmastodon.org/ns#",
      "value": "schema:value"
    }
  ],
  "attachment": [
    {
      "name": "should you follow me?",
      "type": "PropertyValue",
      "value": "maybe!"
    },
    {
      "name": "age",
      "type": "PropertyValue",
      "value": "120"
    }
  ],
  "discoverable": false,
  "featured": "http://localhost:8080/users/1happyturtle/collections/featured",
  "followers": "http://localhost:8080/users/1happyturtle/followers",
  "following": "http://localhost:8080/users/1happyturtle/following",
  "id": "http://localhost:8080/users/1happyturtle",
  "inbox": "http://localhost:8080/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://localhost:8080/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {
    "id": "http://localhost:8080/users/1happyturtle#main-key",
    "owner": "http://localhost:8080/users/1happyturtle",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-06-04T13:12:00Z",
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@1happyturtle"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestAccountToASAliasedAndMoved() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test

	ctx := suite.T().Context()

	// Suppose zork has moved account to turtle.
	testAccount.AlsoKnownAsURIs = []string{"http://localhost:8080/users/1happyturtle"}
	testAccount.MovedToURI = "http://localhost:8080/users/1happyturtle"
	if err := suite.state.DB.UpdateAccount(ctx,
		testAccount,
		"also_known_as_uris",
		"moved_to_uri",
	); err != nil {
		suite.FailNow(err.Error())
	}

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "movedTo": {
        "@id": "as:movedTo",
        "@type": "@id"
      },
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "alsoKnownAs": [
    "http://localhost:8080/users/1happyturtle"
  ],
  "discoverable": true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "name": "a green goblin looking nasty",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "name": "A very old-school screenshot of the original team fortress mod for quake",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "movedTo": "http://localhost:8080/users/1happyturtle",
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqtQQjwFLHPez+7uF9AX7\nuvLFHm3SyNIozhhVmGhxHIs0xdgRnZKmzmZkFdrFuXddBTAglU4C2u3dw10jJO1a\nWIFQF8bGkRHZG7Pd25/XmWWBRPmOJxNLeWBqpj0G+2zTMgnAV72hALSDFY2/QDsx\nUthenKw0Srpj1LUwvRbyVQQ8fGu4v0HACFnlOX2hCQwhfAnGrb0V70Y2IJu++MP7\n6i49md0vR0Mv3WbsEJUNp1fTIUzkgWB31icvfrNmaaAxw5FkAE+KfkkylhRxi5i5\nRR1XQUINWc2Kj2Kro+CJarKG+9zasMyN7+D230gpESi8rXv1SwRu865FR3gANdDS\nMwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-05-20T11:09:18Z",
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestAccountToASWithOneField() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_2"]
	testAccount.Fields = testAccount.Fields[0:1] // Take only one field.

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// Despite only one field being set, attachments should still be a slice/array.
	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "PropertyValue": "schema:PropertyValue",
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "schema": "http://schema.org#",
      "toot": "http://joinmastodon.org/ns#",
      "value": "schema:value"
    }
  ],
  "attachment": [
    {
      "name": "should you follow me?",
      "type": "PropertyValue",
      "value": "maybe!"
    }
  ],
  "discoverable": false,
  "featured": "http://localhost:8080/users/1happyturtle/collections/featured",
  "followers": "http://localhost:8080/users/1happyturtle/followers",
  "following": "http://localhost:8080/users/1happyturtle/following",
  "id": "http://localhost:8080/users/1happyturtle",
  "inbox": "http://localhost:8080/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://localhost:8080/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {
    "id": "http://localhost:8080/users/1happyturtle#main-key",
    "owner": "http://localhost:8080/users/1happyturtle",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-06-04T13:12:00Z",
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@1happyturtle"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestAccountToASWithEmoji() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test
	testAccount.Emojis = []*gtsmodel.Emoji{suite.testEmojis["rainbow"]}

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "discoverable": true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "name": "a green goblin looking nasty",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "name": "A very old-school screenshot of the original team fortress mod for quake",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqtQQjwFLHPez+7uF9AX7\nuvLFHm3SyNIozhhVmGhxHIs0xdgRnZKmzmZkFdrFuXddBTAglU4C2u3dw10jJO1a\nWIFQF8bGkRHZG7Pd25/XmWWBRPmOJxNLeWBqpj0G+2zTMgnAV72hALSDFY2/QDsx\nUthenKw0Srpj1LUwvRbyVQQ8fGu4v0HACFnlOX2hCQwhfAnGrb0V70Y2IJu++MP7\n6i49md0vR0Mv3WbsEJUNp1fTIUzkgWB31icvfrNmaaAxw5FkAE+KfkkylhRxi5i5\nRR1XQUINWc2Kj2Kro+CJarKG+9zasMyN7+D230gpESi8rXv1SwRu865FR3gANdDS\nMwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-05-20T11:09:18Z",
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
    },
    "id": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
    "name": ":rainbow:",
    "type": "Emoji",
    "updated": "2021-09-20T12:40:37+02:00"
  },
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestAccountToASWithSharedInbox() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test
	sharedInbox := "http://localhost:8080/sharedInbox"
	testAccount.SharedInboxURI = &sharedInbox

	accountable, err := suite.typeconverter.AccountToAS(suite.T().Context(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(accountable)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    {
      "discoverable": "toot:discoverable",
      "featured": {
        "@id": "toot:featured",
        "@type": "@id"
      },
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "discoverable": true,
  "endpoints": {
    "sharedInbox": "http://localhost:8080/sharedInbox"
  },
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "name": "a green goblin looking nasty",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "name": "A very old-school screenshot of the original team fortress mod for quake",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqtQQjwFLHPez+7uF9AX7\nuvLFHm3SyNIozhhVmGhxHIs0xdgRnZKmzmZkFdrFuXddBTAglU4C2u3dw10jJO1a\nWIFQF8bGkRHZG7Pd25/XmWWBRPmOJxNLeWBqpj0G+2zTMgnAV72hALSDFY2/QDsx\nUthenKw0Srpj1LUwvRbyVQQ8fGu4v0HACFnlOX2hCQwhfAnGrb0V70Y2IJu++MP7\n6i49md0vR0Mv3WbsEJUNp1fTIUzkgWB31icvfrNmaaAxw5FkAE+KfkkylhRxi5i5\nRR1XQUINWc2Kj2Kro+CJarKG+9zasMyN7+D230gpESi8rXv1SwRu865FR3gANdDS\nMwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-05-20T11:09:18Z",
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToAS() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	ctx := suite.T().Context()

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "sensitive": "as:sensitive"
    }
  ],
  "attachment": [],
  "attributedTo": "http://localhost:8080/users/the_mighty_zork",
  "cc": "http://localhost:8080/users/the_mighty_zork/followers",
  "content": "\u003cp\u003ehello everyone!\u003c/p\u003e",
  "contentMap": {
    "en": "\u003cp\u003ehello everyone!\u003c/p\u003e"
  },
  "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "interactionPolicy": {
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    }
  },
  "published": "2021-10-20T12:40:37+02:00",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?page=true",
      "next": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies",
    "type": "Collection"
  },
  "sensitive": true,
  "summary": "introduction post",
  "tag": [],
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusWithTagsToASWithIDs() {
	// use the status with just IDs of attachments and emojis pinned on it
	testStatus := suite.testStatuses["admin_account_status_1"]
	ctx := suite.T().Context()

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "Hashtag": "as:Hashtag",
      "blurhash": "toot:blurhash",
      "focalPoint": {
        "@container": "@list",
        "@id": "toot:focalPoint"
      },
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "attachment": [
    {
      "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj",
      "focalPoint": [
        -0.5,
        0.5
      ],
      "mediaType": "image/jpeg",
      "summary": "Black and white image of some 50's style text saying: Welcome On Board",
      "type": "Image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg"
    }
  ],
  "attributedTo": "http://localhost:8080/users/admin",
  "cc": "http://localhost:8080/users/admin/followers",
  "content": "\u003cp\u003ehello world! \u003ca href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\"\u003e#\u003cspan\u003ewelcome\u003c/span\u003e\u003c/a\u003e ! first post on the instance :rainbow: !\u003c/p\u003e",
  "contentMap": {
    "en": "\u003cp\u003ehello world! \u003ca href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\"\u003e#\u003cspan\u003ewelcome\u003c/span\u003e\u003c/a\u003e ! first post on the instance :rainbow: !\u003c/p\u003e"
  },
  "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "interactionPolicy": {
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    }
  },
  "published": "2021-10-20T11:36:45Z",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?page=true",
      "next": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
    "type": "Collection"
  },
  "sensitive": false,
  "summary": "",
  "tag": [
    {
      "icon": {
        "mediaType": "image/png",
        "type": "Image",
        "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
      },
      "id": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
      "name": ":rainbow:",
      "type": "Emoji",
      "updated": "2021-09-20T10:40:37Z"
    },
    {
      "href": "http://localhost:8080/tags/welcome",
      "name": "#welcome",
      "type": "Hashtag"
    }
  ],
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusWithTagsToASFromDB() {
	ctx := suite.T().Context()
	// get the entire status with all tags
	testStatus, err := suite.db.GetStatusByID(ctx, suite.testStatuses["admin_account_status_1"].ID)
	suite.NoError(err)

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "Hashtag": "as:Hashtag",
      "blurhash": "toot:blurhash",
      "focalPoint": {
        "@container": "@list",
        "@id": "toot:focalPoint"
      },
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "attachment": [
    {
      "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj",
      "focalPoint": [
        -0.5,
        0.5
      ],
      "mediaType": "image/jpeg",
      "summary": "Black and white image of some 50's style text saying: Welcome On Board",
      "type": "Image",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg"
    }
  ],
  "attributedTo": "http://localhost:8080/users/admin",
  "cc": "http://localhost:8080/users/admin/followers",
  "content": "\u003cp\u003ehello world! \u003ca href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\"\u003e#\u003cspan\u003ewelcome\u003c/span\u003e\u003c/a\u003e ! first post on the instance :rainbow: !\u003c/p\u003e",
  "contentMap": {
    "en": "\u003cp\u003ehello world! \u003ca href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\"\u003e#\u003cspan\u003ewelcome\u003c/span\u003e\u003c/a\u003e ! first post on the instance :rainbow: !\u003c/p\u003e"
  },
  "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "interactionPolicy": {
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    }
  },
  "published": "2021-10-20T11:36:45Z",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?page=true",
      "next": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
    "type": "Collection"
  },
  "sensitive": false,
  "summary": "",
  "tag": [
    {
      "icon": {
        "mediaType": "image/png",
        "type": "Image",
        "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
      },
      "id": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
      "name": ":rainbow:",
      "type": "Emoji",
      "updated": "2021-09-20T10:40:37Z"
    },
    {
      "href": "http://localhost:8080/tags/welcome",
      "name": "#welcome",
      "type": "Hashtag"
    }
  ],
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASWithMentions() {
	testStatusID := suite.testStatuses["admin_account_status_3"].ID
	ctx := suite.T().Context()

	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "sensitive": "as:sensitive"
    }
  ],
  "attachment": [],
  "attributedTo": "http://localhost:8080/users/admin",
  "cc": [
    "http://localhost:8080/users/admin/followers",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "content": "\u003cp\u003ehi \u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003ethe_mighty_zork\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e welcome to the instance!\u003c/p\u003e",
  "contentMap": {
    "en": "\u003cp\u003ehi \u003cspan class=\"h-card\"\u003e\u003ca href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\"\u003e@\u003cspan\u003ethe_mighty_zork\u003c/span\u003e\u003c/a\u003e\u003c/span\u003e welcome to the instance!\u003c/p\u003e"
  },
  "id": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "inReplyTo": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "interactionPolicy": {
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": [],
      "automaticApproval": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "manualApproval": []
    }
  },
  "published": "2021-11-20T13:32:16Z",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies?page=true",
      "next": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies",
    "type": "Collection"
  },
  "sensitive": false,
  "summary": "",
  "tag": {
    "href": "http://localhost:8080/users/the_mighty_zork",
    "name": "@the_mighty_zork@localhost:8080",
    "type": "Mention"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeletePublicReply() {
	testStatus := suite.testStatuses["admin_account_status_3"]
	ctx := suite.T().Context()

	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/admin",
  "cc": [
    "http://localhost:8080/users/admin/followers",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "object": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeletePublicReplyOriginalDeleted() {
	testStatus := suite.testStatuses["admin_account_status_3"]
	ctx := suite.T().Context()

	// Delete the status this replies to.
	if err := suite.db.DeleteStatusByID(ctx, testStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Delete the mention the reply created.
	mention := suite.testMentions["admin_account_mention_zork"]
	if err := suite.db.DeleteByID(ctx, mention.ID, mention); err != nil {
		suite.FailNow(err.Error())
	}

	// The delete should still be created OK.
	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/admin",
  "cc": [
    "http://localhost:8080/users/admin/followers",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "object": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeletePublic() {
	testStatus := suite.testStatuses["admin_account_status_1"]
	ctx := suite.T().Context()

	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/admin",
  "cc": "http://localhost:8080/users/admin/followers",
  "object": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeleteDirectMessage() {
	testStatus := suite.testStatuses["local_account_2_status_6"]
	ctx := suite.T().Context()

	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/1happyturtle",
  "cc": [],
  "object": "http://localhost:8080/users/1happyturtle/statuses/01FN3VJGFH10KR7S2PB0GFJZYG",
  "to": "http://localhost:8080/users/the_mighty_zork",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusesToASOutboxPage() {
	testAccount := suite.testAccounts["admin_account"]
	ctx := suite.T().Context()

	// get public statuses from testaccount
	statuses, err := suite.db.GetAccountStatuses(ctx, testAccount.ID, 30, true, true, "", "", false, true)
	suite.NoError(err)

	page, err := suite.typeconverter.StatusesToASOutboxPage(ctx, testAccount.OutboxURI, "", "", statuses)
	suite.NoError(err)

	ser, err := ap.Serialize(page)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/admin/outbox?page=true",
  "next": "http://localhost:8080/users/admin/outbox?page=true\u0026max_id=01F8MH75CBF9JFX4ZAD54N0W0R",
  "orderedItems": [
    {
      "actor": "http://localhost:8080/users/admin",
      "cc": "http://localhost:8080/users/admin/followers",
      "id": "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37/activity#Create",
      "object": "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
      "published": "2021-10-20T12:36:45Z",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    },
    {
      "actor": "http://localhost:8080/users/admin",
      "cc": "http://localhost:8080/users/admin/followers",
      "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/activity#Create",
      "object": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
      "published": "2021-10-20T11:36:45Z",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    }
  ],
  "partOf": "http://localhost:8080/users/admin/outbox",
  "prev": "http://localhost:8080/users/admin/outbox?page=true\u0026min_id=01F8MHAAY43M6RJ473VQFCVH37",
  "type": "OrderedCollectionPage"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestSelfBoostFollowersOnlyToAS() {
	ctx := suite.T().Context()

	testStatus := suite.testStatuses["local_account_1_status_5"]
	testAccount := suite.testAccounts["local_account_1"]

	boostWrapperStatus, err := suite.typeconverter.StatusToBoost(ctx, testStatus, testAccount, "")
	suite.NoError(err)
	suite.NotNil(boostWrapperStatus)

	// Set some fields to predictable values for the test.
	boostWrapperStatus.ID = "01G74JJ1KS331G2JXHRMZCE0ER"
	boostWrapperStatus.URI = "http://localhost:8080/users/the_mighty_zork/statuses/01G74JJ1KS331G2JXHRMZCE0ER"
	boostWrapperStatus.CreatedAt = testrig.TimeMustParse("2022-06-09T13:12:00Z")

	asBoost, err := suite.typeconverter.BoostToAS(ctx, boostWrapperStatus, testAccount, testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asBoost)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "cc": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/users/the_mighty_zork/statuses/01G74JJ1KS331G2JXHRMZCE0ER",
  "object": "http://localhost:8080/users/the_mighty_zork/statuses/01FCTA44PW9H1TB328S9AQXKDS",
  "published": "2022-06-09T13:12:00Z",
  "to": "http://localhost:8080/users/the_mighty_zork/followers",
  "type": "Announce"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestReportToAS() {
	ctx := suite.T().Context()

	testReport := suite.testReports["local_account_2_report_remote_account_1"]
	account := suite.testAccounts["local_account_2"]
	targetAccount := suite.testAccounts["remote_account_1"]
	statuses := []*gtsmodel.Status{suite.testStatuses["remote_account_1_status_1"]}

	testReport.Account = account
	testReport.TargetAccount = targetAccount
	testReport.Statuses = statuses

	flag, err := suite.typeconverter.ReportToASFlag(ctx, testReport)
	suite.NoError(err)

	ser, err := ap.Serialize(flag)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/localhost:8080",
  "content": "dark souls sucks, please yeet this nerd",
  "id": "http://localhost:8080/reports/01GP3AWY4CRDVRNZKW0TEAMB5R",
  "object": [
    "http://fossbros-anonymous.io/users/foss_satan",
    "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M"
  ],
  "type": "Flag"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPinnedStatusesToASSomeItems() {
	ctx := suite.T().Context()

	testAccount := suite.testAccounts["admin_account"]
	statuses, err := suite.db.GetAccountPinnedStatuses(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	collection, err := suite.typeconverter.StatusesToASFeaturedCollection(ctx, testAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/admin/collections/featured",
  "orderedItems": [
    "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
    "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"
  ],
  "totalItems": 2,
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPinnedStatusesToASNoItems() {
	ctx := suite.T().Context()

	testAccount := suite.testAccounts["local_account_1"]
	statuses, err := suite.db.GetAccountPinnedStatuses(ctx, testAccount.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	collection, err := suite.typeconverter.StatusesToASFeaturedCollection(ctx, testAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "orderedItems": [],
  "totalItems": 0,
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPinnedStatusesToASOneItem() {
	ctx := suite.T().Context()

	testAccount := suite.testAccounts["local_account_2"]
	statuses, err := suite.db.GetAccountPinnedStatuses(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	collection, err := suite.typeconverter.StatusesToASFeaturedCollection(ctx, testAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/1happyturtle/collections/featured",
  "orderedItems": [
    "http://localhost:8080/users/1happyturtle/statuses/01G20ZM733MGN8J344T4ZDDFY1"
  ],
  "totalItems": 1,
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPollVoteToASCreate() {
	vote := suite.testPollVotes["remote_account_1_status_2_poll_vote_local_account_1"]

	creates, err := suite.typeconverter.PollVoteToASCreates(suite.T().Context(), vote)
	suite.NoError(err)
	suite.Len(creates, 2)

	createI0, err := ap.Serialize(creates[0])
	suite.NoError(err)

	createI1, err := ap.Serialize(creates[1])
	suite.NoError(err)

	bytes0, err := json.MarshalIndent(createI0, "", "  ")
	suite.NoError(err)

	bytes1, err := json.MarshalIndent(createI1, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/users/the_mighty_zork/activity#vote0/http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
  "object": {
    "attributedTo": "http://localhost:8080/users/the_mighty_zork",
    "id": "http://localhost:8080/users/the_mighty_zork#01HEN2R65468ZG657C4ZPHJ4EX/votes/1",
    "inReplyTo": "http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
    "name": "tissues",
    "to": "http://fossbros-anonymous.io/users/foss_satan",
    "type": "Note"
  },
  "published": "2021-09-11T11:45:37+02:00",
  "to": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Create"
}`, string(bytes0))

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/users/the_mighty_zork/activity#vote1/http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
  "object": {
    "attributedTo": "http://localhost:8080/users/the_mighty_zork",
    "id": "http://localhost:8080/users/the_mighty_zork#01HEN2R65468ZG657C4ZPHJ4EX/votes/2",
    "inReplyTo": "http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
    "name": "financial times",
    "to": "http://fossbros-anonymous.io/users/foss_satan",
    "type": "Note"
  },
  "published": "2021-09-11T11:45:37+02:00",
  "to": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Create"
}`, string(bytes1))
}

func (suite *InternalToASTestSuite) TestInteractionReqToASAcceptAnnounce() {
	acceptingAccount := suite.testAccounts["local_account_1"]
	interactingAccount := suite.testAccounts["remote_account_1"]

	req := &gtsmodel.InteractionRequest{
		ID:                   "01J1AKMZ8JE5NW0ZSFTRC1JJNE",
		CreatedAt:            testrig.TimeMustParse("2022-06-09T13:12:00Z"),
		StatusID:             "01JJYCVKCXB9JTQD1XW2KB8MT3",
		Status:               &gtsmodel.Status{URI: "http://localhost:8080/users/the_mighty_zork/statuses/01JJYCVKCXB9JTQD1XW2KB8MT3"},
		TargetAccountID:      acceptingAccount.ID,
		TargetAccount:        acceptingAccount,
		InteractingAccountID: interactingAccount.ID,
		InteractingAccount:   interactingAccount,
		InteractionURI:       "https://fossbros-anonymous.io/users/foss_satan/statuses/01J1AKRRHQ6MDDQHV0TP716T2K",
		InteractionType:      gtsmodel.InteractionAnnounce,
		URI:                  "http://localhost:8080/users/the_mighty_zork/accepts/01J1AKMZ8JE5NW0ZSFTRC1JJNE",
		AcceptedAt:           testrig.TimeMustParse("2022-06-09T13:12:00Z"),
	}

	accept, err := suite.typeconverter.InteractionReqToASAccept(
		suite.T().Context(),
		req,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	i, err := ap.Serialize(accept)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "http://localhost:8080/users/the_mighty_zork/followers"
  ],
  "id": "http://localhost:8080/users/the_mighty_zork/accepts/01J1AKMZ8JE5NW0ZSFTRC1JJNE",
  "object": "https://fossbros-anonymous.io/users/foss_satan/statuses/01J1AKRRHQ6MDDQHV0TP716T2K",
  "target": "http://localhost:8080/users/the_mighty_zork/statuses/01JJYCVKCXB9JTQD1XW2KB8MT3",
  "to": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Accept"
}`, string(b))
}

func (suite *InternalToASTestSuite) TestInteractionReqToASAcceptLike() {
	acceptingAccount := suite.testAccounts["local_account_1"]
	interactingAccount := suite.testAccounts["remote_account_1"]

	req := &gtsmodel.InteractionRequest{
		ID:                   "01J1AKMZ8JE5NW0ZSFTRC1JJNE",
		CreatedAt:            testrig.TimeMustParse("2022-06-09T13:12:00Z"),
		StatusID:             "01JJYCVKCXB9JTQD1XW2KB8MT3",
		Status:               &gtsmodel.Status{URI: "http://localhost:8080/users/the_mighty_zork/statuses/01JJYCVKCXB9JTQD1XW2KB8MT3"},
		TargetAccountID:      acceptingAccount.ID,
		TargetAccount:        acceptingAccount,
		InteractingAccountID: interactingAccount.ID,
		InteractingAccount:   interactingAccount,
		InteractionURI:       "https://fossbros-anonymous.io/users/foss_satan/statuses/01J1AKRRHQ6MDDQHV0TP716T2K",
		InteractionType:      gtsmodel.InteractionLike,
		URI:                  "http://localhost:8080/users/the_mighty_zork/accepts/01J1AKMZ8JE5NW0ZSFTRC1JJNE",
		AcceptedAt:           testrig.TimeMustParse("2022-06-09T13:12:00Z"),
	}

	accept, err := suite.typeconverter.InteractionReqToASAccept(
		suite.T().Context(),
		req,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	i, err := ap.Serialize(accept)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/users/the_mighty_zork/accepts/01J1AKMZ8JE5NW0ZSFTRC1JJNE",
  "object": "https://fossbros-anonymous.io/users/foss_satan/statuses/01J1AKRRHQ6MDDQHV0TP716T2K",
  "target": "http://localhost:8080/users/the_mighty_zork/statuses/01JJYCVKCXB9JTQD1XW2KB8MT3",
  "to": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Accept"
}`, string(b))
}

func TestInternalToASTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToASTestSuite))
}
