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
}

func TestGetFileTestSuite(t *testing.T) {
	suite.Run(t, &GetFileTestSuite{})
}
