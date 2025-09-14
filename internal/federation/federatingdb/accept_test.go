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
	"bytes"
	"io"
	"testing"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

const (
	rMediaPath    = "../../../testrig/media"
	rTemplatePath = "../../../web/template"
)

type AcceptTestSuite struct {
	FederatingDBTestSuite
}

func (suite *AcceptTestSuite) TestAcceptRemoteReplyRequest() {
	// Accept of a reply by
	// brand_new_person to foss_satan.
	const acceptJSON = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "type": "Accept",
  "to": "https://unknown-instance.com/users/brand_new_person",
  "cc": "https://www.w3.org/ns/activitystreams#Public",
  "id": "http://fossbros-anonymous.io/users/foss_satan/accepts/1234",
  "actor": "http://fossbros-anonymous.io/users/foss_satan",
  "object": {
    "type": "ReplyRequest",
    "id": "https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7/replyRequest",
    "actor": "https://unknown-instance.com/users/brand_new_person",
    "object": "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
    "instrument": "https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7"
  },
  "result": "http://fossbros-anonymous.io/users/foss_satan/authorizations/1234"
}`

	// The accept will be delivered by foss_satan to zork.
	ctx := createTestContext(
		suite.T(),
		suite.testAccounts["local_account_1"],
		suite.testAccounts["remote_account_1"],
	)

	// Have zork follow foss_satan for this test,
	// else the message will be scattered unto the four winds.
	follow := &gtsmodel.Follow{
		ID:              "01K4STEH5NWAXBZ4TFNGQQQ984",
		CreatedAt:       testrig.TimeMustParse("2022-05-14T13:21:09+02:00"),
		UpdatedAt:       testrig.TimeMustParse("2022-05-14T13:21:09+02:00"),
		AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
		TargetAccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
		URI:             "http://localhost:8080/users/the_mighty_zork/follow/01G1TK3PQKFW1BQZ9WVYRTFECK",
	}
	if err := suite.state.DB.PutFollow(ctx, follow); err != nil {
		suite.FailNow(err.Error())
	}

	// Parse accept into vocab.Type.
	t, err := ap.DecodeType(ctx, io.NopCloser(bytes.NewBufferString(acceptJSON)))
	if err != nil {
		suite.FailNow(err.Error())
	}
	accept := t.(vocab.ActivityStreamsAccept)

	// Process the accept.
	if err := suite.federatingDB.Accept(ctx, accept); err != nil {
		suite.FailNow(err.Error())
	}

	// There should be an accept msg
	// heading to the processor now.
	msg, ok := suite.state.Workers.Federator.Queue.PopCtx(ctx)
	if !ok {
		suite.FailNow("no message in queue")
	}

	suite.EqualValues(
		&messages.FromFediAPI{
			APObjectType:   "ReplyRequest",
			APActivityType: "Accept",
			APIRI:          testrig.URLMustParse("http://fossbros-anonymous.io/users/foss_satan/authorizations/1234"),
			APObject:       testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7"),
			Requesting:     suite.testAccounts["remote_account_1"],
			Receiving:      suite.testAccounts["local_account_1"],
		},
		msg,
	)
}

func TestAcceptTestSuite(t *testing.T) {
	suite.Run(t, new(AcceptTestSuite))
}
