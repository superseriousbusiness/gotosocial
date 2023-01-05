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

package ap_test

import (
	"context"
	"encoding/json"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
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
	note.SetActivityStreamsContent(content)

	return note
}

func addressable1() ap.Addressable {
	// make a note addressed to public with followers in cc
	note := streams.NewActivityStreamsNote()

	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(testrig.URLMustParse(pub.PublicActivityPubIRI))

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
	ccProp.AppendIRI(testrig.URLMustParse(pub.PublicActivityPubIRI))

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

type ExtractTestSuite struct {
	suite.Suite
	document1         vocab.ActivityStreamsDocument
	attachment1       vocab.ActivityStreamsAttachmentProperty
	noteWithMentions1 vocab.ActivityStreamsNote
	addressable1      ap.Addressable
	addressable2      ap.Addressable
	addressable3      ap.Addressable
	addressable4      vocab.ActivityStreamsAnnounce
	addressable5      ap.Addressable
}

func (suite *ExtractTestSuite) SetupTest() {
	suite.document1 = document1()
	suite.attachment1 = attachment1()
	suite.noteWithMentions1 = noteWithMentions1()
	suite.addressable1 = addressable1()
	suite.addressable2 = addressable2()
	suite.addressable3 = addressable3()
	suite.addressable4 = addressable4()
	suite.addressable5 = addressable5()
}
