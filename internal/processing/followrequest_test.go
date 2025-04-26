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

package processing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

// TODO: move this to the "internal/processing/account" pkg
type FollowRequestTestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *FollowRequestTestSuite) TestFollowRequestAccept() {
	// The authed local account we are going to use for HTTP requests
	requestingAccount := suite.testAccounts["local_account_1"]

	// The remote account whose follow request we are accepting
	targetAccount := suite.testAccounts["remote_account_2"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", targetAccount.URI),
		AccountID:       targetAccount.ID,
		TargetAccountID: requestingAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	relationship, errWithCode := suite.processor.Account().FollowRequestAccept(
		context.Background(),
		requestingAccount,
		targetAccount.ID,
	)
	suite.NoError(errWithCode)
	suite.EqualValues(&apimodel.Relationship{
		ID:                  "01FHMQX3GAABWSM0S2VZEC2SWC",
		Following:           false,
		ShowingReblogs:      false,
		Notifying:           false,
		FollowedBy:          true,
		Blocking:            false,
		BlockedBy:           false,
		Muting:              false,
		MutingNotifications: false,
		Requested:           false,
		DomainBlocking:      false,
		Endorsed:            false,
		Note:                "",
	}, relationship)

	accept := &struct {
		Actor  string `json:"actor"`
		ID     string `json:"id"`
		Object struct {
			Actor  string `json:"actor"`
			ID     string `json:"id"`
			Object string `json:"object"`
			To     string `json:"to"`
			Type   string `json:"type"`
		}
		To   string `json:"to"`
		Type string `json:"type"`
	}{}

	// accept should be sent to Some_User
	var sent []byte
	if !testrig.WaitFor(func() bool {
		delivery, ok := suite.state.Workers.Delivery.Queue.Pop()
		if !ok {
			return false
		}
		if !testrig.EqualRequestURIs(delivery.Request.URL, targetAccount.InboxURI) {
			panic("differing request uris")
		}
		sent, err = io.ReadAll(delivery.Request.Body)
		if err != nil {
			panic("error reading body: " + err.Error())
		}
		err = json.Unmarshal(sent, accept)
		if err != nil {
			panic("error unmarshaling json: " + err.Error())
		}
		return true
	}) {
		suite.FailNow("timed out waiting for message")
	}

	suite.Equal(requestingAccount.URI, accept.Actor)
	suite.Equal(targetAccount.URI, accept.Object.Actor)
	suite.Equal(fr.URI, accept.Object.ID)
	suite.Equal(requestingAccount.URI, accept.Object.Object)
	suite.Equal(requestingAccount.URI, accept.Object.To)
	suite.Equal("Follow", accept.Object.Type)
	suite.Equal(targetAccount.URI, accept.To)
	suite.Equal("Accept", accept.Type)
}

func (suite *FollowRequestTestSuite) TestFollowRequestReject() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["remote_account_2"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", targetAccount.URI),
		AccountID:       targetAccount.ID,
		TargetAccountID: requestingAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	relationship, errWithCode := suite.processor.Account().FollowRequestReject(
		context.Background(),
		requestingAccount,
		targetAccount.ID,
	)
	suite.NoError(errWithCode)
	suite.EqualValues(&apimodel.Relationship{ID: "01FHMQX3GAABWSM0S2VZEC2SWC", Following: false, ShowingReblogs: false, Notifying: false, FollowedBy: false, Blocking: false, BlockedBy: false, Muting: false, MutingNotifications: false, Requested: false, DomainBlocking: false, Endorsed: false, Note: ""}, relationship)

	reject := &struct {
		Actor  string `json:"actor"`
		ID     string `json:"id"`
		Object struct {
			Actor  string `json:"actor"`
			ID     string `json:"id"`
			Object string `json:"object"`
			To     string `json:"to"`
			Type   string `json:"type"`
		}
		To   string `json:"to"`
		Type string `json:"type"`
	}{}

	// reject should be sent to Some_User
	var sent []byte
	if !testrig.WaitFor(func() bool {
		delivery, ok := suite.state.Workers.Delivery.Queue.Pop()
		if !ok {
			return false
		}
		if !testrig.EqualRequestURIs(delivery.Request.URL, targetAccount.InboxURI) {
			panic("differing request uris")
		}
		sent, err = io.ReadAll(delivery.Request.Body)
		if err != nil {
			panic("error reading body: " + err.Error())
		}
		err = json.Unmarshal(sent, reject)
		if err != nil {
			panic("error unmarshaling json: " + err.Error())
		}
		return true
	}) {
		suite.FailNow("timed out waiting for message")
	}

	suite.Equal(requestingAccount.URI, reject.Actor)
	suite.Equal(targetAccount.URI, reject.Object.Actor)
	suite.Equal(fr.URI, reject.Object.ID)
	suite.Equal(requestingAccount.URI, reject.Object.Object)
	suite.Equal(requestingAccount.URI, reject.Object.To)
	suite.Equal("Follow", reject.Object.Type)
	suite.Equal(targetAccount.URI, reject.To)
	suite.Equal("Reject", reject.Type)
}

func TestFollowRequestTestSuite(t *testing.T) {
	suite.Run(t, &FollowRequestTestSuite{})
}
