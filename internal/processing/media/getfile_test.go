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
	"context"
	"io"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

type GetFileTestSuite struct {
	MediaStandardTestSuite
}

func (suite *GetFileTestSuite) TestGetRemoteFileCached() {
	ctx := context.Background()

	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	fileName := path.Base(testAttachment.File.Path)
	requestingAccount := suite.testAccounts["local_account_1"]

	content, errWithCode := suite.mediaProcessor.GetFile(ctx, requestingAccount, &apimodel.GetContentRequestForm{
		AccountID: testAttachment.AccountID,
		MediaType: string(media.TypeAttachment),
		MediaSize: string(media.SizeOriginal),
		FileName:  fileName,
	})

	suite.NoError(errWithCode)
	suite.NotNil(content)
	b, err := io.ReadAll(content.Content)
	suite.NoError(err)

	if closer, ok := content.Content.(io.Closer); ok {
		suite.NoError(closer.Close())
	}

	suite.Equal(suite.testRemoteAttachments[testAttachment.RemoteURL].Data, b)
	suite.Equal(suite.testRemoteAttachments[testAttachment.RemoteURL].ContentType, content.ContentType)
	suite.EqualValues(len(suite.testRemoteAttachments[testAttachment.RemoteURL].Data), content.ContentLength)
}

func (suite *GetFileTestSuite) TestGetRemoteFileUncached() {
	ctx := context.Background()

	// uncache the file from local
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	testAttachment.Cached = false
	err := suite.db.UpdateByPrimaryKey(ctx, testAttachment)
	suite.NoError(err)
	err = suite.storage.Delete(testAttachment.File.Path)
	suite.NoError(err)
	err = suite.storage.Delete(testAttachment.Thumbnail.Path)
	suite.NoError(err)

	// now fetch it
	fileName := path.Base(testAttachment.File.Path)
	requestingAccount := suite.testAccounts["local_account_1"]

	content, errWithCode := suite.mediaProcessor.GetFile(ctx, requestingAccount, &apimodel.GetContentRequestForm{
		AccountID: testAttachment.AccountID,
		MediaType: string(media.TypeAttachment),
		MediaSize: string(media.SizeOriginal),
		FileName:  fileName,
	})

	suite.NoError(errWithCode)
	suite.NotNil(content)
	b, err := io.ReadAll(content.Content)
	suite.NoError(err)

	if closer, ok := content.Content.(io.Closer); ok {
		suite.NoError(closer.Close())
	}

	suite.Equal(suite.testRemoteAttachments[testAttachment.RemoteURL].Data, b)
	suite.Equal(suite.testRemoteAttachments[testAttachment.RemoteURL].ContentType, content.ContentType)
	suite.EqualValues(len(suite.testRemoteAttachments[testAttachment.RemoteURL].Data), content.ContentLength)
	time.Sleep(2 * time.Second) // wait a few seconds for the media manager to finish doing stuff

	// the attachment should be updated in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, testAttachment.ID)
	suite.NoError(err)
	suite.True(dbAttachment.Cached)

	// the file should be back in storage at the same path as before
	refreshedBytes, err := suite.storage.Get(testAttachment.File.Path)
	suite.NoError(err)
	suite.Equal(suite.testRemoteAttachments[testAttachment.RemoteURL].Data, refreshedBytes)
}

func (suite *GetFileTestSuite) TestGetRemoteFileUncachedInterrupted() {
	ctx := context.Background()

	// uncache the file from local
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	testAttachment.Cached = false
	err := suite.db.UpdateByPrimaryKey(ctx, testAttachment)
	suite.NoError(err)
	err = suite.storage.Delete(testAttachment.File.Path)
	suite.NoError(err)
	err = suite.storage.Delete(testAttachment.Thumbnail.Path)
	suite.NoError(err)

	// now fetch it
	fileName := path.Base(testAttachment.File.Path)
	requestingAccount := suite.testAccounts["local_account_1"]

	content, errWithCode := suite.mediaProcessor.GetFile(ctx, requestingAccount, &apimodel.GetContentRequestForm{
		AccountID: testAttachment.AccountID,
		MediaType: string(media.TypeAttachment),
		MediaSize: string(media.SizeOriginal),
		FileName:  fileName,
	})

	suite.NoError(errWithCode)
	suite.NotNil(content)

	// only read the first kilobyte and then stop
	b := make([]byte, 1024)
	_, err = content.Content.Read(b)
	suite.NoError(err)

	// close the reader
	if closer, ok := content.Content.(io.Closer); ok {
		suite.NoError(closer.Close())
	}

	time.Sleep(2 * time.Second) // wait a few seconds for the media manager to finish doing stuff

	// the attachment should still be updated in the database even though the caller hung up
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, testAttachment.ID)
	suite.NoError(err)
	suite.True(dbAttachment.Cached)

	// the file should be back in storage at the same path as before
	refreshedBytes, err := suite.storage.Get(testAttachment.File.Path)
	suite.NoError(err)
	suite.Equal(suite.testRemoteAttachments[testAttachment.RemoteURL].Data, refreshedBytes)
}

func (suite *GetFileTestSuite) TestGetRemoteFileThumbnailUncached() {
	ctx := context.Background()
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// fetch the existing thumbnail bytes from storage first
	thumbnailBytes, err := suite.storage.Get(testAttachment.Thumbnail.Path)
	suite.NoError(err)

	// uncache the file from local
	testAttachment.Cached = false
	err = suite.db.UpdateByPrimaryKey(ctx, testAttachment)
	suite.NoError(err)
	err = suite.storage.Delete(testAttachment.File.Path)
	suite.NoError(err)
	err = suite.storage.Delete(testAttachment.Thumbnail.Path)
	suite.NoError(err)

	// now fetch the thumbnail
	fileName := path.Base(testAttachment.File.Path)
	requestingAccount := suite.testAccounts["local_account_1"]

	content, errWithCode := suite.mediaProcessor.GetFile(ctx, requestingAccount, &apimodel.GetContentRequestForm{
		AccountID: testAttachment.AccountID,
		MediaType: string(media.TypeAttachment),
		MediaSize: string(media.SizeSmall),
		FileName:  fileName,
	})

	suite.NoError(errWithCode)
	suite.NotNil(content)
	b, err := io.ReadAll(content.Content)
	suite.NoError(err)

	if closer, ok := content.Content.(io.Closer); ok {
		suite.NoError(closer.Close())
	}

	suite.Equal(thumbnailBytes, b)
	suite.Equal("image/jpeg", content.ContentType)
	suite.EqualValues(testAttachment.Thumbnail.FileSize, content.ContentLength)
}

func TestGetFileTestSuite(t *testing.T) {
	suite.Run(t, &GetFileTestSuite{})
}
