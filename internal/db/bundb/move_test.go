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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type MoveTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *MoveTestSuite) TestMoveIntegration() {
	ctx := context.Background()
	firstMove := &gtsmodel.Move{
		ID:        "01HPPN38MZYEC6WBTR21J6241N",
		OriginURI: "https://example.org/users/my_old_account",
		TargetURI: "https://somewhere.else.net/users/my_new_account",
		URI:       "https://example.org/users/my_old_account/activities/Move/652e8361-0182-407d-8b01-4447e7fd10c0",
	}

	// Put the move.
	if err := suite.state.DB.PutMove(ctx, firstMove); err != nil {
		suite.FailNow(err.Error())
	}

	// Test various ways of retrieving the Move.
	if _, err := suite.state.DB.GetMoveByID(ctx, firstMove.ID); err != nil {
		suite.FailNow(err.Error())
	}

	if _, err := suite.state.DB.GetMoveByOriginTarget(ctx, firstMove.OriginURI, firstMove.TargetURI); err != nil {
		suite.FailNow(err.Error())
	}

	// Keep the last one, and check fields set on it.
	dbMove, err := suite.state.DB.GetMoveByURI(ctx, firstMove.URI)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Created/Updated should be set when
	// it's first inserted into the db.
	suite.NotZero(dbMove.CreatedAt)
	suite.NotZero(dbMove.UpdatedAt)

	// URIs should be parsed and set
	// on the move on population.
	suite.NotNil(dbMove.Origin)
	suite.NotNil(dbMove.Target)

	// These should not be set as
	// they have no default values.
	suite.Zero(dbMove.AttemptedAt)
	suite.Zero(dbMove.SucceededAt)

	// Update the Move to emulate
	// us succeeding in processing it.
	dbMove.AttemptedAt = time.Now()
	dbMove.SucceededAt = dbMove.AttemptedAt
	if err := suite.state.DB.UpdateMove(
		ctx,
		dbMove,
		"attempted_at",
		"succeeded_at",
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Store dbMove as firstMove var.
	firstMove = dbMove

	// Store another Move involving one
	// of the original URIs, and mark
	// this one as succeeded. Use a time
	// a few seconds into the future to
	// make sure it's differentiated
	// from the first move.
	secondMove := &gtsmodel.Move{
		ID:          "01HPPPNQWRMQTXRFEPKDV3A4W7",
		OriginURI:   "https://somewhere.else.net/users/my_new_account",
		TargetURI:   "http://localhost:8080/users/the_mighty_zork",
		URI:         "https://somewhere.else.net/activities/01HPPPPPC089VJGV0967P5YQS5",
		AttemptedAt: time.Now().Add(5 * time.Second),
		SucceededAt: time.Now().Add(5 * time.Second),
	}
	if err := suite.state.DB.PutMove(ctx, secondMove); err != nil {
		suite.FailNow(err.Error())
	}

	// Test getting succeeded using the
	// URI shared between the two Moves,
	// and some random account.
	ts, err := suite.state.DB.GetLatestMoveSuccessInvolvingURIs(
		ctx,
		secondMove.OriginURI,
		"https://a.secret.third.place/users/mystery_meat",
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Time should be equivalent to secondMove.
	suite.EqualValues(secondMove.SucceededAt.UnixMilli(), ts.UnixMilli())

	// Test getting succeeded using
	// both URIs from the first move.
	ts, err = suite.state.DB.GetLatestMoveSuccessInvolvingURIs(
		ctx,
		firstMove.OriginURI,
		firstMove.TargetURI,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Time should be equivalent to secondMove.
	suite.EqualValues(secondMove.SucceededAt.UnixMilli(), ts.UnixMilli())

	// Test getting succeeded using
	// URI from the first Move, and
	// some random account.
	ts, err = suite.state.DB.GetLatestMoveSuccessInvolvingURIs(
		ctx,
		firstMove.OriginURI,
		"https://a.secret.third.place/users/mystery_meat",
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Time should be equivalent to firstMove.
	suite.EqualValues(firstMove.SucceededAt.UnixMilli(), ts.UnixMilli())

	// Delete the first Move.
	if err := suite.state.DB.DeleteMoveByID(ctx, firstMove.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Ensure first Move deleted.
	_, err = suite.state.DB.GetMoveByID(ctx, firstMove.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func TestMoveTestSuite(t *testing.T) {
	suite.Run(t, new(MoveTestSuite))
}
