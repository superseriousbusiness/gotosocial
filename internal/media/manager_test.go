/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"time"

	"codeberg.org/gruf/go-store/kv"
	"codeberg.org/gruf/go-store/storage"
	"github.com/stretchr/testify/suite"
	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20211113114307_init"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

type ManagerTestSuite struct {
	MediaStandardTestSuite
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.Reader, int, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return bytes.NewBuffer(b), len(b), nil
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
	suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingWithCallback() {
	ctx := context.Background()

	data := func(_ context.Context) (io.Reader, int, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return bytes.NewBuffer(b), len(b), nil
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
	suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessAsync() {
	ctx := context.Background()

	data := func(_ context.Context) (io.Reader, int, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return bytes.NewBuffer(b), len(b), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia, err := suite.manager.ProcessMedia(ctx, data, nil, accountID, nil)
	suite.NoError(err)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// wait for the media to finish processing
	for finished := processingMedia.Finished(); !finished; finished = processingMedia.Finished() {
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("\n\nnot finished yet...\n\n")
	}
	fmt.Printf("\n\nfinished!\n\n")

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
	suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(attachment.Thumbnail.Path)
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

	data := func(_ context.Context) (io.Reader, int, error) {
		// load bytes from a test image
		return bytes.NewReader(b), len(b), nil
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
		fmt.Printf("\n\n\nactive workers: %d, queue length: %d\n\n\n", suite.manager.ActiveWorkers(), suite.manager.JobsQueued())

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
		suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", attachment.Blurhash)

		// now make sure the attachment is in the database
		dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
		suite.NoError(err)
		suite.NotNil(dbAttachment)

		// make sure the processed file is in storage
		processedFullBytes, err := suite.storage.Get(attachment.File.Path)
		suite.NoError(err)
		suite.NotEmpty(processedFullBytes)

		// load the processed bytes from our test folder, to compare
		processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
		suite.NoError(err)
		suite.NotEmpty(processedFullBytesExpected)

		// the bytes in storage should be what we expected
		suite.Equal(processedFullBytesExpected, processedFullBytes)

		// now do the same for the thumbnail and make sure it's what we expected
		processedThumbnailBytes, err := suite.storage.Get(attachment.Thumbnail.Path)
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

	data := func(_ context.Context) (io.Reader, int, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return bytes.NewBuffer(b), len(b), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	temp := fmt.Sprintf("%s/gotosocial-test", os.TempDir())
	defer os.RemoveAll(temp)

	diskStorage, err := kv.OpenFile(temp, &storage.DiskConfig{
		LockFile: path.Join(temp, "store.lock"),
	})
	if err != nil {
		panic(err)
	}

	diskManager, err := media.NewManager(suite.db, diskStorage)
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
	suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := diskStorage.Get(attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := diskStorage.Get(attachment.Thumbnail.Path)
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
