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

package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type DereferenceTestSuite struct {
	TransportTestSuite
}

func (suite *DereferenceTestSuite) TestDerefLocalUser() {
	iri := testrig.URLMustParse(suite.testAccounts["local_account_1"].URI)

	resp, err := suite.transport.Dereference(context.Background(), iri)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.EqualValues(1887, resp.ContentLength)
	suite.Equal("1887", resp.Header.Get("Content-Length"))
	suite.Equal(apiutil.AppActivityLDJSON, resp.Header.Get("Content-Type"))

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	dst := bytes.Buffer{}
	if err := json.Indent(&dst, b, "", "  "); err != nil {
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
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqtQQjwFLHPez+7uF9AX7\nuvLFHm3SyNIozhhVmGhxHIs0xdgRnZKmzmZkFdrFuXddBTAglU4C2u3dw10jJO1a\nWIFQF8bGkRHZG7Pd25/XmWWBRPmOJxNLeWBqpj0G+2zTMgnAV72hALSDFY2/QDsx\nUthenKw0Srpj1LUwvRbyVQQ8fGu4v0HACFnlOX2hCQwhfAnGrb0V70Y2IJu++MP7\n6i49md0vR0Mv3WbsEJUNp1fTIUzkgWB31icvfrNmaaAxw5FkAE+KfkkylhRxi5i5\nRR1XQUINWc2Kj2Kro+CJarKG+9zasMyN7+D230gpESi8rXv1SwRu865FR3gANdDS\nMwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "published": "2022-05-20T11:09:18Z",
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, dst.String())
}

func (suite *DereferenceTestSuite) TestDerefLocalStatus() {
	iri := testrig.URLMustParse(suite.testStatuses["local_account_1_status_1"].URI)

	resp, err := suite.transport.Dereference(context.Background(), iri)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.EqualValues(1769, resp.ContentLength)
	suite.Equal("1769", resp.Header.Get("Content-Length"))
	suite.Equal(apiutil.AppActivityLDJSON, resp.Header.Get("Content-Type"))

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	dst := bytes.Buffer{}
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

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
  "published": "2021-10-20T10:40:37Z",
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
}`, dst.String())
}

func (suite *DereferenceTestSuite) TestDerefLocalFollowers() {
	iri := testrig.URLMustParse(suite.testAccounts["local_account_1"].FollowersURI)

	resp, err := suite.transport.Dereference(context.Background(), iri)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.EqualValues(161, resp.ContentLength)
	suite.Equal("161", resp.Header.Get("Content-Length"))
	suite.Equal(apiutil.AppActivityLDJSON, resp.Header.Get("Content-Type"))

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	dst := bytes.Buffer{}
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "items": [
    "http://localhost:8080/users/1happyturtle",
    "http://localhost:8080/users/admin"
  ],
  "type": "Collection"
}`, dst.String())
}

func (suite *DereferenceTestSuite) TestDerefLocalFollowing() {
	iri := testrig.URLMustParse(suite.testAccounts["local_account_1"].FollowingURI)

	resp, err := suite.transport.Dereference(context.Background(), iri)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.EqualValues(161, resp.ContentLength)
	suite.Equal("161", resp.Header.Get("Content-Length"))
	suite.Equal(apiutil.AppActivityLDJSON, resp.Header.Get("Content-Type"))

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	dst := bytes.Buffer{}
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "items": [
    "http://localhost:8080/users/admin",
    "http://localhost:8080/users/1happyturtle"
  ],
  "type": "Collection"
}`, dst.String())
}

func TestDereferenceTestSuite(t *testing.T) {
	suite.Run(t, new(DereferenceTestSuite))
}
