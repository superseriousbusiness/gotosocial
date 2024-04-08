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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountTestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *AccountTestSuite) TestAccountDeleteLocal() {
	ctx := context.Background()
	deletingAccount := suite.testAccounts["local_account_1"]
	followingAccount := suite.testAccounts["remote_account_1"]

	// make the following account follow the deleting account so that a delete message will be sent to it via the federating API
	follow := &gtsmodel.Follow{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", followingAccount.URI),
		AccountID:       followingAccount.ID,
		TargetAccountID: deletingAccount.ID,
	}
	err := suite.db.Put(ctx, follow)
	suite.NoError(err)

	errWithCode := suite.processor.Account().DeleteSelf(ctx, suite.testAccounts["local_account_1"])
	suite.NoError(errWithCode)

	// the delete should be federated outwards to the following account's inbox
	var sent []byte
	delete := new(struct {
		Actor  string `json:"actor"`
		ID     string `json:"id"`
		Object string `json:"object"`
		To     string `json:"to"`
		CC     string `json:"cc"`
		Type   string `json:"type"`
	})

	if !testrig.WaitFor(func() bool {
		delivery, ok := suite.state.Workers.Delivery.Queue.Pop()
		if !ok {
			return false
		}
		if !testrig.EqualRequestURIs(delivery.Request.URL, *followingAccount.SharedInboxURI) {
			panic("differing request uris")
		}
		sent, err = io.ReadAll(delivery.Request.Body)
		if err != nil {
			panic("error reading body: " + err.Error())
		}
		err = json.Unmarshal(sent, delete)
		if err != nil {
			panic("error unmarshaling json: " + err.Error())
		}
		return true
	}) {
		suite.FailNow("timed out waiting for message")
	}

	suite.Equal(deletingAccount.URI, delete.Actor)
	suite.Equal(deletingAccount.URI, delete.Object)
	suite.Equal(deletingAccount.FollowersURI, delete.To)
	suite.Equal(pub.PublicActivityPubIRI, delete.CC)
	suite.Equal("Delete", delete.Type)

	if !testrig.WaitFor(func() bool {
		dbAccount, _ := suite.db.GetAccountByID(ctx, deletingAccount.ID)
		return !dbAccount.SuspendedAt.IsZero()
	}) {
		suite.FailNow("timed out waiting for account to be deleted")
	}
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, &AccountTestSuite{})
}
