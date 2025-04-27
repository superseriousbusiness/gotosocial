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

package media_test

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	gtsstorage "code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"codeberg.org/gruf/go-iotools"
	"codeberg.org/gruf/go-storage/disk"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	MediaStandardTestSuite
}

func (suite *ManagerTestSuite) TestEmojiProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/rainbow-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	processing, err := suite.manager.CreateEmoji(ctx,
		"rainbow_test",
		"",
		data,
		media.AdditionalEmojiInfo{},
	)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// file meta should be correctly derived from the image
	suite.Equal("image/apng", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(36702, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emoji.ID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbEmoji.ImagePath, "./test/rainbow-original.png")
	equalFiles(suite.T(), suite.state.Storage, dbEmoji.ImageStaticPath, "./test/rainbow-static.png")
}

func (suite *ManagerTestSuite) TestEmojiProcessRefresh() {
	ctx := context.Background()

	// we're going to 'refresh' the remote 'yell' emoji by changing the image url to the pixellated gts logo
	originalEmoji := suite.testEmojis["yell"]

	emojiToUpdate, err := suite.db.GetEmojiByID(ctx, originalEmoji.ID)
	suite.NoError(err)

	newImageRemoteURL := "http://fossbros-anonymous.io/some/image/path.png"

	oldEmojiImagePath := emojiToUpdate.ImagePath
	oldEmojiImageStaticPath := emojiToUpdate.ImageStaticPath

	data := func(_ context.Context) (io.ReadCloser, error) {
		b, err := os.ReadFile("./test/gts_pixellated-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	processing, err := suite.manager.UpdateEmoji(ctx,
		emojiToUpdate,
		data,
		media.AdditionalEmojiInfo{
			Domain:         &emojiToUpdate.Domain,
			ImageRemoteURL: &newImageRemoteURL,
		},
	)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(originalEmoji.ID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(10296, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emoji.ID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbEmoji.ImagePath, "./test/gts_pixellated-original.png")
	equalFiles(suite.T(), suite.state.Storage, dbEmoji.ImageStaticPath, "./test/gts_pixellated-static.png")

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

	// the old image files should no longer be in storage
	_, err = suite.storage.Get(ctx, oldEmojiImagePath)
	suite.True(storage.IsNotFound(err))
	_, err = suite.storage.Get(ctx, oldEmojiImageStaticPath)
	suite.True(storage.IsNotFound(err))
}

func (suite *ManagerTestSuite) TestEmojiProcessTooLarge() {
	ctx := context.Background()

	// Open test image as file for reading.
	file, err := os.Open("./test/big-panda.gif")
	if err != nil {
		panic(err)
	}

	// Get file size info.
	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	// Set max allowed size UNDER image size.
	lr := io.LimitReader(file, stat.Size()-10)
	rc := iotools.ReadCloser(lr, file)

	processing, err := suite.manager.CreateEmoji(ctx,
		"big_panda",
		"",
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return rc, nil
		},
		media.AdditionalEmojiInfo{},
	)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	_, err = processing.Load(ctx)
	suite.EqualError(err, "store: error draining data to tmp: reached read limit 630kiB")
}

func (suite *ManagerTestSuite) TestEmojiWebpProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/nb-flag-original.webp")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	// process the media with no additional info provided
	processing, err := suite.manager.CreateEmoji(ctx,
		"nb-flag",
		"",
		data,
		media.AdditionalEmojiInfo{},
	)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// file meta should be correctly derived from the image
	suite.Equal("image/webp", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(294, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emoji.ID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// ensure files are equal
	equalFiles(suite.T(), suite.state.Storage, dbEmoji.ImagePath, "./test/nb-flag-original.webp")
	equalFiles(suite.T(), suite.state.Storage, dbEmoji.ImageStaticPath, "./test/nb-flag-static.png")
}

func (suite *ManagerTestSuite) TestSimpleJpegProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
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
	suite.Equal(22858, attachment.Thumbnail.FileSize)
	suite.Equal("LiB|W-#6RQR.~qvzRjWF_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/test-jpeg-processed.jpg")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/test-jpeg-thumbnail.jpeg")
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessTooLarge() {
	ctx := context.Background()

	// Open test image as file for reading.
	file, err := os.Open("./test/test-jpeg.jpg")
	if err != nil {
		panic(err)
	}

	// Get file size info.
	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	// Set max allowed size UNDER image size.
	lr := io.LimitReader(file, stat.Size()-10)
	rc := iotools.ReadCloser(lr, file)

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return rc, nil
		},
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	_, err = processing.Load(ctx)
	suite.EqualError(err, "store: error draining data to tmp: reached read limit 263kiB")
}

func (suite *ManagerTestSuite) TestPDFProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from Frantz
		b, err := os.ReadFile("./test/Frantz-Fanon-The-Wretched-of-the-Earth-1965.pdf")
		if err != nil {
			panic(err)
		}

		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	suite.Zero(attachment.FileMeta)
	suite.Zero(attachment.File.ContentType)
	suite.Zero(attachment.Thumbnail.ContentType)
	suite.Zero(attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// Attachment should have type unknown
	suite.Equal(gtsmodel.FileTypeUnknown, dbAttachment.Type)

	// Nothing should be in storage for this attachment.
	stored, err := suite.storage.Has(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.False(stored)
	stored, err = suite.storage.Has(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.False(stored)
}

func (suite *ManagerTestSuite) TestSlothVineProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/test-mp4-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(338, attachment.FileMeta.Original.Width)
	suite.Equal(240, attachment.FileMeta.Original.Height)
	suite.Equal(81120, attachment.FileMeta.Original.Size)
	suite.EqualValues(float32(1.4083333), attachment.FileMeta.Original.Aspect)
	suite.EqualValues(float32(6.641), *attachment.FileMeta.Original.Duration)
	suite.EqualValues(float32(29), *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0x5be18, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 338, Height: 240, Size: 81120, Aspect: 1.4083333333333334,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/webp", attachment.Thumbnail.ContentType)
	suite.Equal(312453, attachment.File.FileSize)
	suite.Equal(5598, attachment.Thumbnail.FileSize)
	suite.Equal("LgIYH}xtNsofxtfPW.j[_4axn+of", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/test-mp4-processed.mp4")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/test-mp4-thumbnail.webp")
}

func (suite *ManagerTestSuite) TestAnimatedGifProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/clock-original.gif")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width:     528,
		Height:    528,
		Size:      278784,
		Aspect:    1,
		Duration:  util.Ptr(float32(8.58)),
		Framerate: util.Ptr(float32(16)),
		Bitrate:   util.Ptr(uint64(114092)),
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width:  512,
		Height: 512,
		Size:   262144,
		Aspect: 1,
	}, attachment.FileMeta.Small)
	suite.Equal("image/gif", attachment.File.ContentType)
	suite.Equal("image/webp", attachment.Thumbnail.ContentType)
	suite.Equal(122364, attachment.File.FileSize)
	suite.Equal(12962, attachment.Thumbnail.FileSize)
	suite.Equal("LmKUZkt700ofoffQofj[00WBj[WB", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/clock-processed.gif")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/clock-thumbnail.webp")
}

func (suite *ManagerTestSuite) TestLongerMp4Process() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/longer-mp4-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(600, attachment.FileMeta.Original.Width)
	suite.Equal(330, attachment.FileMeta.Original.Height)
	suite.Equal(198000, attachment.FileMeta.Original.Size)
	suite.EqualValues(float32(1.8181819), attachment.FileMeta.Original.Aspect)
	suite.EqualValues(float32(16.6), *attachment.FileMeta.Original.Duration)
	suite.EqualValues(float32(10), *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0xce3a, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 281, Size: 143872, Aspect: 1.8181819,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/webp", attachment.Thumbnail.ContentType)
	suite.Equal(109569, attachment.File.FileSize)
	suite.Equal(2958, attachment.Thumbnail.FileSize)
	suite.Equal("LIQ9}}_3IU?b~qM{ofayWBWVofRj", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/longer-mp4-processed.mp4")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/longer-mp4-thumbnail.webp")
}

func (suite *ManagerTestSuite) TestBirdnestMp4Process() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/birdnest-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(404, attachment.FileMeta.Original.Width)
	suite.Equal(720, attachment.FileMeta.Original.Height)
	suite.Equal(290880, attachment.FileMeta.Original.Size)
	suite.EqualValues(float32(0.5611111), attachment.FileMeta.Original.Aspect)
	suite.EqualValues(float32(9.823), *attachment.FileMeta.Original.Duration)
	suite.EqualValues(float32(30), *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0x11844c, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 287, Height: 512, Size: 146944, Aspect: 0.5611111,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/webp", attachment.Thumbnail.ContentType)
	suite.Equal(1409625, attachment.File.FileSize)
	suite.Equal(15056, attachment.Thumbnail.FileSize)
	suite.Equal("LLF$nqafRO.9DgM_RPadtkV@WCMx", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/birdnest-processed.mp4")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/birdnest-thumbnail.webp")
}

func (suite *ManagerTestSuite) TestOpusProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-opus-original.opus")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Duration: util.Ptr(float32(122.10006)),
		Bitrate:  util.Ptr(uint64(116426)),
	}, attachment.FileMeta.Original)
	suite.Equal("audio/opus", attachment.File.ContentType)
	suite.Equal(1776956, attachment.File.FileSize)
	suite.Empty(attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/test-opus-processed.opus")
	suite.Zero(dbAttachment.Thumbnail.FileSize)
}

func (suite *ManagerTestSuite) TestPngNoAlphaChannelProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-png-noalphachannel.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
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
	suite.Equal(6446, attachment.Thumbnail.FileSize)
	suite.Equal("LGP%YL.A-?tA.9o#RURQ~ojp^~xW", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/test-png-noalphachannel-processed.png")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/test-png-noalphachannel-thumbnail.jpeg")
}

func (suite *ManagerTestSuite) TestPngAlphaChannelProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-png-alphachannel.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Small)
	suite.Equal("image/png", attachment.File.ContentType)
	suite.Equal("image/webp", attachment.Thumbnail.ContentType)
	suite.Equal(18832, attachment.File.FileSize)
	suite.Equal(3592, attachment.Thumbnail.FileSize)
	suite.Equal("LCN^lE.A-?xd?co#N1RQ~ojp~SxW", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/test-png-alphachannel-processed.png")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/test-png-alphachannel-thumbnail.jpeg")
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessWithCallback() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
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
	suite.Equal(22858, attachment.Thumbnail.FileSize)
	suite.Equal("LiB|W-#6RQR.~qvzRjWF_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.File.Path, "./test/test-jpeg-processed.jpg")
	equalFiles(suite.T(), suite.state.Storage, dbAttachment.Thumbnail.Path, "./test/test-jpeg-thumbnail.jpeg")
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessWithDiskStorage() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	temp := fmt.Sprintf("%s/gotosocial-test", os.TempDir())
	defer os.RemoveAll(temp)

	disk, err := disk.Open(temp, nil)
	if err != nil {
		panic(err)
	}

	var state state.State

	testrig.StartNoopWorkers(&state)
	defer testrig.StopWorkers(&state)

	storage := &gtsstorage.Driver{
		Storage: disk,
	}
	state.Storage = storage
	state.DB = suite.db

	diskManager := media.NewManager(&state)
	suite.manager = diskManager

	// process the media with no additional info provided
	processing, err := suite.manager.CreateMedia(ctx,
		accountID,
		data,
		media.AdditionalMediaInfo{},
	)
	suite.NoError(err)
	suite.NotNil(processing)

	// do a blocking call to fetch the attachment
	attachment, err := processing.Load(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(processing.ID(), attachment.ID)
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
	suite.Equal(22858, attachment.Thumbnail.FileSize)
	suite.Equal("LiB|W-#6RQR.~qvzRjWF_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// ensure the files contain the expected data.
	equalFiles(suite.T(), storage, dbAttachment.File.Path, "./test/test-jpeg-processed.jpg")
	equalFiles(suite.T(), storage, dbAttachment.Thumbnail.Path, "./test/test-jpeg-thumbnail.jpeg")
}

func (suite *ManagerTestSuite) TestSmallSizedMediaTypeDetection_issue2263() {
	for index, test := range []struct {
		name     string // Test title
		path     string // File path
		expected string // Expected ContentType
	}{
		{
			name:     "big size JPEG",
			path:     "./test/test-jpeg.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "big size PNG",
			path:     "./test/test-png-noalphachannel.png",
			expected: "image/png",
		},
		{
			name:     "small size JPEG",
			path:     "./test/test-jpeg-1x1px-white.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "golden case PNG (big size)",
			path:     "./test/test-png-alphachannel-1x1px.png",
			expected: "image/png",
		},
	} {
		suite.Run(test.name, func() {
			ctx, cncl := context.WithTimeout(context.Background(), time.Second*60)
			defer cncl()

			data := func(_ context.Context) (io.ReadCloser, error) {
				// load bytes from a test image
				b, err := os.ReadFile(test.path)
				suite.NoError(err, "Test %d: failed during test setup", index+1)

				return io.NopCloser(bytes.NewBuffer(b)), nil
			}

			accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

			// process the media with no additional info provided
			processing, err := suite.manager.CreateMedia(ctx,
				accountID,
				data,
				media.AdditionalMediaInfo{},
			)
			suite.NoError(err)
			suite.NotNil(processing)

			// Load the attachment (but ignore return).
			_, err = processing.Load(ctx)
			suite.NoError(err)

			// fetch the attachment id from the processing media
			attachment, err := suite.db.GetAttachmentByID(ctx, processing.ID())
			if err != nil {
				suite.FailNow(err.Error())
			}

			// make sure it's got the stuff set on it that we expect
			// the attachment ID and accountID we expect
			suite.Equal(processing.ID(), attachment.ID)
			suite.Equal(accountID, attachment.AccountID)

			actual := attachment.File.ContentType
			expect := test.expected

			suite.Equal(expect, actual, "Test %d: %s", index+1, test.name)
		})
	}
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

// equalFiles checks whether
func equalFiles(t *testing.T, st *storage.Driver, storagePath, testPath string) {
	b1, err := st.Get(context.Background(), storagePath)
	if err != nil {
		t.Fatalf("error reading file %s: %v", storagePath, err)
	}

	b2, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("error reading file %s: %v", testPath, err)
	}

	if md5.Sum(b1) != md5.Sum(b2) {
		t.Errorf("%s != %s", storagePath, testPath)
	}
}
