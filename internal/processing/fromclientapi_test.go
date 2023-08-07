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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FromClientAPITestSuite struct {
	ProcessingStandardTestSuite
}

// This test ensures that when admin_account posts a new
// status, it ends up in the correct streaming timelines
// of local_account_1, which follows it.
func (suite *FromClientAPITestSuite) TestProcessStreamNewStatus() {
	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		testList         = suite.testLists["local_account_1_list_1"]
		streams          = suite.openStreams(ctx, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]
	)

	// Make a new status from admin account.
	newStatus := &gtsmodel.Status{
		ID:                       "01FN4B2F88TF9676DYNXWE1WSS",
		URI:                      "http://localhost:8080/users/admin/statuses/01FN4B2F88TF9676DYNXWE1WSS",
		URL:                      "http://localhost:8080/@admin/statuses/01FN4B2F88TF9676DYNXWE1WSS",
		Content:                  "this status should stream :)",
		AttachmentIDs:            []string{},
		TagIDs:                   []string{},
		MentionIDs:               []string{},
		EmojiIDs:                 []string{},
		CreatedAt:                testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		UpdatedAt:                testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		Local:                    util.Ptr(true),
		AccountURI:               "http://localhost:8080/users/admin",
		AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
		InReplyToID:              "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
		Federated:                util.Ptr(false),
		Boostable:                util.Ptr(true),
		Replyable:                util.Ptr(true),
		Likeable:                 util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}

	// Put the status in the db first, to mimic what
	// would have already happened earlier up the flow.
	if err := suite.db.PutStatus(ctx, newStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := suite.processor.ProcessFromClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       newStatus,
		OriginAccount:  postingAccount,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// Check message in home stream.
	homeMsg := <-homeStream.Messages
	suite.Equal(stream.EventTypeUpdate, homeMsg.Event)
	suite.EqualValues([]string{stream.TimelineHome}, homeMsg.Stream)
	suite.Empty(homeStream.Messages) // Stream should now be empty.

	// Check status from home stream.
	homeStreamStatus := &apimodel.Status{}
	if err := json.Unmarshal([]byte(homeMsg.Payload), homeStreamStatus); err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(newStatus.ID, homeStreamStatus.ID)
	suite.Equal(newStatus.Content, homeStreamStatus.Content)

	// Check message in list stream.
	listMsg := <-listStream.Messages
	suite.Equal(stream.EventTypeUpdate, listMsg.Event)
	suite.EqualValues([]string{stream.TimelineList + ":" + testList.ID}, listMsg.Stream)
	suite.Empty(listStream.Messages) // Stream should now be empty.

	// Check status from list stream.
	listStreamStatus := &apimodel.Status{}
	if err := json.Unmarshal([]byte(listMsg.Payload), listStreamStatus); err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(newStatus.ID, listStreamStatus.ID)
	suite.Equal(newStatus.Content, listStreamStatus.Content)
}

func (suite *FromClientAPITestSuite) TestProcessStatusDelete() {
	var (
		ctx                  = context.Background()
		deletingAccount      = suite.testAccounts["local_account_1"]
		receivingAccount     = suite.testAccounts["local_account_2"]
		deletedStatus        = suite.testStatuses["local_account_1_status_1"]
		boostOfDeletedStatus = suite.testStatuses["admin_account_status_4"]
		streams              = suite.openStreams(ctx, receivingAccount, nil)
		homeStream           = streams[stream.TimelineHome]
	)

	// Delete the status from the db first, to mimic what
	// would have already happened earlier up the flow
	if err := suite.db.DeleteStatusByID(ctx, deletedStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the status delete.
	if err := suite.processor.ProcessFromClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityDelete,
		GTSModel:       deletedStatus,
		OriginAccount:  deletingAccount,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// Stream should have the delete of admin's boost in it now.
	msg := <-homeStream.Messages
	suite.Equal(stream.EventTypeDelete, msg.Event)
	suite.Equal(boostOfDeletedStatus.ID, msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)

	// Stream should also have the delete of the message itself in it.
	msg = <-homeStream.Messages
	suite.Equal(stream.EventTypeDelete, msg.Event)
	suite.Equal(deletedStatus.ID, msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)

	// Stream should now be empty.
	suite.Empty(homeStream.Messages)

	// Boost should no longer be in the database.
	if !testrig.WaitFor(func() bool {
		_, err := suite.db.GetStatusByID(ctx, boostOfDeletedStatus.ID)
		return errors.Is(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for status delete")
	}
}

func (suite *FromClientAPITestSuite) TestProcessNewStatusWithNotification() {
	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		streams          = suite.openStreams(ctx, receivingAccount, nil)
		notifStream      = streams[stream.TimelineNotifications]
	)

	// Update the follow from receiving account -> posting account so
	// that receiving account wants notifs when posting account posts.
	follow := &gtsmodel.Follow{}
	*follow = *suite.testFollows["local_account_1_admin_account"]
	follow.Notify = util.Ptr(true)
	if err := suite.db.UpdateFollow(ctx, follow); err != nil {
		suite.FailNow(err.Error())
	}

	// Make a new status from admin account.
	newStatus := &gtsmodel.Status{
		ID:                       "01FN4B2F88TF9676DYNXWE1WSS",
		URI:                      "http://localhost:8080/users/admin/statuses/01FN4B2F88TF9676DYNXWE1WSS",
		URL:                      "http://localhost:8080/@admin/statuses/01FN4B2F88TF9676DYNXWE1WSS",
		Content:                  "this status should create a notification",
		AttachmentIDs:            []string{},
		TagIDs:                   []string{},
		MentionIDs:               []string{},
		EmojiIDs:                 []string{},
		CreatedAt:                testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		UpdatedAt:                testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		Local:                    util.Ptr(true),
		AccountURI:               "http://localhost:8080/users/admin",
		AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
		InReplyToID:              "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
		Federated:                util.Ptr(false),
		Boostable:                util.Ptr(true),
		Replyable:                util.Ptr(true),
		Likeable:                 util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}

	// Put the status in the db first, to mimic what
	// would have already happened earlier up the flow.
	if err := suite.db.PutStatus(ctx, newStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := suite.processor.ProcessFromClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       newStatus,
		OriginAccount:  postingAccount,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// Wait for a notification to appear for the status.
	if !testrig.WaitFor(func() bool {
		_, err := suite.db.GetNotification(
			ctx,
			gtsmodel.NotificationStatus,
			receivingAccount.ID,
			postingAccount.ID,
			newStatus.ID,
		)
		return err == nil
	}) {
		suite.FailNow("timed out waiting for new status notification")
	}

	// Check message in notification stream.
	notifMsg := <-notifStream.Messages
	suite.Equal(stream.EventTypeNotification, notifMsg.Event)
	suite.EqualValues([]string{stream.TimelineNotifications}, notifMsg.Stream)
	suite.Empty(notifStream.Messages) // Stream should now be empty.

	// Check notif.
	notif := &apimodel.Notification{}
	if err := json.Unmarshal([]byte(notifMsg.Payload), notif); err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(newStatus.ID, notif.Status.ID)
}

func TestFromClientAPITestSuite(t *testing.T) {
	suite.Run(t, &FromClientAPITestSuite{})
}
