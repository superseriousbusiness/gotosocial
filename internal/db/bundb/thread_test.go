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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type ThreadTestSuite struct {
	BunDBStandardTestSuite
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

func (suite *ThreadTestSuite) TestMuteUnmuteThread() {
	var (
		threadID   = suite.testThreads["local_account_1_status_1"].ID
		accountID  = suite.testAccounts["local_account_1"].ID
		ctx        = context.Background()
		threadMute = &gtsmodel.ThreadMute{
			ID:        "01HD3K14B62YJHH4RR0DSZ1EQ2",
			ThreadID:  threadID,
			AccountID: accountID,
		}
	)

	// Mute the thread and ensure it's actually muted.
	if err := suite.db.PutThreadMute(ctx, threadMute); err != nil {
		suite.FailNow(err.Error())
	}

	muted, err := suite.db.IsThreadMutedByAccount(ctx, threadID, accountID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !muted {
		suite.FailNow("", "expected thread %s to be muted by account %s", threadID, accountID)
	}

	_, err = suite.db.GetThreadMutedByAccount(ctx, threadID, accountID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Unmute the thread and ensure it's actually unmuted.
	if err := suite.db.DeleteThreadMute(ctx, threadMute.ID); err != nil {
		suite.FailNow(err.Error())
	}

	muted, err = suite.db.IsThreadMutedByAccount(ctx, threadID, accountID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if muted {
		suite.FailNow("", "expected thread %s to not be muted by account %s", threadID, accountID)
	}
}

func TestThreadTestSuite(t *testing.T) {
	suite.Run(t, new(ThreadTestSuite))
}
