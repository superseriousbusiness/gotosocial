/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"testing"

	"github.com/go-fed/activity/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ASToInternalTestSuite struct {
	ConverterStandardTestSuite
}

const (
	gargronAsActivityJson = `{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  "https://w3id.org/security/v1",
		  {
			"manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
			"toot": "http://joinmastodon.org/ns#",
			"featured": {
			  "@id": "toot:featured",
			  "@type": "@id"
			},
			"featuredTags": {
			  "@id": "toot:featuredTags",
			  "@type": "@id"
			},
			"alsoKnownAs": {
			  "@id": "as:alsoKnownAs",
			  "@type": "@id"
			},
			"movedTo": {
			  "@id": "as:movedTo",
			  "@type": "@id"
			},
			"schema": "http://schema.org#",
			"PropertyValue": "schema:PropertyValue",
			"value": "schema:value",
			"IdentityProof": "toot:IdentityProof",
			"discoverable": "toot:discoverable",
			"Device": "toot:Device",
			"Ed25519Signature": "toot:Ed25519Signature",
			"Ed25519Key": "toot:Ed25519Key",
			"Curve25519Key": "toot:Curve25519Key",
			"EncryptedMessage": "toot:EncryptedMessage",
			"publicKeyBase64": "toot:publicKeyBase64",
			"deviceId": "toot:deviceId",
			"claim": {
			  "@type": "@id",
			  "@id": "toot:claim"
			},
			"fingerprintKey": {
			  "@type": "@id",
			  "@id": "toot:fingerprintKey"
			},
			"identityKey": {
			  "@type": "@id",
			  "@id": "toot:identityKey"
			},
			"devices": {
			  "@type": "@id",
			  "@id": "toot:devices"
			},
			"messageFranking": "toot:messageFranking",
			"messageType": "toot:messageType",
			"cipherText": "toot:cipherText",
			"suspended": "toot:suspended",
			"focalPoint": {
			  "@container": "@list",
			  "@id": "toot:focalPoint"
			}
		  }
		],
		"id": "https://mastodon.social/users/Gargron",
		"type": "Person",
		"following": "https://mastodon.social/users/Gargron/following",
		"followers": "https://mastodon.social/users/Gargron/followers",
		"inbox": "https://mastodon.social/users/Gargron/inbox",
		"outbox": "https://mastodon.social/users/Gargron/outbox",
		"featured": "https://mastodon.social/users/Gargron/collections/featured",
		"featuredTags": "https://mastodon.social/users/Gargron/collections/tags",
		"preferredUsername": "Gargron",
		"name": "Eugen",
		"summary": "<p>Developer of Mastodon and administrator of mastodon.social. I post service announcements, development updates, and personal stuff.</p>",
		"url": "https://mastodon.social/@Gargron",
		"manuallyApprovesFollowers": false,
		"discoverable": true,
		"devices": "https://mastodon.social/users/Gargron/collections/devices",
		"alsoKnownAs": [
		  "https://tooting.ai/users/Gargron"
		],
		"publicKey": {
		  "id": "https://mastodon.social/users/Gargron#main-key",
		  "owner": "https://mastodon.social/users/Gargron",
		  "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvXc4vkECU2/CeuSo1wtn\nFoim94Ne1jBMYxTZ9wm2YTdJq1oiZKif06I2fOqDzY/4q/S9uccrE9Bkajv1dnkO\nVm31QjWlhVpSKynVxEWjVBO5Ienue8gND0xvHIuXf87o61poqjEoepvsQFElA5ym\novljWGSA/jpj7ozygUZhCXtaS2W5AD5tnBQUpcO0lhItYPYTjnmzcc4y2NbJV8hz\n2s2G8qKv8fyimE23gY1XrPJg+cRF+g4PqFXujjlJ7MihD9oqtLGxbu7o1cifTn3x\nBfIdPythWu5b4cujNsB3m3awJjVmx+MHQ9SugkSIYXV0Ina77cTNS0M2PYiH1PFR\nTwIDAQAB\n-----END PUBLIC KEY-----\n"
		},
		"tag": [],
		"attachment": [
		  {
			"type": "PropertyValue",
			"name": "Patreon",
			"value": "<a href=\"https://www.patreon.com/mastodon\" rel=\"me nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://www.</span><span class=\"\">patreon.com/mastodon</span><span class=\"invisible\"></span></a>"
		  },
		  {
			"type": "PropertyValue",
			"name": "Homepage",
			"value": "<a href=\"https://zeonfederated.com\" rel=\"me nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"\">zeonfederated.com</span><span class=\"invisible\"></span></a>"
		  },
		  {
			"type": "IdentityProof",
			"name": "gargron",
			"signatureAlgorithm": "keybase",
			"signatureValue": "5cfc20c7018f2beefb42a68836da59a792e55daa4d118498c9b1898de7e845690f"
		  }
		],
		"endpoints": {
		  "sharedInbox": "https://mastodon.social/inbox"
		},
		"icon": {
		  "type": "Image",
		  "mediaType": "image/jpeg",
		  "url": "https://files.mastodon.social/accounts/avatars/000/000/001/original/d96d39a0abb45b92.jpg"
		},
		"image": {
		  "type": "Image",
		  "mediaType": "image/png",
		  "url": "https://files.mastodon.social/accounts/headers/000/000/001/original/c91b871f294ea63e.png"
		}
	  }`
)

func (suite *ASToInternalTestSuite) SetupSuite() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.accounts = testrig.NewTestAccounts()
	suite.people = testrig.NewTestFediPeople()
	suite.typeconverter = typeutils.NewConverter(suite.config, suite.db)
}

func (suite *ASToInternalTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
}

func (suite *ASToInternalTestSuite) TestParsePerson() {

	testPerson := suite.people["new_person_1"]

	acct, err := suite.typeconverter.ASRepresentationToAccount(testPerson)
	assert.NoError(suite.T(), err)

	fmt.Printf("%+v", acct)
	// TODO: write assertions here, rn we're just eyeballing the output
}

func (suite *ASToInternalTestSuite) TestParseGargron() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(gargronAsActivityJson), &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	rep, ok := t.(typeutils.Accountable)
	assert.True(suite.T(), ok)

	acct, err := suite.typeconverter.ASRepresentationToAccount(rep)
	assert.NoError(suite.T(), err)

	fmt.Printf("%+v", acct)
	// TODO: write assertions here, rn we're just eyeballing the output
}

func (suite *ASToInternalTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func TestASToInternalTestSuite(t *testing.T) {
	suite.Run(t, new(ASToInternalTestSuite))
}
