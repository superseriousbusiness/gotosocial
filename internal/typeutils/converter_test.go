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
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

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
		"id": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552/activity",
		"type": "Create",
		"actor": "http://fossbros-anonymous.io/users/foss_satan",
		"published": "2021-05-12T09:58:38.00Z",
		"to": [
		  "http://fossbros-anonymous.io/users/foss_satan/followers"
		],
		"cc": [
		  "https://www.w3.org/ns/activitystreams#Public",
		  "http://localhost:8080/users/the_mighty_zork"
		],
		"object": {
		  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552",
		  "type": "Note",
		  "summary": null,
		  "inReplyTo": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
		  "published": "2021-05-12T09:58:38.00Z",
		  "url": "https://ondergrond.org/@dumpsterqueer/106221634728637552",
		  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
		  "to": [
			"http://fossbros-anonymous.io/users/foss_satan/followers"
		  ],
		  "cc": [
			"https://www.w3.org/ns/activitystreams#Public",
			"http://localhost:8080/users/the_mighty_zork"
		  ],
		  "sensitive": false,
		  "conversation": "tag:ondergrond.org,2021-05-12:objectId=1132361:objectType=Conversation",
		  "content": "<p><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span> nice there it is:</p><p><a href=\"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"ellipsis\">social.pixie.town/users/f0x/st</span><span class=\"invisible\">atuses/106221628567855262/activity</span></a></p>",
		  "contentMap": {
			"en": "<p><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span> nice there it is:</p><p><a href=\"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"ellipsis\">social.pixie.town/users/f0x/st</span><span class=\"invisible\">atuses/106221628567855262/activity</span></a></p>"
		  },
		  "attachment": [],
		  "tag": [
			{
			  "type": "Mention",
			  "href": "http://localhost:8080/users/the_mighty_zork",
			  "name": "@the_mighty_zork@localhost:8080"
			}
		  ],
		  "replies": {
			"id": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552/replies",
			"type": "Collection",
			"first": {
			  "type": "CollectionPage",
			  "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552/replies?only_other_accounts=true&page=true",
			  "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552/replies",
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
		"id": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221567884565704/activity",
		"type": "Create",
		"actor": "http://fossbros-anonymous.io/users/foss_satan",
		"published": "2021-05-12T09:41:38.00Z",
		"to": [
		  "http://fossbros-anonymous.io/users/foss_satan/followers"
		],
		"cc": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"object": {
		  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221567884565704",
		  "type": "Note",
		  "summary": null,
		  "inReplyTo": null,
		  "published": "2021-05-12T09:41:38.00Z",
		  "url": "https://ondergrond.org/@dumpsterqueer/106221567884565704",
		  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
		  "to": [
			"http://fossbros-anonymous.io/users/foss_satan/followers"
		  ],
		  "cc": [
			"https://www.w3.org/ns/activitystreams#Public"
		  ],
		  "sensitive": false,
		  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221567884565704",
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
			  "updated": "2020-11-06T13:42:11.00Z",
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
			  "updated": "2020-09-26T12:29:56.00Z",
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
			  "updated": "2020-09-26T12:13:23.00Z",
			  "icon": {
				"type": "Image",
				"mediaType": "image/png",
				"url": "https://ondergrond.org/system/custom_emojis/images/000/000/764/original/3f8eef9de773c90d.png"
			  }
			}
		  ],
		  "replies": {
			"id": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221567884565704/replies",
			"type": "Collection",
			"first": {
			  "type": "CollectionPage",
			  "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221567884565704/replies?only_other_accounts=true&page=true",
			  "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/106221567884565704/replies",
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
	publicStatusActivityJson = `
	{
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
		"id": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167",
		"type": "Note",
		"summary": "reading: Punishment and Reward in the Corporate University",
		"inReplyTo": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138729399508469",
		"published": "2022-04-15T23:49:37.00Z",
		"url": "http://fossbros-anonymous.io/@foss_satan/108138763199405167",
		"attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
		"to": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"cc": [
		  "http://fossbros-anonymous.io/users/foss_satan/followers"
		],
		"sensitive": true,
		"atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167",
		"inReplyToAtomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138729399508469",
		"content": "<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>",
		"contentMap": {
		  "en": "<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>"
		},
		"attachment": [],
		"tag": [],
		"replies": {
		  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167/replies",
		  "type": "Collection",
		  "first": {
			"type": "CollectionPage",
			"next": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167/replies?only_other_accounts=true&page=true",
			"partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167/replies",
			"items": []
		  }
		}
	  }	  
	`
	publicStatusActivityJsonNoURL = `
	{
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
		"id": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167",
		"type": "Note",
		"summary": "reading: Punishment and Reward in the Corporate University",
		"inReplyTo": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138729399508469",
		"published": "2022-04-15T23:49:37.00Z",
		"attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
		"to": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"cc": [
		  "http://fossbros-anonymous.io/users/foss_satan/followers"
		],
		"sensitive": true,
		"atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167",
		"inReplyToAtomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138729399508469",
		"content": "<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>",
		"contentMap": {
		  "en": "<p>&gt; So we have to examine critical thinking as a signifier, dynamic and ambiguous.  It has a normative definition, a tacit definition, and an ideal definition.  One of the hallmarks of graduate training is learning to comprehend those definitions and applying the correct one as needed for professional success.</p>"
		},
		"attachment": [],
		"tag": [],
		"replies": {
		  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167/replies",
		  "type": "Collection",
		  "first": {
			"type": "CollectionPage",
			"next": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167/replies?only_other_accounts=true&page=true",
			"partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/108138763199405167/replies",
			"items": []
		  }
		}
	  }	  
	`
	owncastService = `
	{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  "http://joinmastodon.org/ns",
		  "https://w3id.org/security/v1"
		],
		"attachment": {
		  "name": "Stream",
		  "type": "PropertyValue",
		  "value": "<a href=\"https://owncast.example.org\" rel=\"me nofollow noopener noreferrer\" target=\"_blank\">https://owncast.example.org</a>"
		},
		"discoverable": true,
		"followers": "https://owncast.example.org/federation/user/rgh/followers",
		"icon": {
		  "type": "Image",
		  "url": "https://owncast.example.org/logo/external"
		},
		"id": "https://owncast.example.org/federation/user/rgh",
		"image": {
		  "type": "Image",
		  "url": "https://owncast.example.org/logo/external"
		},
		"inbox": "https://owncast.example.org/federation/user/rgh/inbox",
		"manuallyApprovesFollowers": false,
		"name": "Rob's Owncast Server",
		"outbox": "https://owncast.example.org/federation/user/rgh/outbox",
		"preferredUsername": "rgh",
		"publicKey": {
		  "id": "https://owncast.example.org/federation/user/rgh#main-key",
		  "owner": "https://owncast.example.org/federation/user/rgh",
		  "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAurN+lUNwcGV2poLNtaoT\naRtJzN6s4SDcBmIFk82lxhdMKC6/Nssm+hvDuxWGqL0+dHwSvrG11rA6irGuSzRk\niHjYyVwYe/p1CxqJxzUfZVJAWdsCFWy+HtDrTWs5sggj1MiL59uGxvkCep+OYBuG\nBI8CvSOMLrDp8soCg3EY+zSgpXtGMuRaaUukavsfuglApShB61ny7W8LG252iKC5\nmyO8L7l8TNa5BrIi/pRHLzvv9aWiCa8VKtvmqj+mClEhpkRhImSk5GPJXgouTTgl\ntT28NYYciSf9YYgZ0SNWHdLUCdkMF592j4+BbkPvdgzc70G4yyu2GcWnTzBuvF5X\nYwIDAQAB\n-----END PUBLIC KEY-----\n"
		},
		"published": "2022-05-22T18:44:57.00Z",
		"summary": "linux audio stuff ",
		"tag": [
		  {
			"href": "https://directory.owncast.online/tags/owncast",
			"name": "#owncast",
			"type": "Hashtag"
		  },
		  {
			"href": "https://directory.owncast.online/tags/streaming",
			"name": "#streaming",
			"type": "Hashtag"
		  }
		],
		"type": "Service",
		"url": "https://owncast.example.org/federation/user/rgh"
	} 
`
)

type TypeUtilsTestSuite struct {
	suite.Suite
	db              db.DB
	testAccounts    map[string]*gtsmodel.Account
	testStatuses    map[string]*gtsmodel.Status
	testAttachments map[string]*gtsmodel.MediaAttachment
	testPeople      map[string]vocab.ActivityStreamsPerson
	testEmojis      map[string]*gtsmodel.Emoji
	testReports     map[string]*gtsmodel.Report

	typeconverter typeutils.TypeConverter
}

func (suite *TypeUtilsTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testPeople = testrig.NewTestFediPeople()
	suite.testEmojis = testrig.NewTestEmojis()
	suite.testReports = testrig.NewTestReports()
	suite.typeconverter = typeutils.NewConverter(suite.db)
}

func (suite *TypeUtilsTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *TypeUtilsTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}
