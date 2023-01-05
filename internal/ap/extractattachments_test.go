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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type ExtractAttachmentsTestSuite struct {
	ExtractTestSuite
}

func (suite *ExtractAttachmentsTestSuite) TestExtractAttachments() {
	note := streams.NewActivityStreamsNote()
	note.SetActivityStreamsAttachment(suite.attachment1)

	attachments, err := ap.ExtractAttachments(note)
	suite.NoError(err)
	suite.Len(attachments, 1)

	attachment1 := attachments[0]
	suite.Equal("image/jpeg", attachment1.File.ContentType)
	suite.Equal("https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg", attachment1.RemoteURL)
	suite.Equal("It's a cute plushie.", attachment1.Description)
	suite.Equal("UxQ0EkRP_4tRxtRjWBt7%hozM_ayV@oLf6WB", attachment1.Blurhash)
}

func (suite *ExtractAttachmentsTestSuite) TestExtractNoAttachments() {
	note := streams.NewActivityStreamsNote()

	attachments, err := ap.ExtractAttachments(note)
	suite.NoError(err)
	suite.Empty(attachments)
}

func (suite *ExtractAttachmentsTestSuite) TestExtractAttachmentsMissingContentType() {
	d1 := suite.document1
	d1.SetActivityStreamsMediaType(streams.NewActivityStreamsMediaTypeProperty())

	a1 := streams.NewActivityStreamsAttachmentProperty()
	a1.AppendActivityStreamsDocument(d1)

	note := streams.NewActivityStreamsNote()
	note.SetActivityStreamsAttachment(a1)

	attachments, err := ap.ExtractAttachments(note)
	suite.NoError(err)
	suite.Empty(attachments)
}

func (suite *ExtractAttachmentsTestSuite) TestExtractAttachmentMissingContentType() {
	d1 := suite.document1
	d1.SetActivityStreamsMediaType(streams.NewActivityStreamsMediaTypeProperty())

	attachment, err := ap.ExtractAttachment(d1)
	suite.EqualError(err, "no media type")
	suite.Nil(attachment)
}

func (suite *ExtractAttachmentsTestSuite) TestExtractAttachmentMissingURL() {
	d1 := suite.document1
	d1.SetActivityStreamsUrl(streams.NewActivityStreamsUrlProperty())

	attachment, err := ap.ExtractAttachment(d1)
	suite.EqualError(err, "could not extract url")
	suite.Nil(attachment)
}

func TestExtractAttachmentsTestSuite(t *testing.T) {
	suite.Run(t, &ExtractAttachmentsTestSuite{})
}
