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
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/testrig"
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
	suite.False(*notif.Read)
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
		Federated:           testrig.TrueBool(),
		Boostable:           testrig.TrueBool(),
		Replyable:           testrig.TrueBool(),
		Likeable:            testrig.FalseBool(),
	}

	wssStream, errWithCode := suite.processor.OpenStreamForAccount(context.Background(), repliedAccount, stream.TimelineHome)
	suite.NoError(errWithCode)

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
	suite.False(*notif.Read)

	// the notification should be streamed
	var msg *stream.Message
	select {
	case msg = <-wssStream.Messages:
		// fine
	case <-time.After(5 * time.Second):
		suite.FailNow("no message from wssStream")
	}

	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)
	notifStreamed := &apimodel.Notification{}
	err = json.Unmarshal([]byte(msg.Payload), notifStreamed)
	suite.NoError(err)
	suite.Equal("mention", notifStreamed.Type)
	suite.Equal(replyingAccount.ID, notifStreamed.Account.ID)
}

func (suite *FromFederatorTestSuite) TestProcessFave() {
	favedAccount := suite.testAccounts["local_account_1"]
	favedStatus := suite.testStatuses["local_account_1_status_1"]
	favingAccount := suite.testAccounts["remote_account_1"]

	wssStream, errWithCode := suite.processor.OpenStreamForAccount(context.Background(), favedAccount, stream.TimelineNotifications)
	suite.NoError(errWithCode)

	fave := &gtsmodel.StatusFave{
		ID:              "01FGKJPXFTVQPG9YSSZ95ADS7Q",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AccountID:       favingAccount.ID,
		Account:         favingAccount,
		TargetAccountID: favedAccount.ID,
		TargetAccount:   favedAccount,
		StatusID:        favedStatus.ID,
		Status:          favedStatus,
		URI:             favingAccount.URI + "/faves/aaaaaaaaaaaa",
	}

	err := suite.db.Put(context.Background(), fave)
	suite.NoError(err)

	err = suite.processor.ProcessFromFederator(context.Background(), messages.FromFederator{
		APObjectType:     ap.ActivityLike,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         fave,
		ReceivingAccount: favedAccount,
	})
	suite.NoError(err)

	// side effects should be triggered
	// 1. a notification should exist for the fave
	where := []db.Where{
		{
			Key:   "status_id",
			Value: favedStatus.ID,
		},
		{
			Key:   "origin_account_id",
			Value: favingAccount.ID,
		},
	}

	notif := &gtsmodel.Notification{}
	err = suite.db.GetWhere(context.Background(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationFave, notif.NotificationType)
	suite.Equal(fave.TargetAccountID, notif.TargetAccountID)
	suite.Equal(fave.AccountID, notif.OriginAccountID)
	suite.Equal(fave.StatusID, notif.StatusID)
	suite.False(*notif.Read)

	// 2. a notification should be streamed
	var msg *stream.Message
	select {
	case msg = <-wssStream.Messages:
		// fine
	case <-time.After(5 * time.Second):
		suite.FailNow("no message from wssStream")
	}
	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineNotifications}, msg.Stream)
}

// TestProcessFaveWithDifferentReceivingAccount ensures that when an account receives a fave that's for
// another account in their AP inbox, a notification isn't streamed to the receiving account.
//
// This tests for an issue we were seeing where Misskey sends out faves to inboxes of people that don't own
// the fave, but just follow the actor who received the fave.
func (suite *FromFederatorTestSuite) TestProcessFaveWithDifferentReceivingAccount() {
	receivingAccount := suite.testAccounts["local_account_2"]
	favedAccount := suite.testAccounts["local_account_1"]
	favedStatus := suite.testStatuses["local_account_1_status_1"]
	favingAccount := suite.testAccounts["remote_account_1"]

	wssStream, errWithCode := suite.processor.OpenStreamForAccount(context.Background(), receivingAccount, stream.TimelineHome)
	suite.NoError(errWithCode)

	fave := &gtsmodel.StatusFave{
		ID:              "01FGKJPXFTVQPG9YSSZ95ADS7Q",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AccountID:       favingAccount.ID,
		Account:         favingAccount,
		TargetAccountID: favedAccount.ID,
		TargetAccount:   favedAccount,
		StatusID:        favedStatus.ID,
		Status:          favedStatus,
		URI:             favingAccount.URI + "/faves/aaaaaaaaaaaa",
	}

	err := suite.db.Put(context.Background(), fave)
	suite.NoError(err)

	err = suite.processor.ProcessFromFederator(context.Background(), messages.FromFederator{
		APObjectType:     ap.ActivityLike,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         fave,
		ReceivingAccount: receivingAccount,
	})
	suite.NoError(err)

	// side effects should be triggered
	// 1. a notification should exist for the fave
	where := []db.Where{
		{
			Key:   "status_id",
			Value: favedStatus.ID,
		},
		{
			Key:   "origin_account_id",
			Value: favingAccount.ID,
		},
	}

	notif := &gtsmodel.Notification{}
	err = suite.db.GetWhere(context.Background(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationFave, notif.NotificationType)
	suite.Equal(fave.TargetAccountID, notif.TargetAccountID)
	suite.Equal(fave.AccountID, notif.OriginAccountID)
	suite.Equal(fave.StatusID, notif.StatusID)
	suite.False(*notif.Read)

	// 2. no notification should be streamed to the account that received the fave message, because they weren't the target
	suite.Empty(wssStream.Messages)
}

func (suite *FromFederatorTestSuite) TestProcessAccountDelete() {
	ctx := context.Background()

	deletedAccount := suite.testAccounts["remote_account_1"]
	receivingAccount := suite.testAccounts["local_account_1"]

	// before doing the delete....
	// make local_account_1 and remote_account_1 into mufos
	zorkFollowSatan := &gtsmodel.Follow{
		ID:              "01FGRY72ASHBSET64353DPHK9T",
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		UpdatedAt:       time.Now().Add(-1 * time.Hour),
		AccountID:       deletedAccount.ID,
		TargetAccountID: receivingAccount.ID,
		ShowReblogs:     testrig.TrueBool(),
		URI:             fmt.Sprintf("%s/follows/01FGRY72ASHBSET64353DPHK9T", deletedAccount.URI),
		Notify:          testrig.FalseBool(),
	}
	err := suite.db.Put(ctx, zorkFollowSatan)
	suite.NoError(err)

	satanFollowZork := &gtsmodel.Follow{
		ID:              "01FGRYAVAWWPP926J175QGM0WV",
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		UpdatedAt:       time.Now().Add(-1 * time.Hour),
		AccountID:       receivingAccount.ID,
		TargetAccountID: deletedAccount.ID,
		ShowReblogs:     testrig.TrueBool(),
		URI:             fmt.Sprintf("%s/follows/01FGRYAVAWWPP926J175QGM0WV", receivingAccount.URI),
		Notify:          testrig.FalseBool(),
	}
	err = suite.db.Put(ctx, satanFollowZork)
	suite.NoError(err)

	// now they are mufos!

	err = suite.processor.ProcessFromFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ObjectProfile,
		APActivityType:   ap.ActivityDelete,
		GTSModel:         deletedAccount,
		ReceivingAccount: receivingAccount,
	})
	suite.NoError(err)

	// local account 2 blocked foss_satan, that block should be gone now
	testBlock := suite.testBlocks["local_account_2_block_remote_account_1"]
	dbBlock := &gtsmodel.Block{}
	err = suite.db.GetByID(ctx, testBlock.ID, dbBlock)
	suite.ErrorIs(err, db.ErrNoEntries)

	// the mufos should be gone now too
	satanFollowsZork, err := suite.db.IsFollowing(ctx, deletedAccount, receivingAccount)
	suite.NoError(err)
	suite.False(satanFollowsZork)
	zorkFollowsSatan, err := suite.db.IsFollowing(ctx, receivingAccount, deletedAccount)
	suite.NoError(err)
	suite.False(zorkFollowsSatan)

	// no statuses from foss satan should be left in the database
	if !testrig.WaitFor(func() bool {
		s, err := suite.db.GetAccountStatuses(ctx, deletedAccount.ID, 0, false, false, "", "", false, false, false)
		return s == nil && err == db.ErrNoEntries
	}) {
		suite.FailNow("timeout waiting for statuses to be deleted")
	}

	dbAccount, err := suite.db.GetAccountByID(ctx, deletedAccount.ID)
	suite.NoError(err)

	suite.Empty(dbAccount.Note)
	suite.Empty(dbAccount.DisplayName)
	suite.Empty(dbAccount.AvatarMediaAttachmentID)
	suite.Empty(dbAccount.AvatarRemoteURL)
	suite.Empty(dbAccount.HeaderMediaAttachmentID)
	suite.Empty(dbAccount.HeaderRemoteURL)
	suite.Empty(dbAccount.Reason)
	suite.Empty(dbAccount.Fields)
	suite.True(*dbAccount.HideCollections)
	suite.False(*dbAccount.Discoverable)
	suite.WithinDuration(time.Now(), dbAccount.SuspendedAt, 30*time.Second)
	suite.Equal(dbAccount.ID, dbAccount.SuspensionOrigin)
}

func (suite *FromFederatorTestSuite) TestProcessFollowRequestLocked() {
	ctx := context.Background()

	originAccount := suite.testAccounts["remote_account_1"]

	// target is a locked account
	targetAccount := suite.testAccounts["local_account_2"]

	wssStream, errWithCode := suite.processor.OpenStreamForAccount(context.Background(), targetAccount, stream.TimelineHome)
	suite.NoError(errWithCode)

	// put the follow request in the database as though it had passed through the federating db already
	satanFollowRequestTurtle := &gtsmodel.FollowRequest{
		ID:              "01FGRYAVAWWPP926J175QGM0WV",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AccountID:       originAccount.ID,
		Account:         originAccount,
		TargetAccountID: targetAccount.ID,
		TargetAccount:   targetAccount,
		ShowReblogs:     testrig.TrueBool(),
		URI:             fmt.Sprintf("%s/follows/01FGRYAVAWWPP926J175QGM0WV", originAccount.URI),
		Notify:          testrig.FalseBool(),
	}

	err := suite.db.Put(ctx, satanFollowRequestTurtle)
	suite.NoError(err)

	err = suite.processor.ProcessFromFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityFollow,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         satanFollowRequestTurtle,
		ReceivingAccount: targetAccount,
	})
	suite.NoError(err)

	// a notification should be streamed
	var msg *stream.Message
	select {
	case msg = <-wssStream.Messages:
		// fine
	case <-time.After(5 * time.Second):
		suite.FailNow("no message from wssStream")
	}
	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)
	notif := &apimodel.Notification{}
	err = json.Unmarshal([]byte(msg.Payload), notif)
	suite.NoError(err)
	suite.Equal("follow_request", notif.Type)
	suite.Equal(originAccount.ID, notif.Account.ID)

	// no messages should have been sent out, since we didn't need to federate an accept
	suite.Empty(suite.httpClient.SentMessages)
}

func (suite *FromFederatorTestSuite) TestProcessFollowRequestUnlocked() {
	ctx := context.Background()

	originAccount := suite.testAccounts["remote_account_1"]

	// target is an unlocked account
	targetAccount := suite.testAccounts["local_account_1"]

	wssStream, errWithCode := suite.processor.OpenStreamForAccount(context.Background(), targetAccount, stream.TimelineHome)
	suite.NoError(errWithCode)

	// put the follow request in the database as though it had passed through the federating db already
	satanFollowRequestTurtle := &gtsmodel.FollowRequest{
		ID:              "01FGRYAVAWWPP926J175QGM0WV",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AccountID:       originAccount.ID,
		Account:         originAccount,
		TargetAccountID: targetAccount.ID,
		TargetAccount:   targetAccount,
		ShowReblogs:     testrig.TrueBool(),
		URI:             fmt.Sprintf("%s/follows/01FGRYAVAWWPP926J175QGM0WV", originAccount.URI),
		Notify:          testrig.FalseBool(),
	}

	err := suite.db.Put(ctx, satanFollowRequestTurtle)
	suite.NoError(err)

	err = suite.processor.ProcessFromFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityFollow,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         satanFollowRequestTurtle,
		ReceivingAccount: targetAccount,
	})
	suite.NoError(err)

	// an accept message should be sent to satan's inbox
	var sent [][]byte
	if !testrig.WaitFor(func() bool {
		sentI, ok := suite.httpClient.SentMessages.Load(*originAccount.SharedInboxURI)
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
	suite.Equal(originAccount.URI, accept.Object.Actor)
	suite.Equal(satanFollowRequestTurtle.URI, accept.Object.ID)
	suite.Equal(targetAccount.URI, accept.Object.Object)
	suite.Equal(targetAccount.URI, accept.Object.To)
	suite.Equal("Follow", accept.Object.Type)
	suite.Equal(originAccount.URI, accept.To)
	suite.Equal("Accept", accept.Type)

	// a notification should be streamed
	var msg *stream.Message
	select {
	case msg = <-wssStream.Messages:
		// fine
	case <-time.After(5 * time.Second):
		suite.FailNow("no message from wssStream")
	}
	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)
	notif := &apimodel.Notification{}
	err = json.Unmarshal([]byte(msg.Payload), notif)
	suite.NoError(err)
	suite.Equal("follow", notif.Type)
	suite.Equal(originAccount.ID, notif.Account.ID)
}

// TestCreateStatusFromIRI checks if a forwarded status can be dereferenced by the processor.
func (suite *FromFederatorTestSuite) TestCreateStatusFromIRI() {
	ctx := context.Background()

	receivingAccount := suite.testAccounts["local_account_1"]
	statusCreator := suite.testAccounts["remote_account_2"]

	err := suite.processor.ProcessFromFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ObjectNote,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         nil, // gtsmodel is nil because this is a forwarded status -- we want to dereference it using the iri
		ReceivingAccount: receivingAccount,
		APIri:            testrig.URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1"),
	})
	suite.NoError(err)

	// status should now be in the database, attributed to remote_account_2
	s, err := suite.db.GetStatusByURI(context.Background(), "http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1")
	suite.NoError(err)
	suite.Equal(statusCreator.URI, s.AccountURI)
}

func TestFromFederatorTestSuite(t *testing.T) {
	suite.Run(t, &FromFederatorTestSuite{})
}
