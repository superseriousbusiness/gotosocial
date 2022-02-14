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

	"github.com/stretchr/testify/suite"
)

type PruneRemoteTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneRemoteTestSuite) TestPruneRemote() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.NotEmpty(testAttachment.File.Path)
	suite.NotEmpty(testAttachment.Thumbnail.Path)

	totalPruned, err := suite.manager.PruneRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	prunedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)

	// the url and thumbnail paths should be cleared
	suite.Empty(prunedAttachment.File.Path)
	suite.Empty(prunedAttachment.Thumbnail.Path)
}

func (suite *PruneRemoteTestSuite) TestPruneRemoteTwice() {
	totalPruned, err := suite.manager.PruneRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(1, totalPruned)

	// final prune should prune nothing, since the first prune already happened
	totalPrunedAgain, err := suite.manager.PruneRemote(context.Background(), 1)
	suite.NoError(err)
	suite.Equal(0, totalPrunedAgain)
}

func TestPruneRemoteTestSuite(t *testing.T) {
	suite.Run(t, &PruneRemoteTestSuite{})
}
