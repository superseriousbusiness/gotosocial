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

package ap_test

import (
	"testing"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type NormalizeTestSuite struct {
	APTestSuite
}

func (suite *NormalizeTestSuite) getStatusable() (vocab.ActivityStreamsNote, map[string]interface{}) {
	t, raw := suite.jsonToType(`{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  "https://example.org/schemas/litepub-0.1.jsonld",
		  {
			"@language": "und"
		  }
		],
		"actor": "https://example.org/users/someone",
		"attachment": [],
		"attributedTo": "https://example.org/users/someone",
		"cc": [
		  "https://example.org/users/someone/followers"
		],
		"content": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class=\"hashtag\" data-tag=\"twittermigration\" href=\"https://example.org/tag/twittermigration\" rel=\"tag ugc\">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night&#39;s spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.",
		"contentMap": {
			"en": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class=\"hashtag\" data-tag=\"twittermigration\" href=\"https://example.org/tag/twittermigration\" rel=\"tag ugc\">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night&#39;s spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues."
		},
		"context": "https://example.org/contexts/01GX0MSHPER1E0FT022Q209EJZ",
		"conversation": "https://example.org/contexts/01GX0MSHPER1E0FT022Q209EJZ",
		"id": "https://example.org/objects/01GX0MT2PA58JNSMK11MCS65YD",
		"published": "2022-11-18T17:43:58.489995Z",
		"replies": {
		  "items": [
			"https://example.org/objects/01GX0MV12MGEG3WF9SWB5K3KRJ"
		  ],
		  "type": "Collection"
		},
		"repliesCount": 0,
		"sensitive": null,
		"source": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the #TwitterMigration.\r\n\r\nIn fact, 100,000 new accounts have been created since last night.\r\n\r\nSince last night's spike 8,000-12,000 new accounts are being created every hour.\r\n\r\nYesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.",
		"summary": "",
		"tag": [
		  {
			"href": "https://example.org/tags/twittermigration",
			"name": "#twittermigration",
			"type": "Hashtag"
		  }
		],
		"to": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"type": "Note"
	  }`)

	return t.(vocab.ActivityStreamsNote), raw
}

func (suite *NormalizeTestSuite) getStatusableWithOneAttachment() (vocab.ActivityStreamsNote, map[string]interface{}) {
	t, raw := suite.jsonToType(`{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
		"type": "Note",
		"url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ",
		"attributedTo": "https://example.org/users/hourlycatbot",
		"to": "https://www.w3.org/ns/activitystreams#Public",
		"attachment": [
		  {
			"type": "Document",
			"mediaType": "image/jpeg",
			"url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg",
			"name": "DESCRIPTION: here's <<a>> picture of a #cat, it's cute! here's some special characters: \"\" \\ weeee''''"
		  }
		]
	  }`)

	return t.(vocab.ActivityStreamsNote), raw
}

func (suite *NormalizeTestSuite) getStatusableWithOneAttachmentEmbedded() (vocab.ActivityStreamsNote, map[string]interface{}) {
	t, raw := suite.jsonToType(`{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
		"type": "Note",
		"url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ",
		"attributedTo": "https://example.org/users/hourlycatbot",
		"to": "https://www.w3.org/ns/activitystreams#Public",
		"attachment": {
		  "type": "Document",
		  "mediaType": "image/jpeg",
		  "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg",
		  "name": "DESCRIPTION: here's <<a>> picture of a #cat, it's cute! here's some special characters: \"\" \\ weeee''''"
		}
	  }`)

	return t.(vocab.ActivityStreamsNote), raw
}

func (suite *NormalizeTestSuite) getStatusableWithMultipleAttachments() (vocab.ActivityStreamsNote, map[string]interface{}) {
	t, raw := suite.jsonToType(`{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
		"type": "Note",
		"url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ",
		"attributedTo": "https://example.org/users/hourlycatbot",
		"to": "https://www.w3.org/ns/activitystreams#Public",
		"attachment": [
		  {
			"type": "Document",
			"mediaType": "image/jpeg",
			"url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg",
			"name": "DESCRIPTION: here's <<a>> picture of a #cat, it's cute! here's some special characters: \"\" \\ weeee''''"
		  },
		  {
			"type": "Document",
			"mediaType": "image/jpeg",
			"url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg",
			"name": "hello: here's another #picture #of #a #cat, hope you like it!!!!!!!"
		  },
		  {
			"type": "Document",
			"mediaType": "image/jpeg",
			"url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
		  },
		  {
			"type": "Document",
			"mediaType": "image/jpeg",
			"url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg",
			"name": "image of a cat &amp; there's a note saying: &lt;danger: #cute but will claw you :(&gt;"
		  }
		]
	  }`)

	return t.(vocab.ActivityStreamsNote), raw
}

func (suite *NormalizeTestSuite) getStatusableWithWeirdSummaryAndName() (vocab.ActivityStreamsNote, map[string]interface{}) {
	t, raw := suite.jsonToType(`{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
		"type": "Note",
		"url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ",
		"attributedTo": "https://example.org/users/hourlycatbot",
		"to": "https://www.w3.org/ns/activitystreams#Public",
		"summary": "warning: #WEIRD #SUMMARY ;;;;a;;a;asv    khop8273987(*^&^)",
		"name": "WARNING: #WEIRD #nameEE ;;;;a;;a;asv    khop8273987(*^&^)"
	  }`)

	return t.(vocab.ActivityStreamsNote), raw
}

func (suite *NormalizeTestSuite) getAccountable() (vocab.ActivityStreamsPerson, map[string]interface{}) {
	t, raw := suite.jsonToType(`{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/someone",
		"summary": "about: I'm a #Barbie #girl in a #Barbie #world\nLife in plastic, it's fantastic\nYou can brush my hair, undress me everywhere\nImagination, life is your creation\nI'm a blonde bimbo girl\nIn a fantasy world\nDress me up, make it tight\nI'm your dolly\nYou're my doll, rock and roll\nFeel the glamour in pink\nKiss me here, touch me there\nHanky panky",
		"attachment": [
			{
				"name": "<strong>cheeky</strong>",
				"type": "PropertyValue",
				"value": "<script>alert(\"teehee!\")</script>"
			},
			{
				"name": "buy me coffee?",
				"type": "PropertyValue",
				"value": "<a href=\"https://example.org/some_link_to_my_ko_fi\">Right here!</a>"
			},
			{
				"name": "hello",
				"type": "PropertyValue",
				"value": "world"
			}
		],
		"type": "Person"
	  }`)

	return t.(vocab.ActivityStreamsPerson), raw
}

func (suite *NormalizeTestSuite) TestNormalizeActivityObject() {
	note, rawNote := suite.getStatusable()
	content := ap.ExtractContent(note)
	suite.Equal(
		`update: As of this morning there are now more than 7 million Mastodon users, most from the <a class="hashtag" data-tag="twittermigration" href="https://example.org/tag/twittermigration" rel="tag ugc">#TwitterMigration%3C/a%3E.%3Cbr%3E%3Cbr%3EIn%20fact,%20100,000%20new%20accounts%20have%20been%20created%20since%20last%20night.%3Cbr%3E%3Cbr%3ESince%20last%20night&%2339;s%20spike%208,000-12,000%20new%20accounts%20are%20being%20created%20every%20hour.%3Cbr%3E%3Cbr%3EYesterday,%20I%20estimated%20that%20Mastodon%20would%20have%208%20million%20users%20by%20the%20end%20of%20the%20week.%20That%20might%20happen%20a%20lot%20sooner%20if%20this%20trend%20continues.`,
		content.Content,
	)

	// Malformed contentMap entry
	// will not be extractable yet.
	suite.Empty(content.ContentMap["en"])

	create := testrig.WrapAPNoteInCreate(
		testrig.URLMustParse("https://example.org/create_something"),
		testrig.URLMustParse("https://example.org/users/someone"),
		testrig.TimeMustParse("2022-11-18T17:43:58.489995Z"),
		note,
	)

	ap.NormalizeIncomingActivity(create, map[string]interface{}{"object": rawNote})
	content = ap.ExtractContent(note)

	suite.Equal(
		`UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class="hashtag" href="https://example.org/tag/twittermigration" rel="tag ugc nofollow noreferrer noopener" target="_blank">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night's spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.`,
		content.Content,
	)

	// Content map entry should now be extractable.
	suite.Equal(
		`UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class="hashtag" href="https://example.org/tag/twittermigration" rel="tag ugc nofollow noreferrer noopener" target="_blank">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night's spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.`,
		content.ContentMap["en"],
	)
}

func (suite *NormalizeTestSuite) TestNormalizeStatusableAttachmentsOneAttachment() {
	note, rawNote := suite.getStatusableWithOneAttachment()

	// Without normalization, the 'name' field of
	// the attachment(s) should be all jacked up.
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [
    {
      "mediaType": "image/jpeg",
      "name": "description: here's \u003c\u003ca\u003e\u003e picture of a #cat,%20it%27s%20cute!%20here%27s%20some%20special%20characters:%20%22%22%20%5C%20weeee%27%27%27%27",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    }
  ],
  "attributedTo": "https://example.org/users/hourlycatbot",
  "id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ"
}`, suite.typeToJson(note))

	// Normalize it!
	ap.NormalizeIncomingAttachments(note, rawNote)

	// After normalization, the 'name' field of the
	// attachment should no longer be all jacked up.
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [
    {
      "mediaType": "image/jpeg",
      "name": "DESCRIPTION: here's \u003c\u003e picture of a #cat, it's cute! here's some special characters: \"\" \\ weeee''''",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    }
  ],
  "attributedTo": "https://example.org/users/hourlycatbot",
  "id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ"
}`, suite.typeToJson(note))
}

func (suite *NormalizeTestSuite) TestNormalizeStatusableAttachmentsOneAttachmentEmbedded() {
	note, rawNote := suite.getStatusableWithOneAttachmentEmbedded()

	// Without normalization, the 'name' field of
	// the attachment(s) should be all jacked up.
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [
    {
      "mediaType": "image/jpeg",
      "name": "description: here's \u003c\u003ca\u003e\u003e picture of a #cat,%20it%27s%20cute!%20here%27s%20some%20special%20characters:%20%22%22%20%5C%20weeee%27%27%27%27",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    }
  ],
  "attributedTo": "https://example.org/users/hourlycatbot",
  "id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ"
}`, suite.typeToJson(note))

	// Normalize it!
	ap.NormalizeIncomingAttachments(note, rawNote)

	// After normalization, the 'name' field of the
	// attachment should no longer be all jacked up.
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [
    {
      "mediaType": "image/jpeg",
      "name": "DESCRIPTION: here's \u003c\u003e picture of a #cat, it's cute! here's some special characters: \"\" \\ weeee''''",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    }
  ],
  "attributedTo": "https://example.org/users/hourlycatbot",
  "id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ"
}`, suite.typeToJson(note))
}

func (suite *NormalizeTestSuite) TestNormalizeStatusableAttachmentsMultipleAttachments() {
	note, rawNote := suite.getStatusableWithMultipleAttachments()

	// Without normalization, the 'name' field of
	// the attachment(s) should be all jacked up.
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [
    {
      "mediaType": "image/jpeg",
      "name": "description: here's \u003c\u003ca\u003e\u003e picture of a #cat,%20it%27s%20cute!%20here%27s%20some%20special%20characters:%20%22%22%20%5C%20weeee%27%27%27%27",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    },
    {
      "mediaType": "image/jpeg",
      "name": "hello: here's another #picture%20%23of%20%23a%20%23cat,%20hope%20you%20like%20it!!!!!!!",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    },
    {
      "mediaType": "image/jpeg",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    },
    {
      "mediaType": "image/jpeg",
      "name": "image of a cat \u0026amp; there's a note saying: \u0026lt;danger: #cute but will claw you :(\u0026gt;",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    }
  ],
  "attributedTo": "https://example.org/users/hourlycatbot",
  "id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ"
}`, suite.typeToJson(note))

	// Normalize it!
	ap.NormalizeIncomingAttachments(note, rawNote)

	// After normalization, the 'name' field of the
	// attachment should no longer be all jacked up.
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [
    {
      "mediaType": "image/jpeg",
      "name": "DESCRIPTION: here's \u003c\u003e picture of a #cat, it's cute! here's some special characters: \"\" \\ weeee''''",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    },
    {
      "mediaType": "image/jpeg",
      "name": "hello: here's another #picture #of #a #cat, hope you like it!!!!!!!",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    },
    {
      "mediaType": "image/jpeg",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    },
    {
      "mediaType": "image/jpeg",
      "name": "image of a cat \u0026 there's a note saying:",
      "type": "Document",
      "url": "https://files.example.org/media_attachments/files/110/258/459/579/509/026/original/b65392ebe0fb04ef.jpeg"
    }
  ],
  "attributedTo": "https://example.org/users/hourlycatbot",
  "id": "https://example.org/users/hourlycatbot/statuses/01GYW48H311PZ78C5G856MGJJJ",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "https://example.org/@hourlycatbot/01GYW48H311PZ78C5G856MGJJJ"
}`, suite.typeToJson(note))
}

func (suite *NormalizeTestSuite) TestNormalizeAccountableSummary() {
	accountable, rawAccount := suite.getAccountable()
	suite.Equal(`about: I'm a #Barbie%20%23girl%20in%20a%20%23Barbie%20%23world%0ALife%20in%20plastic,%20it%27s%20fantastic%0AYou%20can%20brush%20my%20hair,%20undress%20me%20everywhere%0AImagination,%20life%20is%20your%20creation%0AI%27m%20a%20blonde%20bimbo%20girl%0AIn%20a%20fantasy%20world%0ADress%20me%20up,%20make%20it%20tight%0AI%27m%20your%20dolly%0AYou%27re%20my%20doll,%20rock%20and%20roll%0AFeel%20the%20glamour%20in%20pink%0AKiss%20me%20here,%20touch%20me%20there%0AHanky%20panky`, ap.ExtractSummary(accountable))

	ap.NormalizeIncomingSummary(accountable, rawAccount)
	suite.Equal(`about: I'm a #Barbie #girl in a #Barbie #world
Life in plastic, it's fantastic
You can brush my hair, undress me everywhere
Imagination, life is your creation
I'm a blonde bimbo girl
In a fantasy world
Dress me up, make it tight
I'm your dolly
You're my doll, rock and roll
Feel the glamour in pink
Kiss me here, touch me there
Hanky panky`, ap.ExtractSummary(accountable))
}

func (suite *NormalizeTestSuite) TestNormalizeAccountableFields() {
	accountable, rawAccount := suite.getAccountable()
	fields := ap.ExtractFields(accountable)

	// Dodgy field.
	suite.Equal(`<strong>cheeky</strong>`, fields[0].Name)
	suite.Equal(`<script>alert("teehee!")</script>`, fields[0].Value)

	// More or less OK field.
	suite.Equal(`buy me coffee?`, fields[1].Name)
	suite.Equal(`<a href="https://example.org/some_link_to_my_ko_fi">Right here!</a>`, fields[1].Value)

	// Fine field.
	suite.Equal(`hello`, fields[2].Name)
	suite.Equal(`world`, fields[2].Value)

	// Normalize 'em.
	ap.NormalizeIncomingFields(accountable, rawAccount)

	// Dodgy field should be removed.
	fields = ap.ExtractFields(accountable)
	suite.Len(fields, 2)

	// More or less OK field is now very OK.
	suite.Equal(`buy me coffee?`, fields[0].Name)
	suite.Equal(`<a href="https://example.org/some_link_to_my_ko_fi" rel="nofollow noreferrer noopener" target="_blank">Right here!</a>`, fields[0].Value)

	// Fine field continues to be fine.
	suite.Equal(`hello`, fields[1].Name)
	suite.Equal(`world`, fields[1].Value)
}

func (suite *NormalizeTestSuite) TestNormalizeStatusableSummary() {
	statusable, rawAccount := suite.getStatusableWithWeirdSummaryAndName()
	suite.Equal(`warning: #WEIRD%20%23SUMMARY%20;;;;a;;a;asv%20%20%20%20khop8273987(*%5E&%5E)`, ap.ExtractSummary(statusable))

	ap.NormalizeIncomingSummary(statusable, rawAccount)
	suite.Equal(`warning: #WEIRD #SUMMARY ;;;;a;;a;asv khop8273987(*^&^)`, ap.ExtractSummary(statusable))
}

func (suite *NormalizeTestSuite) TestNormalizeStatusableName() {
	statusable, rawAccount := suite.getStatusableWithWeirdSummaryAndName()
	suite.Equal(`warning: #WEIRD%20%23nameEE%20;;;;a;;a;asv%20%20%20%20khop8273987(*%5E&%5E)`, ap.ExtractName(statusable))

	ap.NormalizeIncomingName(statusable, rawAccount)
	suite.Equal(`WARNING: #WEIRD #nameEE ;;;;a;;a;asv    khop8273987(*^&^)`, ap.ExtractName(statusable))
}

func TestNormalizeTestSuite(t *testing.T) {
	suite.Run(t, new(NormalizeTestSuite))
}
