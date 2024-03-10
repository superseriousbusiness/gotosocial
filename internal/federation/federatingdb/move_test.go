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

package federatingdb_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

type MoveTestSuite struct {
	FederatingDBTestSuite
}

func (suite *MoveTestSuite) move(
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
	moveStr string,
) error {
	ctx := createTestContext(receivingAcct, requestingAcct)

	rawMove := make(map[string]interface{})
	if err := json.Unmarshal([]byte(moveStr), &rawMove); err != nil {
		suite.FailNow(err.Error())
	}

	t, err := streams.ToType(ctx, rawMove)
	if err != nil {
		suite.FailNow(err.Error())
	}

	move, ok := t.(vocab.ActivityStreamsMove)
	if !ok {
		suite.FailNow("", "couldn't cast %T to Move", t)
	}

	return suite.federatingDB.Move(ctx, move)
}

func (suite *MoveTestSuite) TestMove() {
	var (
		receivingAcct  = suite.testAccounts["local_account_1"]
		requestingAcct = suite.testAccounts["remote_account_1"]
		moveStr1       = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://fossbros-anonymous.io/users/foss_satan/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
  "actor": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Move",
  "object": "http://fossbros-anonymous.io/users/foss_satan",
  "target": "https://turnip.farm/users/turniplover6969",
  "to": "http://fossbros-anonymous.io/users/foss_satan/followers"
}`
	)

	// Trigger the move.
	suite.move(receivingAcct, requestingAcct, moveStr1)

	// Should be a message heading to the processor.
	var msg1 messages.FromFediAPI
	select {
	case msg1 = <-suite.fromFederator:
		// Fine.
	case <-time.After(5 * time.Second):
		suite.FailNow("", "timeout waiting for suite.fromFederator")
	}
	suite.Equal(ap.ObjectProfile, msg1.APObjectType)
	suite.Equal(ap.ActivityMove, msg1.APActivityType)

	// A Move should now be in the database.
	move, err := suite.state.DB.GetMoveByURI(
		context.Background(),
		"http://fossbros-anonymous.io/users/foss_satan/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("http://fossbros-anonymous.io/users/foss_satan", move.OriginURI)
	suite.Equal("https://turnip.farm/users/turniplover6969", move.TargetURI)

	// Update the Move to set attempted_at > 5 minutes
	// ago, to avoid retry rate limiting.
	move.AttemptedAt = move.AttemptedAt.Add(-10 * time.Minute)
	if err := suite.state.DB.UpdateMove(context.Background(), move, "attempted_at"); err != nil {
		suite.FailNow(err.Error())
	}

	// Trigger the same move again.
	suite.move(receivingAcct, requestingAcct, moveStr1)

	// Should be a message heading to the processor
	// since this is just a straight up retry.
	var msg2 messages.FromFediAPI
	select {
	case msg2 = <-suite.fromFederator:
		// Fine.
	case <-time.After(5 * time.Second):
		suite.FailNow("", "timeout waiting for suite.fromFederator")
	}
	suite.Equal(ap.ObjectProfile, msg2.APObjectType)
	suite.Equal(ap.ActivityMove, msg2.APActivityType)

	// Same as the first Move, but with a different ID.
	moveStr2 := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://fossbros-anonymous.io/users/foss_satan/moves/01HR9XWDD25CKXHW82MYD1GDAR",
  "actor": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Move",
  "object": "http://fossbros-anonymous.io/users/foss_satan",
  "target": "https://turnip.farm/users/turniplover6969",
  "to": "http://fossbros-anonymous.io/users/foss_satan/followers"
}`

	// Update the Move to set attempted_at > 5 minutes
	// ago, to avoid retry rate limiting.
	move.AttemptedAt = move.AttemptedAt.Add(-10 * time.Minute)
	if err := suite.state.DB.UpdateMove(context.Background(), move, "attempted_at"); err != nil {
		suite.FailNow(err.Error())
	}

	// Trigger the move.
	suite.move(receivingAcct, requestingAcct, moveStr2)

	// Should be a message heading to the processor
	// since this is just a retry with a different ID.
	var msg3 messages.FromFediAPI
	select {
	case msg3 = <-suite.fromFederator:
		// Fine.
	case <-time.After(5 * time.Second):
		suite.FailNow("", "timeout waiting for suite.fromFederator")
	}
	suite.Equal(ap.ObjectProfile, msg3.APObjectType)
	suite.Equal(ap.ActivityMove, msg3.APActivityType)

	// The Move in the database should
	// have the new ID/URI set on it.
	move, err = suite.state.DB.GetMoveByID(
		context.Background(),
		move.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("http://fossbros-anonymous.io/users/foss_satan/moves/01HR9XWDD25CKXHW82MYD1GDAR", move.URI)
	suite.Equal("http://fossbros-anonymous.io/users/foss_satan", move.OriginURI)
	suite.Equal("https://turnip.farm/users/turniplover6969", move.TargetURI)
}

func (suite *MoveTestSuite) TestBadMoves() {
	var (
		receivingAcct  = suite.testAccounts["local_account_1"]
		requestingAcct = suite.testAccounts["remote_account_1"]
	)

	type testStruct struct {
		moveStr string
		err     string
	}

	for _, t := range []testStruct{
		{
			// Move signed by someone else.
			moveStr: `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://fossbros-anonymous.io/users/foss_satan/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
  "actor": "http://fossbros-anonymous.io/users/someone_else",
  "type": "Move",
  "object": "http://fossbros-anonymous.io/users/foss_satan",
  "target": "https://turnip.farm/users/turniplover6969",
  "to": "http://fossbros-anonymous.io/users/foss_satan/followers"
}`,
			err: "Move was signed by http://fossbros-anonymous.io/users/foss_satan but actor was http://fossbros-anonymous.io/users/someone_else",
		},
		{
			// Actor and object not the same.
			moveStr: `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://fossbros-anonymous.io/users/foss_satan/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
  "actor": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Move",
  "object": "http://fossbros-anonymous.io/users/someone_else",
  "target": "https://turnip.farm/users/turniplover6969",
  "to": "http://fossbros-anonymous.io/users/foss_satan/followers"
}`,
			err: "Move was signed by http://fossbros-anonymous.io/users/foss_satan but object was http://fossbros-anonymous.io/users/someone_else",
		},
		{
			// Object and target the same.
			moveStr: `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://fossbros-anonymous.io/users/foss_satan/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
  "actor": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Move",
  "object": "http://fossbros-anonymous.io/users/foss_satan",
  "target": "http://fossbros-anonymous.io/users/foss_satan",
  "to": "http://fossbros-anonymous.io/users/foss_satan/followers"
}`,
			err: "Move target and origin were the same (http://fossbros-anonymous.io/users/foss_satan)",
		},
	} {
		// Trigger the move.
		err := suite.move(receivingAcct, requestingAcct, t.moveStr)
		if t.err != "" {
			suite.EqualError(err, t.err)
		}
	}
}

func TestMoveTestSuite(t *testing.T) {
	suite.Run(t, &MoveTestSuite{})
}
