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
	"encoding/json"
	"testing"
	"time"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	msg, _ := suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ActorPerson, msg.APObjectType)
	suite.Equal(ap.ActivityMove, msg.APActivityType)

	// Stub Move should be on the message.
	move, ok := msg.GTSModel.(*gtsmodel.Move)
	if !ok {
		suite.FailNow("", "could not cast %T to *gtsmodel.Move", msg.GTSModel)
	}
	suite.Equal("http://fossbros-anonymous.io/users/foss_satan", move.OriginURI)
	suite.Equal("https://turnip.farm/users/turniplover6969", move.TargetURI)

	// Trigger the same move again.
	suite.move(receivingAcct, requestingAcct, moveStr1)

	// Should be a message heading to the processor
	// since this is just a straight up retry.
	msg, _ = suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ActorPerson, msg.APObjectType)
	suite.Equal(ap.ActivityMove, msg.APActivityType)

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

	// Trigger the move.
	suite.move(receivingAcct, requestingAcct, moveStr2)

	// Should be a message heading to the processor
	// since this is just a retry with a different ID.
	msg, _ = suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ActorPerson, msg.APObjectType)
	suite.Equal(ap.ActivityMove, msg.APActivityType)
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
