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
	"io"
	"os"
	"testing"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/stretchr/testify/suite"
)

type PruneRemoteTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneRemoteTestSuite) TestPruneRemote() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testAttachment.Cached)

	totalPruned, err := suite.manager.PruneAllRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	prunedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)

	// the media should no longer be cached
	suite.False(*prunedAttachment.Cached)
}

func (suite *PruneRemoteTestSuite) TestPruneRemoteTwice() {
	totalPruned, err := suite.manager.PruneAllRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPrunedAgain, err := suite.manager.PruneAllRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(0, totalPrunedAgain)
}

func (suite *PruneRemoteTestSuite) TestPruneAndRecache() {
	ctx := context.Background()
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]

	totalPruned, err := suite.manager.PruneAllRemote(ctx, 1)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	// media should no longer be stored
	_, err = suite.storage.Get(ctx, testAttachment.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, testAttachment.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)

	// now recache the image....
	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("../../testrig/media/thoughtsofdog-original.jpeg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}
	processingRecache, err := suite.manager.RecacheMedia(ctx, data, nil, testAttachment.ID)
	suite.NoError(err)

	// synchronously load the recached attachment
	recachedAttachment, err := processingRecache.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(recachedAttachment)

	// recachedAttachment should be basically the same as the old attachment
	suite.True(*recachedAttachment.Cached)
	suite.Equal(testAttachment.ID, recachedAttachment.ID)
	suite.Equal(testAttachment.File.Path, recachedAttachment.File.Path)           // file should be stored in the same place
	suite.Equal(testAttachment.Thumbnail.Path, recachedAttachment.Thumbnail.Path) // as should the thumbnail
	suite.EqualValues(testAttachment.FileMeta, recachedAttachment.FileMeta)       // and the filemeta should be the same

	// recached files should be back in storage
	_, err = suite.storage.Get(ctx, recachedAttachment.File.Path)
	suite.NoError(err)
	_, err = suite.storage.Get(ctx, recachedAttachment.Thumbnail.Path)
	suite.NoError(err)
}

func (suite *PruneRemoteTestSuite) TestPruneOneNonExistent() {
	ctx := context.Background()
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// Delete this attachment cached on disk
	media, err := suite.db.GetAttachmentByID(ctx, testAttachment.ID)
	suite.NoError(err)
	suite.True(*media.Cached)
	err = suite.storage.Delete(ctx, media.File.Path)
	suite.NoError(err)

	// Now attempt to prune remote for item with db entry no file
	totalPruned, err := suite.manager.PruneAllRemote(ctx, 1)
	suite.NoError(err)
	suite.Equal(2, totalPruned)
}

func TestPruneRemoteTestSuite(t *testing.T) {
	suite.Run(t, &PruneRemoteTestSuite{})
}
