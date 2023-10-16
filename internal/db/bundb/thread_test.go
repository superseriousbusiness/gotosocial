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

package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type ThreadTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *ThreadTestSuite) TestGetThread() {
	testThread := suite.testThreads["local_account_1_status_1"]

	thread, err := suite.db.GetThread(context.Background(), testThread.ID)
	suite.NoError(err)
	suite.NotNil(thread)
	suite.Len(thread.StatusIDs, 3)
}

func (suite *ThreadTestSuite) TestPutThread() {
	suite.NoError(
		suite.db.PutThread(
			context.Background(),
			&gtsmodel.Thread{
				ID: "01HCWK4HVQ4VGSS1G4VQP3AXZF",
			},
		),
	)
}

func (suite *ThreadTestSuite) TestDeleteThread() {
	threadID := suite.testThreads["local_account_1_status_1"].ID
	ctx := context.Background()

	// Get thread and populate status IDs.
	thread, err := suite.db.GetThread(ctx, threadID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Select all the statuses in the thread so
	// they're cached, and ensure threaded correctly.
	statuses, err := suite.db.GetStatusesByIDs(ctx, thread.StatusIDs)
	if err != nil {
		suite.FailNow(err.Error())
	}

	for _, s := range statuses {
		if s.ThreadID != threadID {
			suite.FailNow("", "status %s should have had threadID %s", s.ID, threadID)
		}
	}

	// Delete thread.
	if err := suite.db.DeleteThread(ctx, threadID); err != nil {
		suite.FailNow(err.Error())
	}

	// For each status that was in the
	// thread, ensure threadID is now empty.
	statuses, err = suite.db.GetStatusesByIDs(ctx, thread.StatusIDs)
	if err != nil {
		suite.FailNow(err.Error())
	}

	for _, s := range statuses {
		if s.ThreadID != "" {
			suite.FailNow("", "status %s should have had no threadID", s.ID)
		}
	}
}

func TestThreadTestSuite(t *testing.T) {
	suite.Run(t, new(ThreadTestSuite))
}
