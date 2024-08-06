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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type InteractionTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *InteractionTestSuite) markIntsPending(
	ctx context.Context,
	statusID string,
) (pendingCount int) {
	// Get replies of given status.
	intReplies, err := suite.state.DB.GetStatusReplies(ctx, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	// Mark each reply as pending approval.
	for _, intReply := range intReplies {
		intReply.PendingApproval = util.Ptr(true)
		if err := suite.state.DB.UpdateStatus(
			ctx,
			intReply,
			"pending_approval",
		); err != nil {
			suite.FailNow(err.Error())
		}

		pendingCount++
	}

	// Get boosts of given status.
	intBoosts, err := suite.state.DB.GetStatusBoosts(ctx, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	// Mark each boost as pending approval.
	for _, intBoost := range intBoosts {
		intBoost.PendingApproval = util.Ptr(true)
		if err := suite.state.DB.UpdateStatus(
			ctx,
			intBoost,
			"pending_approval",
		); err != nil {
			suite.FailNow(err.Error())
		}

		pendingCount++
	}

	// Get faves of given status.
	intFaves, err := suite.state.DB.GetStatusFaves(ctx, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	// Mark each fave as pending approval.
	for _, intFave := range intFaves {
		intFave.PendingApproval = util.Ptr(true)
		if err := suite.state.DB.UpdateStatusFave(
			ctx,
			intFave,
			"pending_approval",
		); err != nil {
			suite.FailNow(err.Error())
		}

		pendingCount++
	}

	return pendingCount
}

func (suite *InteractionTestSuite) TestGetPending() {
	var (
		testStatus = suite.testStatuses["local_account_1_status_1"]
		ctx        = context.Background()
		acctID     = suite.testAccounts["local_account_1"].ID
		statusID   = ""
		likes      = true
		replies    = true
		boosts     = true
		page       = &paging.Page{
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		}
	)

	// Update target test status to mark
	// all interactions with it pending.
	pendingCount := suite.markIntsPending(ctx, testStatus.ID)

	// Get pendingInts interactions.
	pendingInts, err := suite.state.DB.GetPendingInteractionsForAcct(
		ctx,
		acctID,
		statusID,
		likes,
		replies,
		boosts,
		page,
	)
	suite.NoError(err)
	suite.Len(pendingInts, pendingCount)

	// Ensure relevant model populated.
	for _, pendingInt := range pendingInts {
		switch pendingInt.InteractionType {

		case gtsmodel.InteractionLike:
			suite.NotNil(pendingInt.Like)

		case gtsmodel.InteractionReply:
			suite.NotNil(pendingInt.Reply)

		case gtsmodel.InteractionAnnounce:
			suite.NotNil(pendingInt.Announce)
		}
	}
}

func (suite *InteractionTestSuite) TestGetPendingRepliesOnly() {
	var (
		testStatus = suite.testStatuses["local_account_1_status_1"]
		ctx        = context.Background()
		acctID     = suite.testAccounts["local_account_1"].ID
		statusID   = ""
		likes      = false
		replies    = true
		boosts     = false
		page       = &paging.Page{
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		}
	)

	// Update target test status to mark
	// all interactions with it pending.
	suite.markIntsPending(ctx, testStatus.ID)

	// Get pendingInts interactions.
	pendingInts, err := suite.state.DB.GetPendingInteractionsForAcct(
		ctx,
		acctID,
		statusID,
		likes,
		replies,
		boosts,
		page,
	)
	suite.NoError(err)

	// Ensure only replies returned.
	for _, pendingInt := range pendingInts {
		suite.Equal(gtsmodel.InteractionReply, pendingInt.InteractionType)
	}
}

func TestInteractionTestSuite(t *testing.T) {
	suite.Run(t, new(InteractionTestSuite))
}
