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

package processing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FollowRequestTestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *FollowRequestTestSuite) TestFollowRequestAccept() {
	requestingAccount := suite.testAccounts["remote_account_2"]
	targetAccount := suite.testAccounts["local_account_1"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", requestingAccount.URI),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	relationship, errWithCode := suite.processor.FollowRequestAccept(context.Background(), suite.testAutheds["local_account_1"], requestingAccount.ID)
	suite.NoError(errWithCode)
	suite.EqualValues(&apimodel.Relationship{ID: "01FHMQX3GAABWSM0S2VZEC2SWC", Following: false, ShowingReblogs: false, Notifying: false, FollowedBy: true, Blocking: false, BlockedBy: false, Muting: false, MutingNotifications: false, Requested: false, DomainBlocking: false, Endorsed: false, Note: ""}, relationship)

	// accept should be sent to Some_User
	var sent [][]byte
	if !testrig.WaitFor(func() bool {
		sentI, ok := suite.httpClient.SentMessages.Load(requestingAccount.InboxURI)
		if ok {
			sent, ok = sentI.([][]byte)
			if !ok {
				panic("SentMessages entry was not []byte")
			}
			return true
		}
		return false
	}) {
		suite.FailNow("timed out waiting for message")
	}

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
	err = json.Unmarshal(sent[0], accept)
	suite.NoError(err)

	suite.Equal(targetAccount.URI, accept.Actor)
	suite.Equal(requestingAccount.URI, accept.Object.Actor)
	suite.Equal(fr.URI, accept.Object.ID)
	suite.Equal(targetAccount.URI, accept.Object.Object)
	suite.Equal(targetAccount.URI, accept.Object.To)
	suite.Equal("Follow", accept.Object.Type)
	suite.Equal(requestingAccount.URI, accept.To)
	suite.Equal("Accept", accept.Type)
}

func (suite *FollowRequestTestSuite) TestFollowRequestReject() {
	requestingAccount := suite.testAccounts["remote_account_2"]
	targetAccount := suite.testAccounts["local_account_1"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", requestingAccount.URI),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	relationship, errWithCode := suite.processor.FollowRequestReject(context.Background(), suite.testAutheds["local_account_1"], requestingAccount.ID)
	suite.NoError(errWithCode)
	suite.EqualValues(&apimodel.Relationship{ID: "01FHMQX3GAABWSM0S2VZEC2SWC", Following: false, ShowingReblogs: false, Notifying: false, FollowedBy: false, Blocking: false, BlockedBy: false, Muting: false, MutingNotifications: false, Requested: false, DomainBlocking: false, Endorsed: false, Note: ""}, relationship)

	// reject should be sent to Some_User
	var sent [][]byte
	if !testrig.WaitFor(func() bool {
		sentI, ok := suite.httpClient.SentMessages.Load(requestingAccount.InboxURI)
		if ok {
			sent, ok = sentI.([][]byte)
			if !ok {
				panic("SentMessages entry was not []byte")
			}
			return true
		}
		return false
	}) {
		suite.FailNow("timed out waiting for message")
	}

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
	err = json.Unmarshal(sent[0], reject)
	suite.NoError(err)

	suite.Equal(targetAccount.URI, reject.Actor)
	suite.Equal(requestingAccount.URI, reject.Object.Actor)
	suite.Equal(fr.URI, reject.Object.ID)
	suite.Equal(targetAccount.URI, reject.Object.Object)
	suite.Equal(targetAccount.URI, reject.Object.To)
	suite.Equal("Follow", reject.Object.Type)
	suite.Equal(requestingAccount.URI, reject.To)
	suite.Equal("Reject", reject.Type)
}

func TestFollowRequestTestSuite(t *testing.T) {
	suite.Run(t, &FollowRequestTestSuite{})
}
