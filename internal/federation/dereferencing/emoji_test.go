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

package dereferencing_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

type EmojiTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *EmojiTestSuite) TestDereferenceEmojiBlocking() {
	ctx := context.Background()
	fetchingAccount := suite.testAccounts["local_account_1"]
	emojiImageRemoteURL := "http://example.org/media/emojis/1781772.gif"
	emojiImageStaticRemoteURL := "http://example.org/media/emojis/1781772.gif"
	emojiURI := "http://example.org/emojis/1781772"
	emojiShortcode := "peglin"
	emojiID := "01GCBMGNZBKMEE1KTZ6PMJEW5D"
	emojiDomain := "example.org"
	emojiDisabled := false
	emojiVisibleInPicker := false

	ai := &media.AdditionalEmojiInfo{
		Domain:               &emojiDomain,
		ImageRemoteURL:       &emojiImageRemoteURL,
		ImageStaticRemoteURL: &emojiImageStaticRemoteURL,
		Disabled:             &emojiDisabled,
		VisibleInPicker:      &emojiVisibleInPicker,
	}

	processingEmoji, err := suite.dereferencer.GetRemoteEmoji(ctx, fetchingAccount.Username, emojiImageRemoteURL, emojiShortcode, emojiDomain, emojiID, emojiURI, ai, false)
	suite.NoError(err)

	// make a blocking call to load the emoji from the in-process media
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	suite.Equal(emojiID, emoji.ID)
	suite.WithinDuration(time.Now(), emoji.CreatedAt, 10*time.Second)
	suite.WithinDuration(time.Now(), emoji.UpdatedAt, 10*time.Second)
	suite.Equal(emojiShortcode, emoji.Shortcode)
	suite.Equal(emojiDomain, emoji.Domain)
	suite.Equal(emojiImageRemoteURL, emoji.ImageRemoteURL)
	suite.Equal(emojiImageStaticRemoteURL, emoji.ImageStaticRemoteURL)
	suite.Contains(emoji.ImageURL, "/emoji/original/01GCBMGNZBKMEE1KTZ6PMJEW5D.gif")
	suite.Contains(emoji.ImageStaticURL, "emoji/static/01GCBMGNZBKMEE1KTZ6PMJEW5D.png")
	suite.Contains(emoji.ImagePath, "/emoji/original/01GCBMGNZBKMEE1KTZ6PMJEW5D.gif")
	suite.Contains(emoji.ImageStaticPath, "/emoji/static/01GCBMGNZBKMEE1KTZ6PMJEW5D.png")
	suite.Equal("image/gif", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(37796, emoji.ImageFileSize)
	suite.Equal(7951, emoji.ImageStaticFileSize)
	suite.WithinDuration(time.Now(), emoji.ImageUpdatedAt, 10*time.Second)
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
