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
	"context"
	"encoding/json"
	"testing"

	"codeberg.org/superseriousbusiness/activity/streams"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type ExtractAttachmentsTestSuite struct {
	APTestSuite
}

func (suite *ExtractAttachmentsTestSuite) TestExtractAttachmentMissingURL() {
	d1 := suite.document1
	d1.SetActivityStreamsUrl(streams.NewActivityStreamsUrlProperty())

	attachment, err := ap.ExtractAttachment(d1)
	suite.EqualError(err, "ExtractAttachment: error extracting attachment URL: ExtractURL: no valid URL property found")
	suite.Nil(attachment)
}

func (suite *ExtractAttachmentsTestSuite) TestExtractDescription() {
	// Note: normally a single attachment on a Note or
	// similar wouldn't have the `@context` field set,
	// but we set it here because we're parsing it as
	// a discrete/standalone AP Object for this test.
	attachmentableJSON := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "mediaType": "image/jpeg",
  "name": "z64KTcw2h2bZ8s67k2.jpg",
  "summary": "A very large panel that is entirely twist switches",
  "type": "Document",
  "url": "https://example.org/d/XzKw4M2Sc1pBxj3hY4.jpg"
}`

	raw := make(map[string]interface{})
	if err := json.Unmarshal([]byte(attachmentableJSON), &raw); err != nil {
		suite.FailNow(err.Error())
	}

	t, err := streams.ToType(context.Background(), raw)
	if err != nil {
		suite.FailNow(err.Error())
	}

	attachmentable, ok := t.(ap.Attachmentable)
	if !ok {
		suite.FailNow("type was not Attachmentable")
	}

	attachment, err := ap.ExtractAttachment(attachmentable)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("A very large panel that is entirely twist switches", attachment.Description)
}

func TestExtractAttachmentsTestSuite(t *testing.T) {
	suite.Run(t, &ExtractAttachmentsTestSuite{})
}
