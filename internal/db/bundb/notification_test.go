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
	"errors"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

func (suite *NotificationTestSuite) spamNotifs() {
	// spam a shit ton of notifs into the database
	// half of them will be for zork, the other half
	// will be for random accounts
	notifCount := 10000

	zork := suite.testAccounts["local_account_1"]

	for i := 0; i < notifCount; i++ {
		notifID := id.NewULID()

		var targetAccountID string
		if i%2 == 0 {
			targetAccountID = zork.ID
		} else {
			randomAssID, err := id.NewRandomULID()
			if err != nil {
				panic(err)
			}
			targetAccountID = randomAssID
		}

		statusID, err := id.NewRandomULID()
		if err != nil {
			panic(err)
		}

		originAccountID, err := id.NewRandomULID()
		if err != nil {
			panic(err)
		}

		notif := &gtsmodel.Notification{
			ID:               notifID,
			NotificationType: gtsmodel.NotificationFavourite,
			CreatedAt:        time.Now(),
			TargetAccountID:  targetAccountID,
			OriginAccountID:  originAccountID,
			StatusOrEditID:   statusID,
			Read:             util.Ptr(false),
		}

		if err := suite.db.PutNotification(suite.T().Context(), notif); err != nil {
			panic(err)
		}
	}

	suite.T().Logf("put %d notifs in the db\n", notifCount)
}

type NotificationTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *NotificationTestSuite) TestGetAccountNotificationsWithSpam() {
	suite.spamNotifs()
	testAccount := suite.testAccounts["local_account_1"]
	before := time.Now()
	notifications, err := suite.db.GetAccountNotifications(
		gtscontext.SetBarebones(suite.T().Context()),
		testAccount.ID,
		&paging.Page{
			Min:   paging.EitherMinID("", id.Lowest),
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		},
		nil,
		nil,
	)
	suite.NoError(err)
	timeTaken := time.Since(before)
	suite.T().Logf("withSpam: got %d notifications in %s\n", len(notifications), timeTaken)

	suite.NotNil(notifications)
	for _, n := range notifications {
		suite.Equal(testAccount.ID, n.TargetAccountID)
	}
}

func (suite *NotificationTestSuite) TestGetAccountNotificationsWithoutSpam() {
	testAccount := suite.testAccounts["local_account_1"]
	before := time.Now()
	notifications, err := suite.db.GetAccountNotifications(
		gtscontext.SetBarebones(suite.T().Context()),
		testAccount.ID,
		&paging.Page{
			Min:   paging.EitherMinID("", id.Lowest),
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		},
		nil,
		nil,
	)
	suite.NoError(err)
	timeTaken := time.Since(before)
	suite.T().Logf("withoutSpam: got %d notifications in %s\n", len(notifications), timeTaken)

	suite.NotNil(notifications)
	for _, n := range notifications {
		suite.Equal(testAccount.ID, n.TargetAccountID)
	}
}

func (suite *NotificationTestSuite) TestDeleteNotificationsWithSpam() {
	suite.spamNotifs()
	testAccount := suite.testAccounts["local_account_1"]

	// Test getting notifs first.
	notifications, err := suite.db.GetAccountNotifications(
		gtscontext.SetBarebones(suite.T().Context()),
		testAccount.ID,
		&paging.Page{
			Min:   paging.EitherMinID("", id.Lowest),
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		},
		nil,
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(notifications, 20)

	// Now delete.
	if err := suite.db.DeleteNotifications(suite.T().Context(), nil, testAccount.ID, ""); err != nil {
		suite.FailNow(err.Error())
	}

	// Now try getting again.
	notifications, err = suite.db.GetAccountNotifications(
		gtscontext.SetBarebones(suite.T().Context()),
		testAccount.ID,
		&paging.Page{
			Min:   paging.EitherMinID("", id.Lowest),
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		},
		nil,
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Empty(notifications)
}

func (suite *NotificationTestSuite) TestDeleteNotificationsWithTwoAccounts() {
	suite.spamNotifs()
	testAccount := suite.testAccounts["local_account_1"]
	err := suite.db.DeleteNotifications(suite.T().Context(), nil, testAccount.ID, "")
	suite.NoError(err)

	notifications, err := suite.db.GetAccountNotifications(
		gtscontext.SetBarebones(suite.T().Context()),
		testAccount.ID,
		&paging.Page{
			Min:   paging.EitherMinID("", id.Lowest),
			Max:   paging.MaxID(id.Highest),
			Limit: 20,
		},
		nil,
		nil,
	)
	suite.NoError(err)
	suite.Nil(notifications)
	suite.Empty(notifications)

	notif := []*gtsmodel.Notification{}
	err = suite.db.GetAll(suite.T().Context(), &notif)
	suite.NoError(err)
	suite.NotEmpty(notif)
}

func (suite *NotificationTestSuite) TestDeleteNotificationsOriginatingFromAccount() {
	testAccount := suite.testAccounts["local_account_2"]

	if err := suite.db.DeleteNotifications(suite.T().Context(), nil, "", testAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	notif := []*gtsmodel.Notification{}
	if err := suite.db.GetAll(suite.T().Context(), &notif); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, n := range notif {
		if n.OriginAccountID == testAccount.ID {
			suite.FailNowf("", "no notifications with origin account id %s should remain", testAccount.ID)
		}
	}
}

func (suite *NotificationTestSuite) TestDeleteNotificationsOriginatingFromAndTargetingAccount() {
	originAccount := suite.testAccounts["local_account_2"]
	targetAccount := suite.testAccounts["admin_account"]

	if err := suite.db.DeleteNotifications(suite.T().Context(), nil, targetAccount.ID, originAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	notif := []*gtsmodel.Notification{}
	if err := suite.db.GetAll(suite.T().Context(), &notif); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, n := range notif {
		if n.OriginAccountID == originAccount.ID && n.TargetAccountID == targetAccount.ID {
			suite.FailNowf(
				"",
				"no notifications with origin account id %s and target account %s should remain",
				originAccount.ID,
				targetAccount.ID,
			)
		}
	}
}

func (suite *NotificationTestSuite) TestDeleteNotificationsPertainingToStatusID() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	if err := suite.db.DeleteNotificationsForStatus(suite.T().Context(), testStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	notif := []*gtsmodel.Notification{}
	if err := suite.db.GetAll(suite.T().Context(), &notif); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, n := range notif {
		if n.StatusOrEditID == testStatus.ID {
			suite.FailNowf("", "no notifications with status id %s should remain", testStatus.ID)
		}
	}
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}
