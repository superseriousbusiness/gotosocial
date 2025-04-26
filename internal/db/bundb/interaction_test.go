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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type InteractionTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *InteractionTestSuite) markInteractionsPending(
	ctx context.Context,
	statusID string,
) (pendingCount int) {
	// Get replies of given status.
	replies, err := suite.state.DB.GetStatusReplies(ctx, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	// Mark each reply as pending approval.
	for _, reply := range replies {
		reply.PendingApproval = util.Ptr(true)
		if err := suite.state.DB.UpdateStatus(
			ctx,
			reply,
			"pending_approval",
		); err != nil {
			suite.FailNow(err.Error())
		}

		// Put an interaction request
		// in the DB for this reply.
		req := typeutils.StatusToInteractionRequest(reply)
		if err := suite.state.DB.PutInteractionRequest(ctx, req); err != nil {
			suite.FailNow(err.Error())
		}

		pendingCount++
	}

	// Get boosts of given status.
	boosts, err := suite.state.DB.GetStatusBoosts(ctx, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	// Mark each boost as pending approval.
	for _, boost := range boosts {
		boost.PendingApproval = util.Ptr(true)
		if err := suite.state.DB.UpdateStatus(
			ctx,
			boost,
			"pending_approval",
		); err != nil {
			suite.FailNow(err.Error())
		}

		// Put an interaction request
		// in the DB for this boost.
		req := typeutils.StatusToInteractionRequest(boost)
		if err := suite.state.DB.PutInteractionRequest(ctx, req); err != nil {
			suite.FailNow(err.Error())
		}

		pendingCount++
	}

	// Get faves of given status.
	faves, err := suite.state.DB.GetStatusFaves(ctx, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	// Mark each fave as pending approval.
	for _, fave := range faves {
		fave.PendingApproval = util.Ptr(true)
		if err := suite.state.DB.UpdateStatusFave(
			ctx,
			fave,
			"pending_approval",
		); err != nil {
			suite.FailNow(err.Error())
		}

		// Put an interaction request
		// in the DB for this fave.
		req := typeutils.StatusFaveToInteractionRequest(fave)
		if err := suite.state.DB.PutInteractionRequest(ctx, req); err != nil {
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
	pendingCount := suite.markInteractionsPending(ctx, testStatus.ID)

	// Get pendingInts interactions.
	pendingInts, err := suite.state.DB.GetInteractionsRequestsForAcct(
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
	suite.markInteractionsPending(ctx, testStatus.ID)

	// Get pendingInts interactions.
	pendingInts, err := suite.state.DB.GetInteractionsRequestsForAcct(
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

func (suite *InteractionTestSuite) TestInteractionRejected() {
	var (
		ctx = context.Background()
		req = new(gtsmodel.InteractionRequest)
	)

	// Make a copy of the request we'll modify.
	*req = *suite.testInteractionRequests["admin_account_reply_turtle"]

	// No rejection in the db for this interaction URI so it should be OK.
	rejected, err := suite.state.DB.IsInteractionRejected(ctx, req.InteractionURI)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if rejected {
		suite.FailNow("wanted rejected = false, got true")
	}

	// Update the interaction request to mark it rejected.
	req.RejectedAt = time.Now()
	req.URI = "https://some.reject.uri"
	if err := suite.state.DB.UpdateInteractionRequest(ctx, req, "uri", "rejected_at"); err != nil {
		suite.FailNow(err.Error())
	}

	// Rejection in the db for this interaction URI now so it should be très mauvais.
	rejected, err = suite.state.DB.IsInteractionRejected(ctx, req.InteractionURI)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if !rejected {
		suite.FailNow("wanted rejected = true, got false")
	}
}

func TestInteractionTestSuite(t *testing.T) {
	suite.Run(t, new(InteractionTestSuite))
}
