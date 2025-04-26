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

package cleaner_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type MediaTestSuite struct {
	suite.Suite

	db                  db.DB
	storage             *storage.Driver
	state               state.State
	manager             *media.Manager
	cleaner             *cleaner.Cleaner
	transportController transport.Controller
	testAttachments     map[string]*gtsmodel.MediaAttachment
	testAccounts        map[string]*gtsmodel.Account
	testEmojis          map[string]*gtsmodel.Emoji
}

func TestMediaTestSuite(t *testing.T) {
	suite.Run(t, &MediaTestSuite{})
}

func (suite *MediaTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	suite.db = testrig.NewTestDB(&suite.state)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.state.Storage = suite.storage

	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
	testrig.StandardDBSetup(suite.db, nil)

	suite.testAttachments = testrig.NewTestAttachments()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testEmojis = testrig.NewTestEmojis()
	suite.manager = testrig.NewTestMediaManager(&suite.state)
	suite.cleaner = cleaner.New(&suite.state)
	suite.transportController = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../testrig/media"))
}

func (suite *MediaTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

// func (suite *MediaTestSuite) TestPruneOrphanedDry() {
// 	// add a big orphan panda to store
// 	b, err := os.ReadFile("../media/test/big-panda.gif")
// 	if err != nil {
// 		suite.FailNow(err.Error())
// 	}

// 	pandaPath := "01GJQJ1YD9QCHCE12GG0EYHVNW/attachment/original/01GJQJ2AYM1VKSRW96YVAJ3NK3.gif"
// 	if _, err := suite.storage.Put(context.Background(), pandaPath, b); err != nil {
// 		suite.FailNow(err.Error())
// 	}

// 	ctx := context.Background()

// 	// dry run should show up 1 orphaned panda
// 	totalPruned, err := suite.cleaner.Media().PruneOrphaned(gtscontext.SetDryRun(ctx))
// 	suite.NoError(err)
// 	suite.Equal(1, totalPruned)

// 	// panda should still be in storage
// 	hasKey, err := suite.storage.Has(ctx, pandaPath)
// 	suite.NoError(err)
// 	suite.True(hasKey)
// }

// func (suite *MediaTestSuite) TestPruneOrphanedMoist() {
// 	// i am not complicit in the moistness of this codebase :|

// 	// add a big orphan panda to store
// 	b, err := os.ReadFile("../media/test/big-panda.gif")
// 	if err != nil {
// 		suite.FailNow(err.Error())
// 	}

// 	pandaPath := "01GJQJ1YD9QCHCE12GG0EYHVNW/attachment/original/01GJQJ2AYM1VKSRW96YVAJ3NK3.gif"
// 	if _, err := suite.storage.Put(context.Background(), pandaPath, b); err != nil {
// 		suite.FailNow(err.Error())
// 	}

// 	ctx := context.Background()

// 	// should show up 1 orphaned panda
// 	totalPruned, err := suite.cleaner.Media().PruneOrphaned(ctx)
// 	suite.NoError(err)
// 	suite.Equal(1, totalPruned)

// 	// panda should no longer be in storage
// 	hasKey, err := suite.storage.Has(ctx, pandaPath)
// 	suite.NoError(err)
// 	suite.False(hasKey)
// }

// func (suite *MediaTestSuite) TestPruneUnusedLocal() {
// 	testAttachment := suite.testAttachments["local_account_1_unattached_1"]
// 	suite.True(*testAttachment.Cached)

// 	totalPruned, err := suite.manager.PruneUnusedLocal(context.Background(), false)
// 	suite.NoError(err)
// 	suite.Equal(1, totalPruned)

// 	_, err = suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
// 	suite.ErrorIs(err, db.ErrNoEntries)
// }

// func (suite *MediaTestSuite) TestPruneUnusedLocalDry() {
// 	testAttachment := suite.testAttachments["local_account_1_unattached_1"]
// 	suite.True(*testAttachment.Cached)

// 	totalPruned, err := suite.manager.PruneUnusedLocal(context.Background(), true)
// 	suite.NoError(err)
// 	suite.Equal(1, totalPruned)

// 	_, err = suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
// 	suite.NoError(err)
// }

// func (suite *MediaTestSuite) TestPruneRemoteTwice() {
// 	totalPruned, err := suite.manager.PruneUnusedLocal(context.Background(), false)
// 	suite.NoError(err)
// 	suite.Equal(1, totalPruned)

// 	// final prune should prune nothing, since the first prune already happened
// 	totalPrunedAgain, err := suite.manager.PruneUnusedLocal(context.Background(), false)
// 	suite.NoError(err)
// 	suite.Equal(0, totalPrunedAgain)
// }

// func (suite *MediaTestSuite) TestPruneOneNonExistent() {
// 	ctx := context.Background()
// 	testAttachment := suite.testAttachments["local_account_1_unattached_1"]

// 	// Delete this attachment cached on disk
// 	media, err := suite.db.GetAttachmentByID(ctx, testAttachment.ID)
// 	suite.NoError(err)
// 	suite.True(*media.Cached)
// 	err = suite.storage.Delete(ctx, media.File.Path)
// 	suite.NoError(err)

// 	// Now attempt to prune for item with db entry no file
// 	totalPruned, err := suite.manager.PruneUnusedLocal(ctx, false)
// 	suite.NoError(err)
// 	suite.Equal(1, totalPruned)
// }

// func (suite *MediaTestSuite) TestPruneUnusedRemote() {
// 	ctx := context.Background()

// 	// start by clearing zork's avatar + header
// 	zorkOldAvatar := suite.testAttachments["local_account_1_avatar"]
// 	zorkOldHeader := suite.testAttachments["local_account_1_avatar"]
// 	zork := suite.testAccounts["local_account_1"]
// 	zork.AvatarMediaAttachmentID = ""
// 	zork.HeaderMediaAttachmentID = ""
// 	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
// 		panic(err)
// 	}

// 	totalPruned, err := suite.manager.PruneUnusedRemote(ctx, false)
// 	suite.NoError(err)
// 	suite.Equal(2, totalPruned)

// 	// media should no longer be stored
// 	_, err = suite.storage.Get(ctx, zorkOldAvatar.File.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)
// 	_, err = suite.storage.Get(ctx, zorkOldAvatar.Thumbnail.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)
// 	_, err = suite.storage.Get(ctx, zorkOldHeader.File.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)
// 	_, err = suite.storage.Get(ctx, zorkOldHeader.Thumbnail.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)

// 	// attachments should no longer be in the db
// 	_, err = suite.db.GetAttachmentByID(ctx, zorkOldAvatar.ID)
// 	suite.ErrorIs(err, db.ErrNoEntries)
// 	_, err = suite.db.GetAttachmentByID(ctx, zorkOldHeader.ID)
// 	suite.ErrorIs(err, db.ErrNoEntries)
// }

// func (suite *MediaTestSuite) TestPruneUnusedRemoteTwice() {
// 	ctx := context.Background()

// 	// start by clearing zork's avatar + header
// 	zork := suite.testAccounts["local_account_1"]
// 	zork.AvatarMediaAttachmentID = ""
// 	zork.HeaderMediaAttachmentID = ""
// 	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
// 		panic(err)
// 	}

// 	totalPruned, err := suite.manager.PruneUnusedRemote(ctx, false)
// 	suite.NoError(err)
// 	suite.Equal(2, totalPruned)

// 	// final prune should prune nothing, since the first prune already happened
// 	totalPruned, err = suite.manager.PruneUnusedRemote(ctx, false)
// 	suite.NoError(err)
// 	suite.Equal(0, totalPruned)
// }

// func (suite *MediaTestSuite) TestPruneUnusedRemoteMultipleAccounts() {
// 	ctx := context.Background()

// 	// start by clearing zork's avatar + header
// 	zorkOldAvatar := suite.testAttachments["local_account_1_avatar"]
// 	zorkOldHeader := suite.testAttachments["local_account_1_avatar"]
// 	zork := suite.testAccounts["local_account_1"]
// 	zork.AvatarMediaAttachmentID = ""
// 	zork.HeaderMediaAttachmentID = ""
// 	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
// 		panic(err)
// 	}

// 	// set zork's unused header as belonging to turtle
// 	turtle := suite.testAccounts["local_account_1"]
// 	zorkOldHeader.AccountID = turtle.ID
// 	if err := suite.db.UpdateByID(ctx, zorkOldHeader, zorkOldHeader.ID, "account_id"); err != nil {
// 		panic(err)
// 	}

// 	totalPruned, err := suite.manager.PruneUnusedRemote(ctx, false)
// 	suite.NoError(err)
// 	suite.Equal(2, totalPruned)

// 	// media should no longer be stored
// 	_, err = suite.storage.Get(ctx, zorkOldAvatar.File.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)
// 	_, err = suite.storage.Get(ctx, zorkOldAvatar.Thumbnail.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)
// 	_, err = suite.storage.Get(ctx, zorkOldHeader.File.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)
// 	_, err = suite.storage.Get(ctx, zorkOldHeader.Thumbnail.Path)
// 	suite.ErrorIs(err, storage.ErrNotFound)

// 	// attachments should no longer be in the db
// 	_, err = suite.db.GetAttachmentByID(ctx, zorkOldAvatar.ID)
// 	suite.ErrorIs(err, db.ErrNoEntries)
// 	_, err = suite.db.GetAttachmentByID(ctx, zorkOldHeader.ID)
// 	suite.ErrorIs(err, db.ErrNoEntries)
// }

func (suite *MediaTestSuite) TestUncacheRemote() {
	ctx := context.Background()

	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testStatusAttachment.Cached)

	testHeader := suite.testAttachments["remote_account_3_header"]
	suite.True(*testHeader.Cached)

	after := time.Now().Add(-24 * time.Hour)
	totalUncached, err := suite.cleaner.Media().UncacheRemote(ctx, after)
	suite.NoError(err)
	suite.Equal(3, totalUncached)

	uncachedAttachment, err := suite.db.GetAttachmentByID(ctx, testStatusAttachment.ID)
	suite.NoError(err)
	suite.False(*uncachedAttachment.Cached)

	uncachedAttachment, err = suite.db.GetAttachmentByID(ctx, testHeader.ID)
	suite.NoError(err)
	suite.False(*uncachedAttachment.Cached)
}

func (suite *MediaTestSuite) TestUncacheRemoteDry() {
	ctx := context.Background()

	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testStatusAttachment.Cached)

	testHeader := suite.testAttachments["remote_account_3_header"]
	suite.True(*testHeader.Cached)

	after := time.Now().Add(-24 * time.Hour)
	totalUncached, err := suite.cleaner.Media().UncacheRemote(gtscontext.SetDryRun(ctx), after)
	suite.NoError(err)
	suite.Equal(3, totalUncached)

	uncachedAttachment, err := suite.db.GetAttachmentByID(ctx, testStatusAttachment.ID)
	suite.NoError(err)
	suite.True(*uncachedAttachment.Cached)

	uncachedAttachment, err = suite.db.GetAttachmentByID(ctx, testHeader.ID)
	suite.NoError(err)
	suite.True(*uncachedAttachment.Cached)
}

func (suite *MediaTestSuite) TestUncacheRemoteTwice() {
	ctx := context.Background()
	after := time.Now().Add(-24 * time.Hour)

	totalUncached, err := suite.cleaner.Media().UncacheRemote(ctx, after)
	suite.NoError(err)
	suite.Equal(3, totalUncached)

	// final uncache should uncache nothing, since the first uncache already happened
	totalUncachedAgain, err := suite.cleaner.Media().UncacheRemote(ctx, after)
	suite.NoError(err)
	suite.Equal(0, totalUncachedAgain)
}

func (suite *MediaTestSuite) TestUncacheAndRecache() {
	ctx := context.Background()
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	testHeader := suite.testAttachments["remote_account_3_header"]

	after := time.Now().Add(-24 * time.Hour)
	totalUncached, err := suite.cleaner.Media().UncacheRemote(ctx, after)
	suite.NoError(err)
	suite.Equal(3, totalUncached)

	// media should no longer be stored
	_, err = suite.storage.Get(ctx, testStatusAttachment.File.Path)
	suite.True(storage.IsNotFound(err))
	_, err = suite.storage.Get(ctx, testStatusAttachment.Thumbnail.Path)
	suite.True(storage.IsNotFound(err))
	_, err = suite.storage.Get(ctx, testHeader.File.Path)
	suite.True(storage.IsNotFound(err))
	_, err = suite.storage.Get(ctx, testHeader.Thumbnail.Path)
	suite.True(storage.IsNotFound(err))

	// now recache the image....
	data := func(_ context.Context) (io.ReadCloser, error) {
		// load bytes from a test image
		b, err := os.ReadFile("../../testrig/media/thoughtsofdog-original.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	for _, original := range []*gtsmodel.MediaAttachment{
		testStatusAttachment,
		testHeader,
	} {
		processing := suite.manager.CacheMedia(original, data)

		// synchronously load the recached attachment
		recachedAttachment, err := processing.Load(ctx)
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

func (suite *MediaTestSuite) TestUncacheOneNonExistent() {
	ctx := context.Background()
	testStatusAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// Delete this attachment cached on disk
	media, err := suite.db.GetAttachmentByID(ctx, testStatusAttachment.ID)
	suite.NoError(err)
	suite.True(*media.Cached)
	err = suite.storage.Delete(ctx, media.File.Path)
	suite.NoError(err)

	// Now attempt to uncache remote for item with db entry no file
	after := time.Now().Add(-24 * time.Hour)
	totalUncached, err := suite.cleaner.Media().UncacheRemote(ctx, after)
	suite.NoError(err)
	suite.Equal(3, totalUncached)
}
