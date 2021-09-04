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

package ap_test

import (
	"testing"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func document1() vocab.ActivityStreamsDocument {
	document1 := streams.NewActivityStreamsDocument()

	document1MediaType := streams.NewActivityStreamsMediaTypeProperty()
	document1MediaType.Set("image/jpeg")
	document1.SetActivityStreamsMediaType(document1MediaType)

	document1URL := streams.NewActivityStreamsUrlProperty()
	document1URL.AppendIRI(testrig.URLMustParse("https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg"))
	document1.SetActivityStreamsUrl(document1URL)

	document1Name := streams.NewActivityStreamsNameProperty()
	document1Name.AppendXMLSchemaString("It's a cute plushie.")
	document1.SetActivityStreamsName(document1Name)

	document1Blurhash := streams.NewTootBlurhashProperty()
	document1Blurhash.Set("UxQ0EkRP_4tRxtRjWBt7%hozM_ayV@oLf6WB")
	document1.SetTootBlurhash(document1Blurhash)

	return document1
}

func attachment1() vocab.ActivityStreamsAttachmentProperty {
	attachment1 := streams.NewActivityStreamsAttachmentProperty()
	attachment1.AppendActivityStreamsDocument(document1())
	return attachment1
}

type ExtractTestSuite struct {
	suite.Suite
}

func (suite *ExtractTestSuite) TestExtractAttachments() {
	note := streams.NewActivityStreamsNote()
	note.SetActivityStreamsAttachment(attachment1())

	attachments, err := ap.ExtractAttachments(note)
	suite.NoError(err)
	suite.Len(attachments, 1)

	attachment1 := attachments[0]
	suite.Equal("image/jpeg", attachment1.File.ContentType)
	suite.Equal("https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg", attachment1.RemoteURL)
	suite.Equal("It's a cute plushie.", attachment1.Description)
	suite.Empty(attachment1.Blurhash) // atm we discard blurhashes and generate them ourselves during processing
}

func (suite *ExtractTestSuite) TestExtractNoAttachments() {
	note := streams.NewActivityStreamsNote()

	attachments, err := ap.ExtractAttachments(note)
	suite.NoError(err)
	suite.Empty(attachments)
}

func (suite *ExtractTestSuite) TestExtractAttachmentsMissingContentType() {
	d1 := document1()
	d1.SetActivityStreamsMediaType(streams.NewActivityStreamsMediaTypeProperty())

	a1 := streams.NewActivityStreamsAttachmentProperty()
	a1.AppendActivityStreamsDocument(d1)

	note := streams.NewActivityStreamsNote()
	note.SetActivityStreamsAttachment(a1)

	attachments, err := ap.ExtractAttachments(note)
	suite.NoError(err)
	suite.Empty(attachments)
}

func (suite *ExtractTestSuite) TestExtractAttachmentMissingContentType() {

	d1 := document1()
	d1.SetActivityStreamsMediaType(streams.NewActivityStreamsMediaTypeProperty())

	attachment, err := ap.ExtractAttachment(d1)
	suite.EqualError(err, "no media type")
	suite.Nil(attachment)
}

func (suite *ExtractTestSuite) TestExtractAttachmentMissingURL() {
	d1 := document1()
	d1.SetActivityStreamsUrl(streams.NewActivityStreamsUrlProperty())

	attachment, err := ap.ExtractAttachment(d1)
	suite.EqualError(err, "could not extract url")
	suite.Nil(attachment)
}

func TestExtractTestSuite(t *testing.T) {
	suite.Run(t, &ExtractTestSuite{})
}
