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

package workers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/stream"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type FromFediAPITestSuite struct {
	WorkersTestSuite
}

// remote_account_1 boosts the first status of local_account_1
func (suite *FromFediAPITestSuite) TestProcessFederationAnnounce() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	boostedStatus := &gtsmodel.Status{}
	*boostedStatus = *suite.testStatuses["local_account_1_status_1"]

	boostingAccount := &gtsmodel.Account{}
	*boostingAccount = *suite.testAccounts["remote_account_1"]

	announceStatus := &gtsmodel.Status{}
	announceStatus.URI = "https://example.org/some-announce-uri"
	announceStatus.BoostOfURI = boostedStatus.URI
	announceStatus.CreatedAt = time.Now()
	announceStatus.AccountID = boostingAccount.ID
	announceStatus.AccountURI = boostingAccount.URI
	announceStatus.Account = boostingAccount
	announceStatus.Visibility = boostedStatus.Visibility

	err := testStructs.Processor.Workers().ProcessFromFediAPI(suite.T().Context(), &messages.FromFediAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityCreate,
		GTSModel:       announceStatus,
		Receiving:      suite.testAccounts["local_account_1"],
		Requesting:     boostingAccount,
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Wait for side effects to trigger:
	// 1. status should have an ID, and be in the database
	if !testrig.WaitFor(func() bool {
		if announceStatus.ID == "" {
			return false
		}

		_, err = testStructs.State.DB.GetStatusByID(
			suite.T().Context(),
			announceStatus.ID,
		)
		return err == nil
	}) {
		suite.FailNow("timed out waiting for announce to be in the database")
	}

	// 2. a notification should exist for the announce
	where := []db.Where{
		{
			Key:   "status_id",
			Value: announceStatus.ID,
		},
	}
	notif := &gtsmodel.Notification{}
	err = testStructs.State.DB.GetWhere(suite.T().Context(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationReblog, notif.NotificationType)
	suite.Equal(boostedStatus.AccountID, notif.TargetAccountID)
	suite.Equal(announceStatus.AccountID, notif.OriginAccountID)
	suite.Equal(announceStatus.ID, notif.StatusOrEditID)
	suite.False(*notif.Read)
}

func (suite *FromFediAPITestSuite) TestProcessReplyMention() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	repliedAccount := &gtsmodel.Account{}
	*repliedAccount = *suite.testAccounts["local_account_1"]

	repliedStatus := &gtsmodel.Status{}
	*repliedStatus = *suite.testStatuses["local_account_1_status_1"]

	replyingAccount := &gtsmodel.Account{}
	*replyingAccount = *suite.testAccounts["remote_account_1"]

	// Set the replyingAccount's last fetched_at
	// date to something recent so no refresh is attempted,
	// and ensure it isn't a suspended account.
	replyingAccount.FetchedAt = time.Now()
	replyingAccount.SuspendedAt = time.Time{}
	replyingAccount.SuspensionOrigin = ""
	err := testStructs.State.DB.UpdateAccount(suite.T().Context(),
		replyingAccount,
		"fetched_at",
		"suspended_at",
		"suspension_origin",
	)
	suite.NoError(err)

	// Get replying statusable to use from remote test statuses.
	const replyingURI = "http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552"
	replyingStatusable := testrig.NewTestFediStatuses()[replyingURI]
	ap.AppendInReplyTo(replyingStatusable, testrig.URLMustParse(repliedStatus.URI))

	// Open a websocket stream to later test the streamed status reply.
	wssStream, errWithCode := testStructs.Processor.Stream().Open(suite.T().Context(), repliedAccount, stream.TimelineHome)
	suite.NoError(errWithCode)

	// Send the replied status off to the fedi worker to be further processed.
	err = testStructs.Processor.Workers().ProcessFromFediAPI(suite.T().Context(), &messages.FromFediAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		APObject:       replyingStatusable,
		Receiving:      repliedAccount,
		Requesting:     replyingAccount,
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Wait for side effects to trigger:
	// 1. status should be in the database
	var replyingStatus *gtsmodel.Status
	if !testrig.WaitFor(func() bool {
		replyingStatus, err = testStructs.State.DB.GetStatusByURI(suite.T().Context(), replyingURI)
		return err == nil
	}) {
		suite.FailNow("timed out waiting for replying status to be in the database")
	}

	// 2. a notification should exist for the mention
	var notif gtsmodel.Notification
	err = testStructs.State.DB.GetWhere(suite.T().Context(), []db.Where{
		{Key: "status_id", Value: replyingStatus.ID},
	}, &notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationMention, notif.NotificationType)
	suite.Equal(replyingStatus.InReplyToAccountID, notif.TargetAccountID)
	suite.Equal(replyingStatus.AccountID, notif.OriginAccountID)
	suite.Equal(replyingStatus.ID, notif.StatusOrEditID)
	suite.False(*notif.Read)

	ctx, _ := context.WithTimeout(suite.T().Context(), time.Second*5)
	msg, ok := wssStream.Recv(ctx)
	suite.True(ok)

	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)
	notifStreamed := &apimodel.Notification{}
	err = json.Unmarshal([]byte(msg.Payload), notifStreamed)
	suite.NoError(err)
	suite.Equal("mention", notifStreamed.Type)
	suite.Equal(replyingAccount.ID, notifStreamed.Account.ID)
}

func (suite *FromFediAPITestSuite) TestProcessFave() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	favedAccount := suite.testAccounts["local_account_1"]
	favedStatus := suite.testStatuses["local_account_1_status_1"]
	favingAccount := suite.testAccounts["remote_account_1"]

	wssStream, errWithCode := testStructs.Processor.Stream().Open(suite.T().Context(), favedAccount, stream.TimelineNotifications)
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

	err := testStructs.State.DB.Put(suite.T().Context(), fave)
	suite.NoError(err)

	err = testStructs.Processor.Workers().ProcessFromFediAPI(suite.T().Context(), &messages.FromFediAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityCreate,
		GTSModel:       fave,
		Receiving:      favedAccount,
		Requesting:     favingAccount,
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
	err = testStructs.State.DB.GetWhere(suite.T().Context(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationFavourite, notif.NotificationType)
	suite.Equal(fave.TargetAccountID, notif.TargetAccountID)
	suite.Equal(fave.AccountID, notif.OriginAccountID)
	suite.Equal(fave.StatusID, notif.StatusOrEditID)
	suite.False(*notif.Read)

	ctx, _ := context.WithTimeout(suite.T().Context(), time.Second*5)
	msg, ok := wssStream.Recv(ctx)
	suite.True(ok)

	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineNotifications}, msg.Stream)
}

// TestProcessFaveWithDifferentReceivingAccount ensures that when an account receives a fave that's for
// another account in their AP inbox, a notification isn't streamed to the receiving account.
//
// This tests for an issue we were seeing where Misskey sends out faves to inboxes of people that don't own
// the fave, but just follow the actor who received the fave.
func (suite *FromFediAPITestSuite) TestProcessFaveWithDifferentReceivingAccount() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	receivingAccount := suite.testAccounts["local_account_2"]
	favedAccount := suite.testAccounts["local_account_1"]
	favedStatus := suite.testStatuses["local_account_1_status_1"]
	favingAccount := suite.testAccounts["remote_account_1"]

	wssStream, errWithCode := testStructs.Processor.Stream().Open(suite.T().Context(), receivingAccount, stream.TimelineHome)
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

	err := testStructs.State.DB.Put(suite.T().Context(), fave)
	suite.NoError(err)

	err = testStructs.Processor.Workers().ProcessFromFediAPI(suite.T().Context(), &messages.FromFediAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityCreate,
		GTSModel:       fave,
		Receiving:      receivingAccount,
		Requesting:     favingAccount,
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
	err = testStructs.State.DB.GetWhere(suite.T().Context(), where, notif)
	suite.NoError(err)
	suite.Equal(gtsmodel.NotificationFavourite, notif.NotificationType)
	suite.Equal(fave.TargetAccountID, notif.TargetAccountID)
	suite.Equal(fave.AccountID, notif.OriginAccountID)
	suite.Equal(fave.StatusID, notif.StatusOrEditID)
	suite.False(*notif.Read)

	// 2. no notification should be streamed to the account that received the fave message, because they weren't the target
	ctx, _ := context.WithTimeout(suite.T().Context(), time.Second*5)
	_, ok := wssStream.Recv(ctx)
	suite.False(ok)
}

func (suite *FromFediAPITestSuite) TestProcessAccountDelete() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	ctx := suite.T().Context()

	deletedAccount := &gtsmodel.Account{}
	*deletedAccount = *suite.testAccounts["remote_account_1"]

	receivingAccount := &gtsmodel.Account{}
	*receivingAccount = *suite.testAccounts["local_account_1"]

	// before doing the delete....
	// make local_account_1 and remote_account_1 into mufos
	zorkFollowSatan := &gtsmodel.Follow{
		ID:              "01FGRY72ASHBSET64353DPHK9T",
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		UpdatedAt:       time.Now().Add(-1 * time.Hour),
		AccountID:       deletedAccount.ID,
		TargetAccountID: receivingAccount.ID,
		ShowReblogs:     util.Ptr(true),
		URI:             fmt.Sprintf("%s/follows/01FGRY72ASHBSET64353DPHK9T", deletedAccount.URI),
		Notify:          util.Ptr(false),
	}
	err := testStructs.State.DB.Put(ctx, zorkFollowSatan)
	suite.NoError(err)

	satanFollowZork := &gtsmodel.Follow{
		ID:              "01FGRYAVAWWPP926J175QGM0WV",
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		UpdatedAt:       time.Now().Add(-1 * time.Hour),
		AccountID:       receivingAccount.ID,
		TargetAccountID: deletedAccount.ID,
		ShowReblogs:     util.Ptr(true),
		URI:             fmt.Sprintf("%s/follows/01FGRYAVAWWPP926J175QGM0WV", receivingAccount.URI),
		Notify:          util.Ptr(false),
	}
	err = testStructs.State.DB.Put(ctx, satanFollowZork)
	suite.NoError(err)

	// now they are mufos!
	err = testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityDelete,
		GTSModel:       deletedAccount,
		Receiving:      receivingAccount,
		Requesting:     deletedAccount,
	})
	suite.NoError(err)

	// local account 2 blocked foss_satan, that block should be gone now
	testBlock := suite.testBlocks["local_account_2_block_remote_account_1"]
	dbBlock := &gtsmodel.Block{}
	err = testStructs.State.DB.GetByID(ctx, testBlock.ID, dbBlock)
	suite.ErrorIs(err, db.ErrNoEntries)

	// the mufos should be gone now too
	satanFollowsZork, err := testStructs.State.DB.IsFollowing(ctx, deletedAccount.ID, receivingAccount.ID)
	suite.NoError(err)
	suite.False(satanFollowsZork)
	zorkFollowsSatan, err := testStructs.State.DB.IsFollowing(ctx, receivingAccount.ID, deletedAccount.ID)
	suite.NoError(err)
	suite.False(zorkFollowsSatan)

	// no statuses from foss satan should be left in the database
	if !testrig.WaitFor(func() bool {
		s, err := testStructs.State.DB.GetAccountStatuses(ctx, deletedAccount.ID, 0, false, false, "", "", false, false)
		return s == nil && err == db.ErrNoEntries
	}) {
		suite.FailNow("timeout waiting for statuses to be deleted")
	}

	var dbAccount *gtsmodel.Account

	// account data should be zeroed.
	if !testrig.WaitFor(func() bool {
		dbAccount, err = testStructs.State.DB.GetAccountByID(ctx, deletedAccount.ID)
		return err == nil && dbAccount.DisplayName == ""
	}) {
		suite.FailNow("timeout waiting for statuses to be deleted")
	}

	suite.Empty(dbAccount.Note)
	suite.Empty(dbAccount.DisplayName)
	suite.Empty(dbAccount.AvatarMediaAttachmentID)
	suite.Empty(dbAccount.AvatarRemoteURL)
	suite.Empty(dbAccount.HeaderMediaAttachmentID)
	suite.Empty(dbAccount.HeaderRemoteURL)
	suite.Empty(dbAccount.Fields)
	suite.False(*dbAccount.Discoverable)
	suite.WithinDuration(time.Now(), dbAccount.SuspendedAt, 30*time.Second)
	suite.Equal(dbAccount.ID, dbAccount.SuspensionOrigin)
}

func (suite *FromFediAPITestSuite) TestProcessFollowRequestLocked() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	ctx := suite.T().Context()

	originAccount := suite.testAccounts["remote_account_1"]

	// target is a locked account
	targetAccount := suite.testAccounts["local_account_2"]

	wssStream, errWithCode := testStructs.Processor.Stream().Open(suite.T().Context(), targetAccount, stream.TimelineHome)
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
		ShowReblogs:     util.Ptr(true),
		URI:             fmt.Sprintf("%s/follows/01FGRYAVAWWPP926J175QGM0WV", originAccount.URI),
		Notify:          util.Ptr(false),
	}

	err := testStructs.State.DB.Put(ctx, satanFollowRequestTurtle)
	suite.NoError(err)

	err = testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityCreate,
		GTSModel:       satanFollowRequestTurtle,
		Receiving:      targetAccount,
		Requesting:     originAccount,
	})
	suite.NoError(err)

	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	msg, ok := wssStream.Recv(suite.T().Context())
	suite.True(ok)

	suite.Equal(stream.EventTypeNotification, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)
	notif := &apimodel.Notification{}
	err = json.Unmarshal([]byte(msg.Payload), notif)
	suite.NoError(err)
	suite.Equal("follow_request", notif.Type)
	suite.Equal(originAccount.ID, notif.Account.ID)

	// no messages should have been sent out, since we didn't need to federate an accept
	suite.Empty(testStructs.HTTPClient.SentMessages)
}

func (suite *FromFediAPITestSuite) TestProcessFollowRequestUnlocked() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	ctx := suite.T().Context()

	originAccount := suite.testAccounts["remote_account_1"]

	// target is an unlocked account
	targetAccount := suite.testAccounts["local_account_1"]

	wssStream, errWithCode := testStructs.Processor.Stream().Open(suite.T().Context(), targetAccount, stream.TimelineHome)
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
		ShowReblogs:     util.Ptr(true),
		URI:             fmt.Sprintf("%s/follows/01FGRYAVAWWPP926J175QGM0WV", originAccount.URI),
		Notify:          util.Ptr(false),
	}

	err := testStructs.State.DB.Put(ctx, satanFollowRequestTurtle)
	suite.NoError(err)

	err = testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityCreate,
		GTSModel:       satanFollowRequestTurtle,
		Receiving:      targetAccount,
		Requesting:     originAccount,
	})
	suite.NoError(err)

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

	// an accept message should be sent to satan's inbox
	var sent []byte
	if !testrig.WaitFor(func() bool {
		delivery, ok := testStructs.State.Workers.Delivery.Queue.Pop()
		if !ok {
			return false
		}
		if !testrig.EqualRequestURIs(delivery.Request.URL, *originAccount.SharedInboxURI) {
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

	suite.Equal(targetAccount.URI, accept.Actor)
	suite.Equal(originAccount.URI, accept.Object.Actor)
	suite.Equal(satanFollowRequestTurtle.URI, accept.Object.ID)
	suite.Equal(targetAccount.URI, accept.Object.Object)
	suite.Equal(targetAccount.URI, accept.Object.To)
	suite.Equal("Follow", accept.Object.Type)
	suite.Equal(originAccount.URI, accept.To)
	suite.Equal("Accept", accept.Type)

	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	msg, ok := wssStream.Recv(suite.T().Context())
	suite.True(ok)

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
func (suite *FromFediAPITestSuite) TestCreateStatusFromIRI() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	ctx := suite.T().Context()

	receivingAccount := suite.testAccounts["local_account_1"]
	statusCreator := suite.testAccounts["remote_account_2"]

	err := testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       nil, // gtsmodel is nil because this is a forwarded status -- we want to dereference it using the iri
		Receiving:      receivingAccount,
		Requesting:     statusCreator,
		APIRI:          testrig.URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1"),
	})
	suite.NoError(err)

	// status should now be in the database, attributed to remote_account_2
	s, err := testStructs.State.DB.GetStatusByURI(suite.T().Context(), "http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1")
	suite.NoError(err)
	suite.Equal(statusCreator.URI, s.AccountURI)
}

func (suite *FromFediAPITestSuite) TestMoveAccount() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	// We're gonna migrate foss_satan to our local admin account.
	ctx := suite.T().Context()
	receivingAcct := suite.testAccounts["local_account_1"]

	// Copy requesting and target accounts
	// since we'll be changing these.
	requestingAcct := &gtsmodel.Account{}
	*requestingAcct = *suite.testAccounts["remote_account_1"]
	targetAcct := &gtsmodel.Account{}
	*targetAcct = *suite.testAccounts["admin_account"]

	// Set alsoKnownAs on the admin account.
	targetAcct.AlsoKnownAsURIs = []string{requestingAcct.URI}
	if err := testStructs.State.DB.UpdateAccount(ctx, targetAcct, "also_known_as_uris"); err != nil {
		suite.FailNow(err.Error())
	}

	// Remove existing follow from zork to admin account.
	if err := testStructs.State.DB.DeleteFollowByID(
		ctx,
		suite.testFollows["local_account_1_admin_account"].ID,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Have Zork follow foss_satan instead.
	if err := testStructs.State.DB.PutFollow(ctx, &gtsmodel.Follow{
		ID:              "01HRA0XZYFZC5MNWTKEBR58SSE",
		URI:             "http://localhost:8080/users/the_mighty_zork/follows/01HRA0XZYFZC5MNWTKEBR58SSE",
		AccountID:       receivingAcct.ID,
		TargetAccountID: requestingAcct.ID,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the Move.
	err := testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityMove,
		GTSModel: &gtsmodel.Move{
			OriginURI: requestingAcct.URI,
			Origin:    testrig.URLMustParse(requestingAcct.URI),
			TargetURI: targetAcct.URI,
			Target:    testrig.URLMustParse(targetAcct.URI),
			URI:       "https://fossbros-anonymous.io/users/foss_satan/moves/01HRA064871MR8HGVSAFJ333GM",
		},
		Receiving:  receivingAcct,
		Requesting: requestingAcct,
	})
	suite.NoError(err)

	// Wait for side effects to trigger:
	// Zork should now be following admin account.
	if !testrig.WaitFor(func() bool {
		follows, err := testStructs.State.DB.IsFollowing(ctx, receivingAcct.ID, targetAcct.ID)
		if err != nil {
			suite.FailNow(err.Error())
		}
		return follows
	}) {
		suite.FailNow("timed out waiting for zork to follow admin account")
	}

	// Move should be in the DB.
	move, err := testStructs.State.DB.GetMoveByURI(ctx, "https://fossbros-anonymous.io/users/foss_satan/moves/01HRA064871MR8HGVSAFJ333GM")
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Move should be marked as completed.
	suite.WithinDuration(time.Now(), move.SucceededAt, 1*time.Minute)
}

func (suite *FromFediAPITestSuite) TestUndoAnnounce() {
	var (
		ctx            = suite.T().Context()
		testStructs    = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		requestingAcct = suite.testAccounts["remote_account_1"]
		receivingAcct  = suite.testAccounts["local_account_1"]
		boostedStatus  = suite.testStatuses["admin_account_status_1"]
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Have remote_account_1 boost admin_account.
	boost, err := testStructs.TypeConverter.StatusToBoost(
		ctx,
		boostedStatus,
		requestingAcct,
		"",
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Set the boost URI + URL to
	// fossbros-anonymous.io.
	boost.URI = "https://fossbros-anonymous.io/users/foss_satan/" + boost.ID
	boost.URL = boost.URI

	// Store the boost.
	if err := testStructs.State.DB.PutStatus(ctx, boost); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the Undo.
	err = testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityUndo,
		GTSModel:       boost,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})
	suite.NoError(err)

	// Wait for side effects to trigger:
	// the boost should be deleted.
	if !testrig.WaitFor(func() bool {
		_, err := testStructs.State.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			boost.ID,
		)
		return errors.Is(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for boost to be removed")
	}
}

func (suite *FromFediAPITestSuite) TestUpdateNote() {
	var (
		ctx            = suite.T().Context()
		testStructs    = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		requestingAcct = suite.testAccounts["remote_account_2"]
		receivingAcct  = suite.testAccounts["local_account_1"]
	)
	defer testrig.TearDownTestStructs(testStructs)

	update := testrig.NewTestActivities(suite.testAccounts)["remote_account_2_status_1_update"]
	statusable := update.Activity.GetActivityStreamsObject().At(0).GetActivityStreamsNote()
	noteURI := ap.GetJSONLDId(statusable)

	// Get the OG status.
	status, err := testStructs.State.DB.GetStatusByURI(ctx, noteURI.String())
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Process the Update.
	err = testStructs.Processor.Workers().ProcessFromFediAPI(ctx, &messages.FromFediAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       status, // original status
		APObject:       (ap.Statusable)(statusable),
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})
	suite.NoError(err)

	// Wait for side effects to trigger:
	// zork should have a mention notif.
	if !testrig.WaitFor(func() bool {
		_, err := testStructs.State.DB.GetNotification(
			gtscontext.SetBarebones(ctx),
			gtsmodel.NotificationMention,
			receivingAcct.ID,
			requestingAcct.ID,
			status.ID,
		)
		return err == nil
	}) {
		suite.FailNow("timed out waiting for mention notif")
	}
}

func TestFromFederatorTestSuite(t *testing.T) {
	suite.Run(t, &FromFediAPITestSuite{})
}
