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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

type PruneUnusedLocalTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneUnusedLocalTestSuite) TestPruneUnusedLocal() {
	testAttachment := suite.testAttachments["local_account_1_unattached_1"]
	suite.True(*testAttachment.Cached)

	totalPruned, err := suite.manager.PruneUnusedLocalAttachments(context.Background())
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	_, err = suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *PruneUnusedLocalTestSuite) TestPruneRemoteTwice() {
	totalPruned, err := suite.manager.PruneUnusedLocalAttachments(context.Background())
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPrunedAgain, err := suite.manager.PruneUnusedLocalAttachments(context.Background())
	suite.NoError(err)
	suite.Equal(0, totalPrunedAgain)
}

func (suite *PruneUnusedLocalTestSuite) TestPruneOneNonExistent() {
	ctx := context.Background()
	testAttachment := suite.testAttachments["local_account_1_unattached_1"]

	// Delete this attachment cached on disk
	media, err := suite.db.GetAttachmentByID(ctx, testAttachment.ID)
	suite.NoError(err)
	suite.True(*media.Cached)
	err = suite.storage.Delete(ctx, media.File.Path)
	suite.NoError(err)

	// Now attempt to prune for item with db entry no file
	totalPruned, err := suite.manager.PruneUnusedLocalAttachments(ctx)
	suite.NoError(err)
	suite.Equal(1, totalPruned)
}

func TestPruneUnusedLocalTestSuite(t *testing.T) {
	suite.Run(t, &PruneUnusedLocalTestSuite{})
}
