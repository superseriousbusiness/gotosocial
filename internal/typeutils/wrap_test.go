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
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type WrapTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *WrapTestSuite) TestWrapNoteInCreateIRIOnly() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	note, err := suite.typeconverter.StatusToAS(context.Background(), testStatus)
	suite.NoError(err)

	create := typeutils.WrapStatusableInCreate(note, true)
	suite.NoError(err)
	suite.NotNil(create)

	createI, err := ap.Serialize(create)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(createI, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "cc": "http://localhost:8080/users/the_mighty_zork/followers",
  "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity#Create",
  "object": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "published": "2021-10-20T12:40:37+02:00",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Create"
}`, string(bytes))
}

func (suite *WrapTestSuite) TestWrapNoteInCreate() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	note, err := suite.typeconverter.StatusToAS(context.Background(), testStatus)
	suite.NoError(err)

	create := typeutils.WrapStatusableInCreate(note, false)
	suite.NoError(err)
	suite.NotNil(create)

	createI, err := ap.Serialize(create)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(createI, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "sensitive": "as:sensitive"
    }
  ],
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "cc": "http://localhost:8080/users/the_mighty_zork/followers",
  "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity#Create",
  "object": {
    "attachment": [],
    "attributedTo": "http://localhost:8080/users/the_mighty_zork",
    "cc": "http://localhost:8080/users/the_mighty_zork/followers",
    "content": "hello everyone!",
    "contentMap": {
      "en": "hello everyone!"
    },
    "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
    "interactionPolicy": {
      "canAnnounce": {
        "always": [
          "https://www.w3.org/ns/activitystreams#Public"
        ],
        "approvalRequired": []
      },
      "canLike": {
        "always": [
          "https://www.w3.org/ns/activitystreams#Public"
        ],
        "approvalRequired": []
      },
      "canReply": {
        "always": [
          "https://www.w3.org/ns/activitystreams#Public"
        ],
        "approvalRequired": []
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
  },
  "published": "2021-10-20T12:40:37+02:00",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Create"
}`, string(bytes))
}

func (suite *WrapTestSuite) TestWrapAccountableInUpdate() {
	testAccount := suite.testAccounts["local_account_1"]

	accountable, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	if err != nil {
		suite.FailNow(err.Error())
	}

	create, err := suite.typeconverter.WrapAccountableInUpdate(accountable)
	if err != nil {
		suite.FailNow(err.Error())
	}

	createI, err := ap.Serialize(create)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Get the ID as it's not determinate.
	createID := ap.GetJSONLDId(create)

	bytes, err := json.MarshalIndent(createI, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

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
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "bcc": "http://localhost:8080/users/the_mighty_zork/followers",
  "id": "`+createID.String()+`",
  "object": {
    "discoverable": true,
    "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
    "followers": "http://localhost:8080/users/the_mighty_zork/followers",
    "following": "http://localhost:8080/users/the_mighty_zork/following",
    "icon": {
      "mediaType": "image/jpeg",
      "type": "Image",
      "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
    },
    "id": "http://localhost:8080/users/the_mighty_zork",
    "image": {
      "mediaType": "image/jpeg",
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
      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
    },
    "published": "2022-05-20T11:09:18Z",
    "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
    "tag": [],
    "type": "Person",
    "url": "http://localhost:8080/@the_mighty_zork"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Update"
}`, string(bytes))
}

func TestWrapTestSuite(t *testing.T) {
	suite.Run(t, new(WrapTestSuite))
}
