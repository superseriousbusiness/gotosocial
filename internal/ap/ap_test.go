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
	"bytes"
	"context"
	"encoding/json"
	"io"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func document1() vocab.ActivityStreamsDocument {
	d := streams.NewActivityStreamsDocument()

	dMediaType := streams.NewActivityStreamsMediaTypeProperty()
	dMediaType.Set("image/jpeg")
	d.SetActivityStreamsMediaType(dMediaType)

	dURL := streams.NewActivityStreamsUrlProperty()
	dURL.AppendIRI(testrig.URLMustParse("https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg"))
	d.SetActivityStreamsUrl(dURL)

	dName := streams.NewActivityStreamsNameProperty()
	dName.AppendXMLSchemaString("It's a cute plushie.")
	d.SetActivityStreamsName(dName)

	dBlurhash := streams.NewTootBlurhashProperty()
	dBlurhash.Set("UxQ0EkRP_4tRxtRjWBt7%hozM_ayV@oLf6WB")
	d.SetTootBlurhash(dBlurhash)

	dSensitive := streams.NewActivityStreamsSensitiveProperty()
	dSensitive.AppendXMLSchemaBoolean(true)
	d.SetActivityStreamsSensitive(dSensitive)

	return d
}

func attachment1() vocab.ActivityStreamsAttachmentProperty {
	a := streams.NewActivityStreamsAttachmentProperty()
	a.AppendActivityStreamsDocument(document1())
	return a
}

func noteWithMentions1() vocab.ActivityStreamsNote {
	note := streams.NewActivityStreamsNote()

	tags := streams.NewActivityStreamsTagProperty()

	mention1 := streams.NewActivityStreamsMention()

	mention1Href := streams.NewActivityStreamsHrefProperty()
	mention1Href.Set(testrig.URLMustParse("https://gts.superseriousbusiness.org/users/dumpsterqueer"))
	mention1.SetActivityStreamsHref(mention1Href)

	mention1Name := streams.NewActivityStreamsNameProperty()
	mention1Name.AppendXMLSchemaString("@dumpsterqueer@superseriousbusiness.org")
	mention1.SetActivityStreamsName(mention1Name)

	mention2 := streams.NewActivityStreamsMention()

	mention2Href := streams.NewActivityStreamsHrefProperty()
	mention2Href.Set(testrig.URLMustParse("https://gts.superseriousbusiness.org/users/f0x"))
	mention2.SetActivityStreamsHref(mention2Href)

	mention2Name := streams.NewActivityStreamsNameProperty()
	mention2Name.AppendXMLSchemaString("@f0x@superseriousbusiness.org")
	mention2.SetActivityStreamsName(mention2Name)

	tags.AppendActivityStreamsMention(mention1)
	tags.AppendActivityStreamsMention(mention2)

	note.SetActivityStreamsTag(tags)

	content := streams.NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("hey @f0x and @dumpsterqueer")

	rdfLangString := make(map[string]string)
	rdfLangString["en"] = "hey @f0x and @dumpsterqueer"
	rdfLangString["fr"] = "bonjour @f0x et @dumpsterqueer"
	content.AppendRDFLangString(rdfLangString)

	note.SetActivityStreamsContent(content)

	policy := streams.NewGoToSocialInteractionPolicy()

	// Set canLike.
	canLike := streams.NewGoToSocialCanLike()

	// Anyone can like.
	canLikeAlwaysProp := streams.NewGoToSocialAlwaysProperty()
	canLikeAlwaysProp.AppendIRI(ap.PublicURI())
	canLike.SetGoToSocialAlways(canLikeAlwaysProp)

	// Empty approvalRequired.
	canLikeApprovalRequiredProp := streams.NewGoToSocialApprovalRequiredProperty()
	canLike.SetGoToSocialApprovalRequired(canLikeApprovalRequiredProp)

	// Set canLike on the policy.
	canLikeProp := streams.NewGoToSocialCanLikeProperty()
	canLikeProp.AppendGoToSocialCanLike(canLike)
	policy.SetGoToSocialCanLike(canLikeProp)

	// Build canReply.
	canReply := streams.NewGoToSocialCanReply()

	// Anyone can reply.
	canReplyAlwaysProp := streams.NewGoToSocialAlwaysProperty()
	canReplyAlwaysProp.AppendIRI(ap.PublicURI())
	canReply.SetGoToSocialAlways(canReplyAlwaysProp)

	// Set empty approvalRequired.
	canReplyApprovalRequiredProp := streams.NewGoToSocialApprovalRequiredProperty()
	canReply.SetGoToSocialApprovalRequired(canReplyApprovalRequiredProp)

	// Set canReply on the policy.
	canReplyProp := streams.NewGoToSocialCanReplyProperty()
	canReplyProp.AppendGoToSocialCanReply(canReply)
	policy.SetGoToSocialCanReply(canReplyProp)

	// Build canAnnounce.
	canAnnounce := streams.NewGoToSocialCanAnnounce()

	// Only f0x and dumpsterqueer can announce.
	canAnnounceAlwaysProp := streams.NewGoToSocialAlwaysProperty()
	canAnnounceAlwaysProp.AppendIRI(testrig.URLMustParse("https://gts.superseriousbusiness.org/users/dumpsterqueer"))
	canAnnounceAlwaysProp.AppendIRI(testrig.URLMustParse("https://gts.superseriousbusiness.org/users/f0x"))
	canAnnounce.SetGoToSocialAlways(canAnnounceAlwaysProp)

	// Public requires approval to announce.
	canAnnounceApprovalRequiredProp := streams.NewGoToSocialApprovalRequiredProperty()
	canAnnounceApprovalRequiredProp.AppendIRI(ap.PublicURI())
	canAnnounce.SetGoToSocialApprovalRequired(canAnnounceApprovalRequiredProp)

	// Set canAnnounce on the policy.
	canAnnounceProp := streams.NewGoToSocialCanAnnounceProperty()
	canAnnounceProp.AppendGoToSocialCanAnnounce(canAnnounce)
	policy.SetGoToSocialCanAnnounce(canAnnounceProp)

	// Set the policy on the note.
	policyProp := streams.NewGoToSocialInteractionPolicyProperty()
	policyProp.AppendGoToSocialInteractionPolicy(policy)
	note.SetGoToSocialInteractionPolicy(policyProp)

	return note
}

func (suite *APTestSuite) noteWithHashtags1() ap.Statusable {
	noteJson := []byte(`
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
		"votersCount": "toot:votersCount",
		"Hashtag": "as:Hashtag"
	  }
	],
	"id": "https://mastodon.social/users/pixelfed/statuses/110609702372389319",
	"type": "Note",
	"summary": null,
	"inReplyTo": null,
	"published": "2023-06-26T09:01:56Z",
	"url": "https://mastodon.social/@pixelfed/110609702372389319",
	"attributedTo": "https://mastodon.social/users/pixelfed",
	"to": [
	  "https://www.w3.org/ns/activitystreams#Public"
	],
	"cc": [
	  "https://mastodon.social/users/pixelfed/followers",
	  "https://gts.superseriousbusiness.org/users/gotosocial"
	],
	"sensitive": false,
	"atomUri": "https://mastodon.social/users/pixelfed/statuses/110609702372389319",
	"inReplyToAtomUri": null,
	"conversation": "tag:mastodon.social,2023-06-26:objectId=474977189:objectType=Conversation",
	"content": "<p>⚡ Heard of <span class=\"h-card\" translate=\"no\"><a href=\"https://gts.superseriousbusiness.org/@gotosocial\" class=\"u-url mention\">@<span>gotosocial</span></a></span> ?</p><p>GoToSocial provides a lightweight, customizable, and safety-focused entryway into the <a href=\"https://mastodon.social/tags/fediverse\" class=\"mention hashtag\" rel=\"tag\">#<span>fediverse</span></a>, you can keep in touch with your friends, post, read, and share images and articles.</p><p>Consider <a href=\"https://mastodon.social/tags/GoToSocial\" class=\"mention hashtag\" rel=\"tag\">#<span>GoToSocial</span></a> instead of Pixelfed if you&#39;d like a safety-focused alternative with text-only post support that is maintained by a stellar developer community!</p><p>We ❤️ GtS, check them out!</p><p>🌍 <a href=\"https://gotosocial.org/\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">gotosocial.org/</span><span class=\"invisible\"></span></a></p><p>🔍 <a href=\"https://fedidb.org/software/gotosocial\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">fedidb.org/software/gotosocial</span><span class=\"invisible\"></span></a></p>",
	"contentMap": {
	  "en": "<p>⚡ Heard of <span class=\"h-card\" translate=\"no\"><a href=\"https://gts.superseriousbusiness.org/@gotosocial\" class=\"u-url mention\">@<span>gotosocial</span></a></span> ?</p><p>GoToSocial provides a lightweight, customizable, and safety-focused entryway into the <a href=\"https://mastodon.social/tags/fediverse\" class=\"mention hashtag\" rel=\"tag\">#<span>fediverse</span></a>, you can keep in touch with your friends, post, read, and share images and articles.</p><p>Consider <a href=\"https://mastodon.social/tags/GoToSocial\" class=\"mention hashtag\" rel=\"tag\">#<span>GoToSocial</span></a> instead of Pixelfed if you&#39;d like a safety-focused alternative with text-only post support that is maintained by a stellar developer community!</p><p>We ❤️ GtS, check them out!</p><p>🌍 <a href=\"https://gotosocial.org/\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">gotosocial.org/</span><span class=\"invisible\"></span></a></p><p>🔍 <a href=\"https://fedidb.org/software/gotosocial\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">fedidb.org/software/gotosocial</span><span class=\"invisible\"></span></a></p>"
	},
	"attachment": [],
	"tag": [
	  {
		"type": "Mention",
		"href": "https://gts.superseriousbusiness.org/users/gotosocial",
		"name": "@gotosocial@superseriousbusiness.org"
	  },
	  {
		"type": "Hashtag",
		"href": "https://mastodon.social/tags/fediverse",
		"name": "#fediverse"
	  },
	  {
		"type": "Hashtag",
		"href": "https://mastodon.social/tags/gotosocial",
		"name": "#gotosocial"
	  },
	  {
		"type": "Hashtag",
		"href": "https://mastodon.social/tags/this_hashtag_will_be_ignored_since_it_cant_be_normalized",
		"name": "#b̴̛͇̒̌͑̓̐̑͗̏̐̇͗̎̕͝O̵̧̧͎̟̰̭̊͌͒́̊̑̄̐͐͗Ọ̷̧̡̰̟̪̫̹͖͇̱͕̺̦̲̀̐̽̓̇̚͠b̶̨̖͍͙͈̹͉̯͕̯̯̯̞̼̞̏͊͂̐̔͛s̴̢̞̺͈͇̘͚͉͔̥̔͛͆͑͑̍̄̌̚͜͜ͅ"
	  },
	  {
		"type": "Hashtag",
		"href": "https://mastodon.social/tags/this_hashtag_will_be_included_correctly",
		"name": "#Grüvy"
	  },
	  {
		"type": "Hashtag",
		"href": "https://mastodon.social/tags/this_hashtag_will_be_squashed_into_a_single_character",
		"name": "#` + `ᄀ` + `ᅡ` + `ᆨ` + `"
	  }
	],
	"replies": {
	  "id": "https://mastodon.social/users/pixelfed/statuses/110609702372389319/replies",
	  "type": "Collection",
	  "first": {
		"type": "CollectionPage",
		"next": "https://mastodon.social/users/pixelfed/statuses/110609702372389319/replies?only_other_accounts=true&page=true",
		"partOf": "https://mastodon.social/users/pixelfed/statuses/110609702372389319/replies",
		"items": []
	  }
	}
}`)

	statusable, err := ap.ResolveStatusable(
		context.Background(),
		io.NopCloser(bytes.NewReader(noteJson)),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return statusable
}

func addressable1() ap.Addressable {
	// make a note addressed to public with followers in cc
	note := streams.NewActivityStreamsNote()

	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(ap.PublicURI())

	note.SetActivityStreamsTo(toProp)

	ccProp := streams.NewActivityStreamsCcProperty()
	ccProp.AppendIRI(testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork/followers"))

	note.SetActivityStreamsCc(ccProp)

	return note
}

func addressable2() ap.Addressable {
	// make a note addressed to followers with public in cc
	note := streams.NewActivityStreamsNote()

	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork/followers"))

	note.SetActivityStreamsTo(toProp)

	ccProp := streams.NewActivityStreamsCcProperty()
	ccProp.AppendIRI(ap.PublicURI())

	note.SetActivityStreamsCc(ccProp)

	return note
}

func addressable3() ap.Addressable {
	// make a note addressed to followers
	note := streams.NewActivityStreamsNote()

	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork/followers"))

	note.SetActivityStreamsTo(toProp)

	return note
}

func addressable4() vocab.ActivityStreamsAnnounce {
	// https://github.com/superseriousbusiness/gotosocial/issues/267
	announceJson := []byte(`
{
	"@context": "https://www.w3.org/ns/activitystreams",
	"actor": "https://example.org/users/someone",
	"cc": "https://another.instance/users/someone_else",
	"id": "https://example.org/users/someone/statuses/107043888547829808/activity",
	"object": "https://another.instance/users/someone_else/statuses/107026674805188668",
	"published": "2021-10-04T15:08:35.00Z",
	"to": "https://example.org/users/someone/followers",
	"type": "Announce"
}`)

	var jsonAsMap map[string]interface{}
	err := json.Unmarshal(announceJson, &jsonAsMap)
	if err != nil {
		panic(err)
	}

	t, err := streams.ToType(context.Background(), jsonAsMap)
	if err != nil {
		panic(err)
	}

	return t.(vocab.ActivityStreamsAnnounce)
}

func addressable5() ap.Addressable {
	// make a note addressed to one person (direct message)
	note := streams.NewActivityStreamsNote()

	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(testrig.URLMustParse("http://localhost:8080/users/1_happy_turtle"))

	note.SetActivityStreamsTo(toProp)

	return note
}

type APTestSuite struct {
	suite.Suite
	document1         vocab.ActivityStreamsDocument
	attachment1       vocab.ActivityStreamsAttachmentProperty
	noteWithMentions1 vocab.ActivityStreamsNote
	addressable1      ap.Addressable
	addressable2      ap.Addressable
	addressable3      ap.Addressable
	addressable4      vocab.ActivityStreamsAnnounce
	addressable5      ap.Addressable
	testAccounts      map[string]*gtsmodel.Account
}

func (suite *APTestSuite) jsonToType(rawJson string) (vocab.Type, map[string]interface{}) {
	var raw map[string]interface{}
	err := json.Unmarshal([]byte(rawJson), &raw)
	if err != nil {
		panic(err)
	}

	t, err := streams.ToType(context.Background(), raw)
	if err != nil {
		panic(err)
	}

	return t, raw
}

func (suite *APTestSuite) typeToJson(t vocab.Type) string {
	m, err := ap.Serialize(t)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	return string(b)
}

func (suite *APTestSuite) SetupTest() {
	suite.document1 = document1()
	suite.attachment1 = attachment1()
	suite.noteWithMentions1 = noteWithMentions1()
	suite.addressable1 = addressable1()
	suite.addressable2 = addressable2()
	suite.addressable3 = addressable3()
	suite.addressable4 = addressable4()
	suite.addressable5 = addressable5()
	suite.testAccounts = testrig.NewTestAccounts()
}
