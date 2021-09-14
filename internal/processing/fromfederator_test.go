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
	"github.com/superseriousbusiness/gotosocial/internal/id"
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

func (suite *FromFederatorTestSuite) TestProcessReplyMention() {
	repliedAccount := suite.testAccounts["local_account_1"]
	repliedStatus := suite.testStatuses["local_account_1_status_1"]
	replyingAccount := suite.testAccounts["remote_account_1"]
	replyingStatus := &gtsmodel.Status{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		URI:       "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552",
		URL:       "http://fossbros-anonymous.io/@foss_satan/106221634728637552",
		Content:   `<p><span class="h-card"><a href="http://localhost:8080/@the_mighty_zork" class="u-url mention">@<span>the_mighty_zork</span></a></span> nice there it is:</p><p><a href="http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity" rel="nofollow noopener noreferrer" target="_blank"><span class="invisible">https://</span><span class="ellipsis">social.pixie.town/users/f0x/st</span><span class="invisible">atuses/106221628567855262/activity</span></a></p>`,
		Mentions: []*gtsmodel.Mention{
			{
				TargetAccountURI: repliedAccount.URI,
				NameString:       "@the_mighty_zork@localhost:8080",
			},
		},
		AccountID:           replyingAccount.ID,
		AccountURI:          replyingAccount.URI,
		InReplyToID:         repliedStatus.ID,
		InReplyToURI:        repliedStatus.URI,
		InReplyToAccountID:  repliedAccount.ID,
		Visibility:          gtsmodel.VisibilityUnlocked,
		ActivityStreamsType: ap.ObjectNote,
		Federated:           true,
		Boostable:           true,
		Replyable:           true,
		Likeable:            true,
	}

	// id the status based on the time it was created
	statusID, err := id.NewULIDFromTime(replyingStatus.CreatedAt)
	suite.NoError(err)
	replyingStatus.ID = statusID

	err = suite.db.PutStatus(context.Background(), replyingStatus)
	suite.NoError(err)

	err = suite.processor.ProcessFromFederator(context.Background(), messages.FromFederator{
		APObjectType:     ap.ObjectNote,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         replyingStatus,
		ReceivingAccount: suite.testAccounts["local_account_1"],
	})
	suite.NoError(err)

	// side effects should be triggered
	// 1. status should be in the database
	suite.NotEmpty(replyingStatus.ID)
	_, err = suite.db.GetStatusByID(context.Background(), replyingStatus.ID)
	suite.NoError(err)

	// 2. a notification should exist for the mention
	where := []db.Where{
		{
			Key:   "status_id",
			Value: replyingStatus.ID,
		},
	}

	notif := &gtsmodel.Notification{}
	err = suite.db.GetWhere(context.Background(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationMention, notif.NotificationType)
	suite.Equal(replyingStatus.InReplyToAccountID, notif.TargetAccountID)
	suite.Equal(replyingStatus.AccountID, notif.OriginAccountID)
	suite.Equal(replyingStatus.ID, notif.StatusID)
	suite.False(notif.Read)
}

func TestFromFederatorTestSuite(t *testing.T) {
	suite.Run(t, &FromFederatorTestSuite{})
}
