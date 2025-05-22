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

package dereferencing_test

import (
	"fmt"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/media"
	"github.com/stretchr/testify/suite"
)

type EmojiTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *EmojiTestSuite) TestDereferenceEmojiBlocking() {
	ctx := suite.T().Context()
	emojiImageRemoteURL := "http://example.org/media/emojis/1781772.gif"
	emojiImageStaticRemoteURL := "http://example.org/media/emojis/1781772.gif"
	emojiURI := "http://example.org/emojis/1781772"
	emojiShortcode := "peglin"
	emojiDomain := "example.org"
	emojiDisabled := false
	emojiVisibleInPicker := false

	emoji, err := suite.dereferencer.GetEmoji(
		ctx,
		emojiShortcode,
		emojiDomain,
		emojiImageRemoteURL,
		media.AdditionalEmojiInfo{
			URI:                  &emojiURI,
			Domain:               &emojiDomain,
			ImageRemoteURL:       &emojiImageRemoteURL,
			ImageStaticRemoteURL: &emojiImageStaticRemoteURL,
			Disabled:             &emojiDisabled,
			VisibleInPicker:      &emojiVisibleInPicker,
		},
		false,
	)
	suite.NoError(err)
	suite.NotNil(emoji)

	expectPath := fmt.Sprintf("/emoji/original/%s.gif", emoji.ID)
	expectStaticPath := fmt.Sprintf("/emoji/static/%s.png", emoji.ID)

	suite.WithinDuration(time.Now(), emoji.CreatedAt, 10*time.Second)
	suite.WithinDuration(time.Now(), emoji.UpdatedAt, 10*time.Second)
	suite.Equal(emojiShortcode, emoji.Shortcode)
	suite.Equal(emojiDomain, emoji.Domain)
	suite.Equal(emojiImageRemoteURL, emoji.ImageRemoteURL)
	suite.Equal(emojiImageStaticRemoteURL, emoji.ImageStaticRemoteURL)
	suite.Contains(emoji.ImageURL, expectPath)
	suite.Contains(emoji.ImageStaticURL, expectStaticPath)
	suite.Contains(emoji.ImagePath, expectPath)
	suite.Contains(emoji.ImageStaticPath, expectStaticPath)
	suite.Equal("image/gif", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(37796, emoji.ImageFileSize)
	suite.Equal(9824, emoji.ImageStaticFileSize)
	suite.WithinDuration(time.Now(), emoji.UpdatedAt, 10*time.Second)
	suite.False(*emoji.Disabled)
	suite.Equal(emojiURI, emoji.URI)
	suite.False(*emoji.VisibleInPicker)
	suite.Empty(emoji.CategoryID)

	// ensure that emoji is now in storage
	stored, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.Len(stored, emoji.ImageFileSize)

	storedStatic, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.Len(storedStatic, emoji.ImageStaticFileSize)
}

func TestEmojiTestSuite(t *testing.T) {
	suite.Run(t, new(EmojiTestSuite))
}
