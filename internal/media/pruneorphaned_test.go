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
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PruneOrphanedTestSuite struct {
	MediaStandardTestSuite
}

func (suite *PruneOrphanedTestSuite) TestPruneOrphanedDry() {
	// add a big orphan panda to store
	b, err := os.ReadFile("./test/big-panda.gif")
	if err != nil {
		panic(err)
	}

	pandaPath := "01GJQJ1YD9QCHCE12GG0EYHVNW/attachments/original/01GJQJ2AYM1VKSRW96YVAJ3NK3.gif"
	if err := suite.storage.PutStream(context.Background(), pandaPath, bytes.NewBuffer(b)); err != nil {
		panic(err)
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

func (suite *PruneOrphanedTestSuite) TestPruneOrphanedMoist() {
	// add a big orphan panda to store
	b, err := os.ReadFile("./test/big-panda.gif")
	if err != nil {
		panic(err)
	}

	pandaPath := "01GJQJ1YD9QCHCE12GG0EYHVNW/attachments/original/01GJQJ2AYM1VKSRW96YVAJ3NK3.gif"
	if err := suite.storage.PutStream(context.Background(), pandaPath, bytes.NewBuffer(b)); err != nil {
		panic(err)
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

func TestPruneOrphanedTestSuite(t *testing.T) {
	suite.Run(t, &PruneOrphanedTestSuite{})
}
