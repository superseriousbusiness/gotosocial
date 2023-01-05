/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"github.com/superseriousbusiness/activity/pub"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
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

	errWithCode := suite.processor.AccountDeleteLocal(ctx, suite.testAutheds["local_account_1"], &apimodel.AccountDeleteRequest{
		Password:       "password",
		DeleteOriginID: deletingAccount.ID,
	})
	suite.NoError(errWithCode)

	// the delete should be federated outwards to the following account's inbox
	var sent [][]byte
	delete := new(struct {
		Actor  string `json:"actor"`
		ID     string `json:"id"`
		Object string `json:"object"`
		To     string `json:"to"`
		CC     string `json:"cc"`
		Type   string `json:"type"`
	})

	if !testrig.WaitFor(func() bool {
		sentI, ok := suite.httpClient.SentMessages.Load(*followingAccount.SharedInboxURI)
		if ok {
			sent, ok = sentI.([][]byte)
			if !ok {
				panic("SentMessages entry was not [][]byte")
			}
			err = json.Unmarshal(sent[0], delete)
			return err == nil
		}
		return false
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
		return suite.WithinDuration(dbAccount.SuspendedAt, time.Now(), 30*time.Second)
	}) {
		suite.FailNow("timed out waiting for account to be deleted")
	}
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, &AccountTestSuite{})
}
