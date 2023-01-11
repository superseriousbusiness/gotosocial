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
	"io"
	"os"
	"testing"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type PruneRemoteTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneRemoteTestSuite) TestPruneRemote() {
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testStatusAttachment.Cached)

	testHeader := suite.testAttachments["remote_account_3_header"]
	suite.True(*testHeader.Cached)

	totalPruned, err := suite.manager.PruneAllRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(3, totalPruned)

	prunedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testStatusAttachment.ID)
	suite.NoError(err)
	suite.False(*prunedAttachment.Cached)

	prunedAttachment, err = suite.db.GetAttachmentByID(context.Background(), testHeader.ID)
	suite.NoError(err)
	suite.False(*prunedAttachment.Cached)
}

func (suite *PruneRemoteTestSuite) TestPruneRemoteTwice() {
	totalPruned, err := suite.manager.PruneAllRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(3, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPrunedAgain, err := suite.manager.PruneAllRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(0, totalPrunedAgain)
}

func (suite *PruneRemoteTestSuite) TestPruneAndRecache() {
	ctx := context.Background()
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	testHeader := suite.testAttachments["remote_account_3_header"]

	totalPruned, err := suite.manager.PruneAllRemote(ctx, 1)
	suite.NoError(err)
	suite.Equal(3, totalPruned)

	// media should no longer be stored
	_, err = suite.storage.Get(ctx, testStatusAttachment.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, testStatusAttachment.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, testHeader.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, testHeader.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)

	// now recache the image....
	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("../../testrig/media/thoughtsofdog-original.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	for _, original := range []*gtsmodel.MediaAttachment{
		testStatusAttachment,
		testHeader,
	} {
		processingRecache, err := suite.manager.RecacheMedia(ctx, data, nil, original.ID)
		suite.NoError(err)

		// synchronously load the recached attachment
		recachedAttachment, err := processingRecache.LoadAttachment(ctx)
		suite.NoError(err)
		suite.NotNil(recachedAttachment)

		// recachedAttachment should be basically the same as the old attachment
		suite.True(*recachedAttachment.Cached)
		suite.Equal(original.ID, recachedAttachment.ID)
		suite.Equal(original.File.Path, recachedAttachment.File.Path)           // file should be stored in the same place
		suite.Equal(original.Thumbnail.Path, recachedAttachment.Thumbnail.Path) // as should the thumbnail
		suite.EqualValues(original.FileMeta, recachedAttachment.FileMeta)       // and the filemeta should be the same

		// recached files should be back in storage
		_, err = suite.storage.Get(ctx, recachedAttachment.File.Path)
		suite.NoError(err)
		_, err = suite.storage.Get(ctx, recachedAttachment.Thumbnail.Path)
		suite.NoError(err)
	}
}

func (suite *PruneRemoteTestSuite) TestPruneOneNonExistent() {
	ctx := context.Background()
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// Delete this attachment cached on disk
	media, err := suite.db.GetAttachmentByID(ctx, testStatusAttachment.ID)
	suite.NoError(err)
	suite.True(*media.Cached)
	err = suite.storage.Delete(ctx, media.File.Path)
	suite.NoError(err)

	// Now attempt to prune remote for item with db entry no file
	totalPruned, err := suite.manager.PruneAllRemote(ctx, 1)
	suite.NoError(err)
	suite.Equal(3, totalPruned)
}

func TestPruneRemoteTestSuite(t *testing.T) {
	suite.Run(t, &PruneRemoteTestSuite{})
}
