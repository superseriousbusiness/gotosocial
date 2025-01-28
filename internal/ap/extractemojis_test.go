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
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type ExtractEmojisTestSuite struct {
	APTestSuite
}

func (suite *ExtractEmojisTestSuite) TestExtractEmojis() {
	const noteWithEmojis = `{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "attributedTo": "https://example.org/users/tobi",
  "content": "<p>i hear that the GoToSocial devs are anti-capitalists and even <em>shocked gasp</em> communists :shocked_pikachu: totally unreasonable people</p>",
  "id": "https://example.org/users/tobi/statuses/01HV11D2BS7M94ZS499VBW7RX5",
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png"
    },
    "id": "https://example.org/emoji/01AZY1Y5YQD6TREB5W50HGTCSZ",
    "name": ":shocked_pikachu:",
    "type": "Emoji",
    "updated": "2022-11-17T11:36:05Z"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note"
}`

	statusable, err := ap.ResolveStatusable(
		context.Background(),
		io.NopCloser(bytes.NewBufferString(noteWithEmojis)),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	emojis, err := ap.ExtractEmojis(statusable, "example.org")
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(emojis); l != 1 {
		suite.FailNow("", "expected length 1 for emojis, got %d", l)
	}

	emoji := emojis[0]
	suite.Equal("shocked_pikachu", emoji.Shortcode)
	suite.Equal("example.org", emoji.Domain)
	suite.Equal("https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png", emoji.ImageRemoteURL)
	suite.False(*emoji.Disabled)
	suite.Equal("https://example.org/emoji/01AZY1Y5YQD6TREB5W50HGTCSZ", emoji.URI)
	suite.False(*emoji.VisibleInPicker)
}

func (suite *ExtractEmojisTestSuite) TestExtractEmojisNoID() {
	const noteWithEmojis = `{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "attributedTo": "https://example.org/users/tobi",
  "content": "<p>i hear that the GoToSocial devs are anti-capitalists and even <em>shocked gasp</em> communists :shocked_pikachu: totally unreasonable people</p>",
  "id": "https://example.org/users/tobi/statuses/01HV11D2BS7M94ZS499VBW7RX5",
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png"
    },
    "name": ":shocked_pikachu:",
    "type": "Emoji",
    "updated": "2022-11-17T11:36:05Z"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note"
}`

	statusable, err := ap.ResolveStatusable(
		context.Background(),
		io.NopCloser(bytes.NewBufferString(noteWithEmojis)),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	emojis, err := ap.ExtractEmojis(statusable, "example.org")
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(emojis); l != 1 {
		suite.FailNow("", "expected length 1 for emojis, got %d", l)
	}

	emoji := emojis[0]
	suite.Equal("shocked_pikachu", emoji.Shortcode)
	suite.Equal("example.org", emoji.Domain)
	suite.Equal("https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png", emoji.ImageRemoteURL)
	suite.False(*emoji.Disabled)
	suite.Equal("https://example.org/dummy_emoji_path?shortcode=shocked_pikachu", emoji.URI)
	suite.False(*emoji.VisibleInPicker)
}

func (suite *ExtractEmojisTestSuite) TestExtractEmojisNullID() {
	const noteWithEmojis = `{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
    "attributedTo": "https://example.org/users/tobi",
  "content": "<p>i hear that the GoToSocial devs are anti-capitalists and even <em>shocked gasp</em> communists :shocked_pikachu: totally unreasonable people</p>",
  "id": "https://example.org/users/tobi/statuses/01HV11D2BS7M94ZS499VBW7RX5",
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png"
    },
    "id": null,
    "name": ":shocked_pikachu:",
    "type": "Emoji",
    "updated": "2022-11-17T11:36:05Z"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note"
}`

	statusable, err := ap.ResolveStatusable(
		context.Background(),
		io.NopCloser(bytes.NewBufferString(noteWithEmojis)),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	emojis, err := ap.ExtractEmojis(statusable, "example.org")
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(emojis); l != 1 {
		suite.FailNow("", "expected length 1 for emojis, got %d", l)
	}

	emoji := emojis[0]
	suite.Equal("shocked_pikachu", emoji.Shortcode)
	suite.Equal("example.org", emoji.Domain)
	suite.Equal("https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png", emoji.ImageRemoteURL)
	suite.False(*emoji.Disabled)
	suite.Equal("https://example.org/dummy_emoji_path?shortcode=shocked_pikachu", emoji.URI)
	suite.False(*emoji.VisibleInPicker)
}

func (suite *ExtractEmojisTestSuite) TestExtractEmojisEmptyID() {
	const noteWithEmojis = `{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
    "attributedTo": "https://example.org/users/tobi",
  "content": "<p>i hear that the GoToSocial devs are anti-capitalists and even <em>shocked gasp</em> communists :shocked_pikachu: totally unreasonable people</p>",
  "id": "https://example.org/users/tobi/statuses/01HV11D2BS7M94ZS499VBW7RX5",
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png"
    },
    "id": "",
    "name": ":shocked_pikachu:",
    "type": "Emoji",
    "updated": "2022-11-17T11:36:05Z"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note"
}`

	statusable, err := ap.ResolveStatusable(
		context.Background(),
		io.NopCloser(bytes.NewBufferString(noteWithEmojis)),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	emojis, err := ap.ExtractEmojis(statusable, "example.org")
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(emojis); l != 1 {
		suite.FailNow("", "expected length 1 for emojis, got %d", l)
	}

	emoji := emojis[0]
	suite.Equal("shocked_pikachu", emoji.Shortcode)
	suite.Equal("example.org", emoji.Domain)
	suite.Equal("https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png", emoji.ImageRemoteURL)
	suite.False(*emoji.Disabled)
	suite.Equal("https://example.org/dummy_emoji_path?shortcode=shocked_pikachu", emoji.URI)
	suite.False(*emoji.VisibleInPicker)
}

func TestExtractEmojisTestSuite(t *testing.T) {
	suite.Run(t, &ExtractEmojisTestSuite{})
}
