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
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ASToInternalTestSuite struct {
	ConverterStandardTestSuite
}

const (
	statusWithMentionsActivityJson = `{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  {
			"ostatus": "http://ostatus.org#",
			"atomUri": "ostatus:atomUri",
			"inReplyToAtomUri": "ostatus:inReplyToAtomUri",
			"conversation": "ostatus:conversation",
			"sensitive": "as:sensitive",
			"toot": "http://joinmastodon.org/ns#",
			"votersCount": "toot:votersCount"
		  }
		],
		"id": "https://ondergrond.org/users/dumpsterqueer/statuses/106221634728637552/activity",
		"type": "Create",
		"actor": "https://ondergrond.org/users/dumpsterqueer",
		"published": "2021-05-12T09:58:38Z",
		"to": [
		  "https://ondergrond.org/users/dumpsterqueer/followers"
		],
		"cc": [
		  "https://www.w3.org/ns/activitystreams#Public",
		  "https://social.pixie.town/users/f0x"
		],
		"object": {
		  "id": "https://ondergrond.org/users/dumpsterqueer/statuses/106221634728637552",
		  "type": "Note",
		  "summary": null,
		  "inReplyTo": "https://social.pixie.town/users/f0x/statuses/106221628567855262",
		  "published": "2021-05-12T09:58:38Z",
		  "url": "https://ondergrond.org/@dumpsterqueer/106221634728637552",
		  "attributedTo": "https://ondergrond.org/users/dumpsterqueer",
		  "to": [
			"https://ondergrond.org/users/dumpsterqueer/followers"
		  ],
		  "cc": [
			"https://www.w3.org/ns/activitystreams#Public",
			"https://social.pixie.town/users/f0x"
		  ],
		  "sensitive": false,
		  "atomUri": "https://ondergrond.org/users/dumpsterqueer/statuses/106221634728637552",
		  "inReplyToAtomUri": "https://social.pixie.town/users/f0x/statuses/106221628567855262",
		  "conversation": "tag:ondergrond.org,2021-05-12:objectId=1132361:objectType=Conversation",
		  "content": "<p><span class=\"h-card\"><a href=\"https://social.pixie.town/@f0x\" class=\"u-url mention\">@<span>f0x</span></a></span> nice there it is:</p><p><a href=\"https://social.pixie.town/users/f0x/statuses/106221628567855262/activity\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"ellipsis\">social.pixie.town/users/f0x/st</span><span class=\"invisible\">atuses/106221628567855262/activity</span></a></p>",
		  "contentMap": {
			"en": "<p><span class=\"h-card\"><a href=\"https://social.pixie.town/@f0x\" class=\"u-url mention\">@<span>f0x</span></a></span> nice there it is:</p><p><a href=\"https://social.pixie.town/users/f0x/statuses/106221628567855262/activity\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"ellipsis\">social.pixie.town/users/f0x/st</span><span class=\"invisible\">atuses/106221628567855262/activity</span></a></p>"
		  },
		  "attachment": [],
		  "tag": [
			{
			  "type": "Mention",
			  "href": "https://social.pixie.town/users/f0x",
			  "name": "@f0x@pixie.town"
			}
		  ],
		  "replies": {
			"id": "https://ondergrond.org/users/dumpsterqueer/statuses/106221634728637552/replies",
			"type": "Collection",
			"first": {
			  "type": "CollectionPage",
			  "next": "https://ondergrond.org/users/dumpsterqueer/statuses/106221634728637552/replies?only_other_accounts=true&page=true",
			  "partOf": "https://ondergrond.org/users/dumpsterqueer/statuses/106221634728637552/replies",
			  "items": []
			}
		  }
		}
	  }`
	statusWithEmojisAndTagsAsActivityJson = `{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  {
			"ostatus": "http://ostatus.org#",
			"atomUri": "ostatus:atomUri",
			"inReplyToAtomUri": "ostatus:inReplyToAtomUri",
			"conversation": "ostatus:conversation",
			"sensitive": "as:sensitive",
			"toot": "http://joinmastodon.org/ns#",
			"votersCount": "toot:votersCount",
			"Hashtag": "as:Hashtag",
			"Emoji": "toot:Emoji",
			"focalPoint": {
			  "@container": "@list",
			  "@id": "toot:focalPoint"
			}
		  }
		],
		"id": "https://ondergrond.org/users/dumpsterqueer/statuses/106221567884565704/activity",
		"type": "Create",
		"actor": "https://ondergrond.org/users/dumpsterqueer",
		"published": "2021-05-12T09:41:38Z",
		"to": [
		  "https://ondergrond.org/users/dumpsterqueer/followers"
		],
		"cc": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"object": {
		  "id": "https://ondergrond.org/users/dumpsterqueer/statuses/106221567884565704",
		  "type": "Note",
		  "summary": null,
		  "inReplyTo": null,
		  "published": "2021-05-12T09:41:38Z",
		  "url": "https://ondergrond.org/@dumpsterqueer/106221567884565704",
		  "attributedTo": "https://ondergrond.org/users/dumpsterqueer",
		  "to": [
			"https://ondergrond.org/users/dumpsterqueer/followers"
		  ],
		  "cc": [
			"https://www.w3.org/ns/activitystreams#Public"
		  ],
		  "sensitive": false,
		  "atomUri": "https://ondergrond.org/users/dumpsterqueer/statuses/106221567884565704",
		  "inReplyToAtomUri": null,
		  "conversation": "tag:ondergrond.org,2021-05-12:objectId=1132361:objectType=Conversation",
		  "content": "<p>just testing activitypub representations of <a href=\"https://ondergrond.org/tags/tags\" class=\"mention hashtag\" rel=\"tag\">#<span>tags</span></a> and <a href=\"https://ondergrond.org/tags/emoji\" class=\"mention hashtag\" rel=\"tag\">#<span>emoji</span></a>  :party_parrot: :amaze: :blobsunglasses: </p><p>don&apos;t mind me....</p>",
		  "contentMap": {
			"en": "<p>just testing activitypub representations of <a href=\"https://ondergrond.org/tags/tags\" class=\"mention hashtag\" rel=\"tag\">#<span>tags</span></a> and <a href=\"https://ondergrond.org/tags/emoji\" class=\"mention hashtag\" rel=\"tag\">#<span>emoji</span></a>  :party_parrot: :amaze: :blobsunglasses: </p><p>don&apos;t mind me....</p>"
		  },
		  "attachment": [],
		  "tag": [
			{
			  "type": "Hashtag",
			  "href": "https://ondergrond.org/tags/tags",
			  "name": "#tags"
			},
			{
			  "type": "Hashtag",
			  "href": "https://ondergrond.org/tags/emoji",
			  "name": "#emoji"
			},
			{
			  "id": "https://ondergrond.org/emojis/2390",
			  "type": "Emoji",
			  "name": ":party_parrot:",
			  "updated": "2020-11-06T13:42:11Z",
			  "icon": {
				"type": "Image",
				"mediaType": "image/gif",
				"url": "https://ondergrond.org/system/custom_emojis/images/000/002/390/original/ef133aac7ab23341.gif"
			  }
			},
			{
			  "id": "https://ondergrond.org/emojis/2395",
			  "type": "Emoji",
			  "name": ":amaze:",
			  "updated": "2020-09-26T12:29:56Z",
			  "icon": {
				"type": "Image",
				"mediaType": "image/png",
				"url": "https://ondergrond.org/system/custom_emojis/images/000/002/395/original/2c7d9345e57367ed.png"
			  }
			},
			{
			  "id": "https://ondergrond.org/emojis/764",
			  "type": "Emoji",
			  "name": ":blobsunglasses:",
			  "updated": "2020-09-26T12:13:23Z",
			  "icon": {
				"type": "Image",
				"mediaType": "image/png",
				"url": "https://ondergrond.org/system/custom_emojis/images/000/000/764/original/3f8eef9de773c90d.png"
			  }
			}
		  ],
		  "replies": {
			"id": "https://ondergrond.org/users/dumpsterqueer/statuses/106221567884565704/replies",
			"type": "Collection",
			"first": {
			  "type": "CollectionPage",
			  "next": "https://ondergrond.org/users/dumpsterqueer/statuses/106221567884565704/replies?only_other_accounts=true&page=true",
			  "partOf": "https://ondergrond.org/users/dumpsterqueer/statuses/106221567884565704/replies",
			  "items": []
			}
		  }
		}
	  }`
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
	suite.typeconverter = typeutils.NewConverter(suite.config, suite.db, suite.log)
}

func (suite *ASToInternalTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *ASToInternalTestSuite) TestParsePerson() {
	testPerson := suite.people["https://unknown-instance.com/users/brand_new_person"]

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), testPerson, false)
	assert.NoError(suite.T(), err)

	suite.Equal("https://unknown-instance.com/users/brand_new_person", acct.URI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/following", acct.FollowingURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/followers", acct.FollowersURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/inbox", acct.InboxURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/outbox", acct.OutboxURI)
	suite.Equal("https://unknown-instance.com/users/brand_new_person/collections/featured", acct.FeaturedCollectionURI)
	suite.Equal("brand_new_person", acct.Username)
	suite.Equal("Geoff Brando New Personson", acct.DisplayName)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", acct.Note)
	suite.Equal("https://unknown-instance.com/@brand_new_person", acct.URL)
	suite.True(acct.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person#main-key", acct.PublicKeyURI)
	suite.False(acct.Locked)
}

func (suite *ASToInternalTestSuite) TestParseGargron() {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(gargronAsActivityJson), &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	rep, ok := t.(ap.Accountable)
	assert.True(suite.T(), ok)

	acct, err := suite.typeconverter.ASRepresentationToAccount(context.Background(), rep, false)
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
