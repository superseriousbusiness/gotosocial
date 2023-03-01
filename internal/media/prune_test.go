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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type PruneTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneTestSuite) TestPruneOrphanedDry() {
	// add a big orphan panda to store
	b, err := os.ReadFile("./test/big-panda.gif")
	if err != nil {
		suite.FailNow(err.Error())
	}

	pandaPath := "01GJQJ1YD9QCHCE12GG0EYHVNW/attachment/original/01GJQJ2AYM1VKSRW96YVAJ3NK3.gif"
	if _, err := suite.storage.Put(context.Background(), pandaPath, b); err != nil {
		suite.FailNow(err.Error())
	}

	// dry run should show up 1 orphaned panda
	totalPruned, err := suite.manager.PruneOrphaned(context.Background(), true)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	// panda should still be in storage
	hasKey, err := suite.storage.Has(context.Background(), pandaPath)
	suite.NoError(err)
	suite.True(hasKey)
}

func (suite *PruneTestSuite) TestPruneOrphanedMoist() {
	// add a big orphan panda to store
	b, err := os.ReadFile("./test/big-panda.gif")
	if err != nil {
		suite.FailNow(err.Error())
	}

	pandaPath := "01GJQJ1YD9QCHCE12GG0EYHVNW/attachment/original/01GJQJ2AYM1VKSRW96YVAJ3NK3.gif"
	if _, err := suite.storage.Put(context.Background(), pandaPath, b); err != nil {
		suite.FailNow(err.Error())
	}

	// should show up 1 orphaned panda
	totalPruned, err := suite.manager.PruneOrphaned(context.Background(), false)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	// panda should no longer be in storage
	hasKey, err := suite.storage.Has(context.Background(), pandaPath)
	suite.NoError(err)
	suite.False(hasKey)
}

func (suite *PruneTestSuite) TestPruneUnusedLocal() {
	testAttachment := suite.testAttachments["local_account_1_unattached_1"]
	suite.True(*testAttachment.Cached)

	totalPruned, err := suite.manager.PruneUnusedLocal(context.Background(), false)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	_, err = suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *PruneTestSuite) TestPruneUnusedLocalDry() {
	testAttachment := suite.testAttachments["local_account_1_unattached_1"]
	suite.True(*testAttachment.Cached)

	totalPruned, err := suite.manager.PruneUnusedLocal(context.Background(), true)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	_, err = suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)
}

func (suite *PruneTestSuite) TestPruneRemoteTwice() {
	totalPruned, err := suite.manager.PruneUnusedLocal(context.Background(), false)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPrunedAgain, err := suite.manager.PruneUnusedLocal(context.Background(), false)
	suite.NoError(err)
	suite.Equal(0, totalPrunedAgain)
}

func (suite *PruneTestSuite) TestPruneOneNonExistent() {
	ctx := context.Background()
	testAttachment := suite.testAttachments["local_account_1_unattached_1"]

	// Delete this attachment cached on disk
	media, err := suite.db.GetAttachmentByID(ctx, testAttachment.ID)
	suite.NoError(err)
	suite.True(*media.Cached)
	err = suite.storage.Delete(ctx, media.File.Path)
	suite.NoError(err)

	// Now attempt to prune for item with db entry no file
	totalPruned, err := suite.manager.PruneUnusedLocal(ctx, false)
	suite.NoError(err)
	suite.Equal(1, totalPruned)
}

func (suite *PruneTestSuite) TestPruneUnusedRemote() {
	ctx := context.Background()

	// start by clearing zork's avatar + header
	zorkOldAvatar := suite.testAttachments["local_account_1_avatar"]
	zorkOldHeader := suite.testAttachments["local_account_1_avatar"]
	zork := suite.testAccounts["local_account_1"]
	zork.AvatarMediaAttachmentID = ""
	zork.HeaderMediaAttachmentID = ""
	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
		panic(err)
	}

	totalPruned, err := suite.manager.PruneUnusedRemote(ctx, false)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	// media should no longer be stored
	_, err = suite.storage.Get(ctx, zorkOldAvatar.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, zorkOldAvatar.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, zorkOldHeader.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, zorkOldHeader.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)

	// attachments should no longer be in the db
	_, err = suite.db.GetAttachmentByID(ctx, zorkOldAvatar.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	_, err = suite.db.GetAttachmentByID(ctx, zorkOldHeader.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *PruneTestSuite) TestPruneUnusedRemoteTwice() {
	ctx := context.Background()

	// start by clearing zork's avatar + header
	zork := suite.testAccounts["local_account_1"]
	zork.AvatarMediaAttachmentID = ""
	zork.HeaderMediaAttachmentID = ""
	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
		panic(err)
	}

	totalPruned, err := suite.manager.PruneUnusedRemote(ctx, false)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPruned, err = suite.manager.PruneUnusedRemote(ctx, false)
	suite.NoError(err)
	suite.Equal(0, totalPruned)
}

func (suite *PruneTestSuite) TestPruneUnusedRemoteMultipleAccounts() {
	ctx := context.Background()

	// start by clearing zork's avatar + header
	zorkOldAvatar := suite.testAttachments["local_account_1_avatar"]
	zorkOldHeader := suite.testAttachments["local_account_1_avatar"]
	zork := suite.testAccounts["local_account_1"]
	zork.AvatarMediaAttachmentID = ""
	zork.HeaderMediaAttachmentID = ""
	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
		panic(err)
	}

	// set zork's unused header as belonging to turtle
	turtle := suite.testAccounts["local_account_1"]
	zorkOldHeader.AccountID = turtle.ID
	if err := suite.db.UpdateByID(ctx, zorkOldHeader, zorkOldHeader.ID, "account_id"); err != nil {
		panic(err)
	}

	totalPruned, err := suite.manager.PruneUnusedRemote(ctx, false)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	// media should no longer be stored
	_, err = suite.storage.Get(ctx, zorkOldAvatar.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, zorkOldAvatar.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, zorkOldHeader.File.Path)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, zorkOldHeader.Thumbnail.Path)
	suite.ErrorIs(err, storage.ErrNotFound)

	// attachments should no longer be in the db
	_, err = suite.db.GetAttachmentByID(ctx, zorkOldAvatar.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	_, err = suite.db.GetAttachmentByID(ctx, zorkOldHeader.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *PruneTestSuite) TestUncacheRemote() {
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testStatusAttachment.Cached)

	testHeader := suite.testAttachments["remote_account_3_header"]
	suite.True(*testHeader.Cached)

	totalUncached, err := suite.manager.UncacheRemote(context.Background(), 1, false)
	suite.NoError(err)
	suite.Equal(2, totalUncached)

	uncachedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testStatusAttachment.ID)
	suite.NoError(err)
	suite.False(*uncachedAttachment.Cached)

	uncachedAttachment, err = suite.db.GetAttachmentByID(context.Background(), testHeader.ID)
	suite.NoError(err)
	suite.False(*uncachedAttachment.Cached)
}

func (suite *PruneTestSuite) TestUncacheRemoteDry() {
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testStatusAttachment.Cached)

	testHeader := suite.testAttachments["remote_account_3_header"]
	suite.True(*testHeader.Cached)

	totalUncached, err := suite.manager.UncacheRemote(context.Background(), 1, true)
	suite.NoError(err)
	suite.Equal(2, totalUncached)

	uncachedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testStatusAttachment.ID)
	suite.NoError(err)
	suite.True(*uncachedAttachment.Cached)

	uncachedAttachment, err = suite.db.GetAttachmentByID(context.Background(), testHeader.ID)
	suite.NoError(err)
	suite.True(*uncachedAttachment.Cached)
}

func (suite *PruneTestSuite) TestUncacheRemoteTwice() {
	totalUncached, err := suite.manager.UncacheRemote(context.Background(), 1, false)
	suite.NoError(err)
	suite.Equal(2, totalUncached)

	// final uncache should uncache nothing, since the first uncache already happened
	totalUncachedAgain, err := suite.manager.UncacheRemote(context.Background(), 1, false)
	suite.NoError(err)
	suite.Equal(0, totalUncachedAgain)
}

func (suite *PruneTestSuite) TestUncacheAndRecache() {
	ctx := context.Background()
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	testHeader := suite.testAttachments["remote_account_3_header"]

	totalUncached, err := suite.manager.UncacheRemote(ctx, 1, false)
	suite.NoError(err)
	suite.Equal(2, totalUncached)

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
		processingRecache, err := suite.manager.PreProcessMediaRecache(ctx, data, nil, original.ID)
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

func (suite *PruneTestSuite) TestUncacheOneNonExistent() {
	ctx := context.Background()
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// Delete this attachment cached on disk
	media, err := suite.db.GetAttachmentByID(ctx, testStatusAttachment.ID)
	suite.NoError(err)
	suite.True(*media.Cached)
	err = suite.storage.Delete(ctx, media.File.Path)
	suite.NoError(err)

	// Now attempt to uncache remote for item with db entry no file
	totalUncached, err := suite.manager.UncacheRemote(ctx, 1, false)
	suite.NoError(err)
	suite.Equal(2, totalUncached)
}

func TestPruneOrphanedTestSuite(t *testing.T) {
	suite.Run(t, &PruneTestSuite{})
}
