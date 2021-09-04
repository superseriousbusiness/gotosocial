/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

type FromFederatorTestSuite struct {
	ProcessingStandardTestSuite
}

// remote_account_1 boosts the first status of local_account_1
func (suite *FromFederatorTestSuite) TestProcessFederationAnnounce() {
	boostedStatus := suite.testStatuses["local_account_1_status_1"]
	boostingAccount := suite.testAccounts["remote_account_1"]
	announceStatus := &gtsmodel.Status{}
	announceStatus.URI = "https://example.org/some-announce-uri"
	announceStatus.BoostOf = &gtsmodel.Status{
		URI: boostedStatus.URI,
	}
	announceStatus.CreatedAt = time.Now()
	announceStatus.UpdatedAt = time.Now()
	announceStatus.AccountID = boostingAccount.ID
	announceStatus.AccountURI = boostingAccount.URI
	announceStatus.Account = boostingAccount
	announceStatus.Visibility = boostedStatus.Visibility

	err := suite.processor.ProcessFromFederator(context.Background(), messages.FromFederator{
		APObjectType:     ap.ActivityAnnounce,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         announceStatus,
		ReceivingAccount: suite.testAccounts["local_account_1"],
	})
	suite.NoError(err)

	// side effects should be triggered
	// 1. status should have an ID, and be in the database
	suite.NotEmpty(announceStatus.ID)
	_, err = suite.db.GetStatusByID(context.Background(), announceStatus.ID)
	suite.NoError(err)

	// 2. a notification should exist for the announce
	where := []db.Where{
		{
			Key:   "status_id",
			Value: announceStatus.ID,
		},
	}
	notif := &gtsmodel.Notification{}
	err = suite.db.GetWhere(context.Background(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationReblog, notif.NotificationType)
	suite.Equal(boostedStatus.AccountID, notif.TargetAccountID)
	suite.Equal(announceStatus.AccountID, notif.OriginAccountID)
	suite.Equal(announceStatus.ID, notif.StatusID)
	suite.False(notif.Read)
}

func TestFromFederatorTestSuite(t *testing.T) {
	suite.Run(t, &FromFederatorTestSuite{})
}
