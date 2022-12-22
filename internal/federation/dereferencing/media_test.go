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

package dereferencing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AttachmentTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *AttachmentTestSuite) TestDereferenceAttachmentBlocking() {
	ctx := context.Background()

	fetchingAccount := suite.testAccounts["local_account_1"]

	attachmentOwner := "01FENS9F666SEQ6TYQWEEY78GM"
	attachmentStatus := "01FENS9NTTVNEX1YZV7GB63MT8"
	attachmentContentType := "image/jpeg"
	attachmentURL := "https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg"
	attachmentDescription := "It's a cute plushie."
	attachmentBlurhash := "LwP?p=aK_4%N%MRjWXt7%hozM_a}"

	media, err := suite.dereferencer.GetRemoteMedia(ctx, fetchingAccount.Username, attachmentOwner, attachmentURL, &media.AdditionalMediaInfo{
		StatusID:    &attachmentStatus,
		RemoteURL:   &attachmentURL,
		Description: &attachmentDescription,
		Blurhash:    &attachmentBlurhash,
	})
	suite.NoError(err)

	// make a blocking call to load the attachment from the in-process media
	attachment, err := media.LoadAttachment(ctx)
	suite.NoError(err)

	suite.NotNil(attachment)

	suite.Equal(attachmentOwner, attachment.AccountID)
	suite.Equal(attachmentStatus, attachment.StatusID)
	suite.Equal(attachmentURL, attachment.RemoteURL)
	suite.NotEmpty(attachment.URL)
	suite.NotEmpty(attachment.Blurhash)
	suite.NotEmpty(attachment.ID)
	suite.NotEmpty(attachment.CreatedAt)
	suite.NotEmpty(attachment.UpdatedAt)
	suite.EqualValues(1.3365462, attachment.FileMeta.Original.Aspect)
	suite.Equal(2071680, attachment.FileMeta.Original.Size)
	suite.Equal(1245, attachment.FileMeta.Original.Height)
	suite.Equal(1664, attachment.FileMeta.Original.Width)
	suite.Equal(attachmentBlurhash, attachment.Blurhash)
	suite.Equal(gtsmodel.ProcessingStatusProcessed, attachment.Processing)
	suite.NotEmpty(attachment.File.Path)
	suite.Equal(attachmentContentType, attachment.File.ContentType)
	suite.Equal(attachmentDescription, attachment.Description)

	suite.NotEmpty(attachment.Thumbnail.Path)
	suite.NotEmpty(attachment.Type)

	// attachment should also now be in the database
	dbAttachment, err := suite.db.GetAttachmentByID(context.Background(), attachment.ID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	suite.Equal(attachmentOwner, dbAttachment.AccountID)
	suite.Equal(attachmentStatus, dbAttachment.StatusID)
	suite.Equal(attachmentURL, dbAttachment.RemoteURL)
	suite.NotEmpty(dbAttachment.URL)
	suite.NotEmpty(dbAttachment.Blurhash)
	suite.NotEmpty(dbAttachment.ID)
	suite.NotEmpty(dbAttachment.CreatedAt)
	suite.NotEmpty(dbAttachment.UpdatedAt)
	suite.EqualValues(1.3365462, dbAttachment.FileMeta.Original.Aspect)
	suite.Equal(2071680, dbAttachment.FileMeta.Original.Size)
	suite.Equal(1245, dbAttachment.FileMeta.Original.Height)
	suite.Equal(1664, dbAttachment.FileMeta.Original.Width)
	suite.Equal(attachmentBlurhash, dbAttachment.Blurhash)
	suite.Equal(gtsmodel.ProcessingStatusProcessed, dbAttachment.Processing)
	suite.NotEmpty(dbAttachment.File.Path)
	suite.Equal(attachmentContentType, dbAttachment.File.ContentType)
	suite.Equal(attachmentDescription, dbAttachment.Description)

	suite.NotEmpty(dbAttachment.Thumbnail.Path)
	suite.NotEmpty(dbAttachment.Type)
}

func (suite *AttachmentTestSuite) TestDereferenceAttachmentAsync() {
	ctx := context.Background()

	fetchingAccount := suite.testAccounts["local_account_1"]

	attachmentOwner := "01FENS9F666SEQ6TYQWEEY78GM"
	attachmentStatus := "01FENS9NTTVNEX1YZV7GB63MT8"
	attachmentContentType := "image/jpeg"
	attachmentURL := "https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg"
	attachmentDescription := "It's a cute plushie."
	attachmentBlurhash := "LwP?p=aK_4%N%MRjWXt7%hozM_a}"

	processingMedia, err := suite.dereferencer.GetRemoteMedia(ctx, fetchingAccount.Username, attachmentOwner, attachmentURL, &media.AdditionalMediaInfo{
		StatusID:    &attachmentStatus,
		RemoteURL:   &attachmentURL,
		Description: &attachmentDescription,
		Blurhash:    &attachmentBlurhash,
	})
	suite.NoError(err)
	attachmentID := processingMedia.AttachmentID()

	if !testrig.WaitFor(func() bool {
		return processingMedia.Finished()
	}) {
		suite.FailNow("timed out waiting for media to be processed")
	}

	// now get the attachment from the database
	attachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)

	suite.NotNil(attachment)

	suite.Equal(attachmentOwner, attachment.AccountID)
	suite.Equal(attachmentStatus, attachment.StatusID)
	suite.Equal(attachmentURL, attachment.RemoteURL)
	suite.NotEmpty(attachment.URL)
	suite.NotEmpty(attachment.Blurhash)
	suite.NotEmpty(attachment.ID)
	suite.NotEmpty(attachment.CreatedAt)
	suite.NotEmpty(attachment.UpdatedAt)
	suite.EqualValues(1.3365462, attachment.FileMeta.Original.Aspect)
	suite.Equal(2071680, attachment.FileMeta.Original.Size)
	suite.Equal(1245, attachment.FileMeta.Original.Height)
	suite.Equal(1664, attachment.FileMeta.Original.Width)
	suite.Equal(attachmentBlurhash, attachment.Blurhash)
	suite.Equal(gtsmodel.ProcessingStatusProcessed, attachment.Processing)
	suite.NotEmpty(attachment.File.Path)
	suite.Equal(attachmentContentType, attachment.File.ContentType)
	suite.Equal(attachmentDescription, attachment.Description)

	suite.NotEmpty(attachment.Thumbnail.Path)
	suite.NotEmpty(attachment.Type)
}

func TestAttachmentTestSuite(t *testing.T) {
	suite.Run(t, new(AttachmentTestSuite))
}
