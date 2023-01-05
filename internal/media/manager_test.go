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

package media_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"codeberg.org/gruf/go-store/v2/kv"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/stretchr/testify/suite"
	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ManagerTestSuite struct {
	MediaStandardTestSuite
}

func (suite *ManagerTestSuite) TestEmojiProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/rainbow-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, nil, "rainbow_test", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(36702, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/rainbow-original.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/rainbow-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingRefresh() {
	ctx := context.Background()

	// we're going to 'refresh' the remote 'yell' emoji by changing the image url to the pixellated gts logo
	originalEmoji := suite.testEmojis["yell"]

	emojiToUpdate := &gtsmodel.Emoji{}
	*emojiToUpdate = *originalEmoji
	newImageRemoteURL := "http://fossbros-anonymous.io/some/image/path.png"

	oldEmojiImagePath := emojiToUpdate.ImagePath
	oldEmojiImageStaticPath := emojiToUpdate.ImageStaticPath

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		b, err := os.ReadFile("./test/gts_pixellated-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := emojiToUpdate.ID
	emojiURI := emojiToUpdate.URI

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, nil, "yell", emojiID, emojiURI, &media.AdditionalEmojiInfo{
		CreatedAt:      &emojiToUpdate.CreatedAt,
		Domain:         &emojiToUpdate.Domain,
		ImageRemoteURL: &newImageRemoteURL,
	}, true)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(10296, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/gts_pixellated-original.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/gts_pixellated-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)

	// most fields should be different on the emoji now from what they were before
	suite.Equal(originalEmoji.ID, dbEmoji.ID)
	suite.NotEqual(originalEmoji.ImageRemoteURL, dbEmoji.ImageRemoteURL)
	suite.NotEqual(originalEmoji.ImageURL, dbEmoji.ImageURL)
	suite.NotEqual(originalEmoji.ImageStaticURL, dbEmoji.ImageStaticURL)
	suite.NotEqual(originalEmoji.ImageFileSize, dbEmoji.ImageFileSize)
	suite.NotEqual(originalEmoji.ImageStaticFileSize, dbEmoji.ImageStaticFileSize)
	suite.NotEqual(originalEmoji.ImagePath, dbEmoji.ImagePath)
	suite.NotEqual(originalEmoji.ImageStaticPath, dbEmoji.ImageStaticPath)
	suite.NotEqual(originalEmoji.ImageStaticPath, dbEmoji.ImageStaticPath)
	suite.NotEqual(originalEmoji.UpdatedAt, dbEmoji.UpdatedAt)
	suite.NotEqual(originalEmoji.ImageUpdatedAt, dbEmoji.ImageUpdatedAt)

	// the old image files should no longer be in storage
	_, err = suite.storage.Get(ctx, oldEmojiImagePath)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, oldEmojiImageStaticPath)
	suite.ErrorIs(err, storage.ErrNotFound)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingTooLarge() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/big-panda.gif")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, nil, "big_panda", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.EqualError(err, "store: given emoji fileSize (645688b) is larger than allowed size (51200b)")
	suite.Nil(emoji)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingTooLargeNoSizeGiven() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/big-panda.gif")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, nil, "big_panda", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.EqualError(err, "store: given emoji fileSize (645688b) is larger than allowed size (51200b)")
	suite.Nil(emoji)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingNoFileSizeGiven() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/rainbow-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), -1, nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	// process the media with no additional info provided
	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, nil, "rainbow_test", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(36702, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/rainbow-original.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/rainbow-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSlothVineProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/test-mp4-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(338, attachment.FileMeta.Original.Width)
	suite.Equal(240, attachment.FileMeta.Original.Height)
	suite.Equal(81120, attachment.FileMeta.Original.Size)
	suite.EqualValues(1.4083333, attachment.FileMeta.Original.Aspect)
	suite.EqualValues(6.5862, *attachment.FileMeta.Original.Duration)
	suite.EqualValues(29.000029, *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0x3b3e1, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 338, Height: 240, Size: 81120, Aspect: 1.4083333333333334,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(312413, attachment.File.FileSize)
	suite.Equal("", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-mp4-processed.mp4")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-mp4-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestLongerMp4ProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/longer-mp4-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(600, attachment.FileMeta.Original.Width)
	suite.Equal(330, attachment.FileMeta.Original.Height)
	suite.Equal(198000, attachment.FileMeta.Original.Size)
	suite.EqualValues(1.8181819, attachment.FileMeta.Original.Aspect)
	suite.EqualValues(16.6, *attachment.FileMeta.Original.Duration)
	suite.EqualValues(10, *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0xc8fb, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 600, Height: 330, Size: 198000, Aspect: 1.8181819,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(109549, attachment.File.FileSize)
	suite.Equal("", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/longer-mp4-processed.mp4")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/longer-mp4-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestNotAnMp4ProcessBlocking() {
	// try to load an 'mp4' that's actually an mkv in disguise

	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/not-an.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// pre processing should go fine but...
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)

	// we should get an error while loading
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.EqualError(err, "\"video width could not be discovered\",\"video height could not be discovered\",\"video duration could not be discovered\",\"video framerate could not be discovered\",\"video bitrate could not be discovered\"")
	suite.Nil(attachment)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingNoContentLengthGiven() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		// give length as -1 to indicate unknown
		return io.NopCloser(bytes.NewBuffer(b)), -1, nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingReadCloser() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// open test image as a file
		f, err := os.Open("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		// give length as -1 to indicate unknown
		return f, -1, nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestPngNoAlphaChannelProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-png-noalphachannel.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Small)
	suite.Equal("image/png", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(17471, attachment.File.FileSize)
	suite.Equal("LFQT7e.A%O%4?co$M}M{_1W9~TxV", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-png-noalphachannel-processed.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-png-noalphachannel-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestPngAlphaChannelProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-png-alphachannel.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Small)
	suite.Equal("image/png", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(18904, attachment.File.FileSize)
	suite.Equal("LFQT7e.A%O%4?co$M}M{_1W9~TxV", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-png-alphachannel-processed.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-png-alphachannel-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingWithCallback() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	// test the callback function by setting a simple boolean
	var calledPostData bool
	postData := func(_ context.Context) error {
		calledPostData = true
		return nil
	}
	suite.False(calledPostData) // not called yet (obvs)

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, postData, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// the post data callback should have been called
	suite.True(calledPostData)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessAsync() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// wait for the media to finish processing
	if !testrig.WaitFor(func() bool {
		return processingMedia.Finished()
	}) {
		suite.FailNow("timed out waiting for media to be processed")
	}

	// fetch the attachment from the database
	attachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegQueueSpamming() {
	// in this test, we spam the manager queue with 50 new media requests, just to see how it holds up
	ctx := context.Background()

	b, err := os.ReadFile("./test/test-jpeg.jpg")
	if err != nil {
		panic(err)
	}

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		return io.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	spam := 50
	inProcess := []*media.ProcessingMedia{}
	for i := 0; i < spam; i++ {
		// process the media with no additional info provided
		processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
		suite.NoError(err)
		inProcess = append(inProcess, processingMedia)
	}

	for _, processingMedia := range inProcess {
		// fetch the attachment id from the processing media
		attachmentID := processingMedia.AttachmentID()

		// do a blocking call to fetch the attachment
		attachment, err := processingMedia.LoadAttachment(ctx)
		suite.NoError(err)
		suite.NotNil(attachment)

		// make sure it's got the stuff set on it that we expect
		// the attachment ID and accountID we expect
		suite.Equal(attachmentID, attachment.ID)
		suite.Equal(accountID, attachment.AccountID)

		// file meta should be correctly derived from the image
		suite.EqualValues(gtsmodel.Original{
			Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
		}, attachment.FileMeta.Original)
		suite.EqualValues(gtsmodel.Small{
			Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
		}, attachment.FileMeta.Small)
		suite.Equal("image/jpeg", attachment.File.ContentType)
		suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
		suite.Equal(269739, attachment.File.FileSize)
		suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

		// now make sure the attachment is in the database
		dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
		suite.NoError(err)
		suite.NotNil(dbAttachment)

		// make sure the processed file is in storage
		processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
		suite.NoError(err)
		suite.NotEmpty(processedFullBytes)

		// load the processed bytes from our test folder, to compare
		processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
		suite.NoError(err)
		suite.NotEmpty(processedFullBytesExpected)

		// the bytes in storage should be what we expected
		suite.Equal(processedFullBytesExpected, processedFullBytes)

		// now do the same for the thumbnail and make sure it's what we expected
		processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
		suite.NoError(err)
		suite.NotEmpty(processedThumbnailBytes)

		processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
		suite.NoError(err)
		suite.NotEmpty(processedThumbnailBytesExpected)

		suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
	}
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingWithDiskStorage() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	temp := fmt.Sprintf("%s/gotosocial-test", os.TempDir())
	defer os.RemoveAll(temp)

	disk, err := storage.OpenDisk(temp, &storage.DiskConfig{
		LockFile: path.Join(temp, "store.lock"),
	})
	if err != nil {
		panic(err)
	}

	storage := &gtsstorage.Driver{
		KVStore: kv.New(disk),
		Storage: disk,
	}

	diskManager, err := media.NewManager(suite.db, storage)
	if err != nil {
		panic(err)
	}
	suite.manager = diskManager

	// process the media with no additional info provided
	processingMedia, err := diskManager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
