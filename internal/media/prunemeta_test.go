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
	"testing"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

type PruneMetaTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneMetaTestSuite) TestPruneMeta() {
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

	totalPruned, err := suite.manager.PruneAllMeta(ctx)
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

func (suite *PruneMetaTestSuite) TestPruneMetaTwice() {
	ctx := context.Background()

	// start by clearing zork's avatar + header
	zork := suite.testAccounts["local_account_1"]
	zork.AvatarMediaAttachmentID = ""
	zork.HeaderMediaAttachmentID = ""
	if err := suite.db.UpdateByID(ctx, zork, zork.ID, "avatar_media_attachment_id", "header_media_attachment_id"); err != nil {
		panic(err)
	}

	totalPruned, err := suite.manager.PruneAllMeta(ctx)
	suite.NoError(err)
	suite.Equal(2, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPruned, err = suite.manager.PruneAllMeta(ctx)
	suite.NoError(err)
	suite.Equal(0, totalPruned)
}

func (suite *PruneMetaTestSuite) TestPruneMetaMultipleAccounts() {
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

	totalPruned, err := suite.manager.PruneAllMeta(ctx)
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

func TestPruneMetaTestSuite(t *testing.T) {
	suite.Run(t, &PruneMetaTestSuite{})
}
