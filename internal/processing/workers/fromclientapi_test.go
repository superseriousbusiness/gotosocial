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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FromClientAPITestSuite struct {
	WorkersTestSuite
}

func (suite *FromClientAPITestSuite) newStatus(
	ctx context.Context,
	state *state.State,
	account *gtsmodel.Account,
	visibility gtsmodel.Visibility,
	policy *gtsmodel.InteractionPolicy,
	replyToStatus *gtsmodel.Status,
	boostOfStatus *gtsmodel.Status,
) *gtsmodel.Status {
	var (
		protocol = config.GetProtocol()
		host     = config.GetHost()
		statusID = id.NewULID()
	)

	// Make a new status from given account.
	newStatus := &gtsmodel.Status{
		ID:                  statusID,
		URI:                 protocol + "://" + host + "/users/" + account.Username + "/statuses/" + statusID,
		URL:                 protocol + "://" + host + "/@" + account.Username + "/statuses/" + statusID,
		Content:             "pee pee poo poo",
		Local:               util.Ptr(true),
		AccountURI:          account.URI,
		AccountID:           account.ID,
		Visibility:          visibility,
		ActivityStreamsType: ap.ObjectNote,
		Federated:           util.Ptr(true),
		InteractionPolicy:   policy,
	}

	if replyToStatus != nil {
		// Status is a reply.
		newStatus.InReplyToAccountID = replyToStatus.AccountID
		newStatus.InReplyToID = replyToStatus.ID
		newStatus.InReplyToURI = replyToStatus.URI
		newStatus.ThreadID = replyToStatus.ThreadID

		// Mention the replied-to account.
		mention := &gtsmodel.Mention{
			ID:               id.NewULID(),
			StatusID:         statusID,
			OriginAccountID:  account.ID,
			OriginAccountURI: account.URI,
			TargetAccountID:  replyToStatus.AccountID,
		}

		if err := state.DB.PutMention(ctx, mention); err != nil {
			suite.FailNow(err.Error())
		}
		newStatus.Mentions = []*gtsmodel.Mention{mention}
		newStatus.MentionIDs = []string{mention.ID}
	}

	if boostOfStatus != nil {
		// Status is a boost.
		newStatus.Content = ""
		newStatus.BoostOfAccountID = boostOfStatus.AccountID
		newStatus.BoostOfID = boostOfStatus.ID
		newStatus.Visibility = boostOfStatus.Visibility
	}

	// Put the status in the db, to mimic what would
	// have already happened earlier up the flow.
	if err := state.DB.PutStatus(ctx, newStatus); err != nil {
		suite.FailNow(err.Error())
	}

	return newStatus
}

func (suite *FromClientAPITestSuite) checkStreamed(
	str *stream.Stream,
	expectMessage bool,
	expectPayload string,
	expectEventType string,
) {

	// Set a 5s timeout on context.
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, time.Second*5)
	defer cncl()

	msg, ok := str.Recv(ctx)

	if expectMessage && !ok {
		suite.FailNow("expected a message but message was not received")
	}

	if !expectMessage && ok {
		suite.FailNow("expected no message but message was received")
	}

	if expectPayload != "" && msg.Payload != expectPayload {
		suite.FailNow("", "expected payload %s but payload was: %s", expectPayload, msg.Payload)
	}

	if expectEventType != "" && msg.Event != expectEventType {
		suite.FailNow("", "expected event type %s but event type was: %s", expectEventType, msg.Event)
	}
}

func (suite *FromClientAPITestSuite) statusJSON(
	ctx context.Context,
	typeConverter *typeutils.Converter,
	status *gtsmodel.Status,
	requestingAccount *gtsmodel.Account,
) string {
	apiStatus, err := typeConverter.StatusToAPIStatus(
		ctx,
		status,
		requestingAccount,
		statusfilter.FilterContextNone,
		nil,
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON, err := json.Marshal(apiStatus)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return string(statusJSON)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusWithNotification() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		testList         = suite.testLists["local_account_1_list_1"]
		streams          = suite.openStreams(ctx,
			testStructs.Processor,
			receivingAccount,
			[]string{testList.ID},
		)
		homeStream  = streams[stream.TimelineHome]
		listStream  = streams[stream.TimelineList+":"+testList.ID]
		notifStream = streams[stream.TimelineNotifications]

		// Admin account posts a new top-level status.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			nil,
			nil,
		)
	)

	// Update the follow from receiving account -> posting account so
	// that receiving account wants notifs when posting account posts.
	follow := new(gtsmodel.Follow)
	*follow = *suite.testFollows["local_account_1_admin_account"]

	follow.Notify = util.Ptr(true)
	if err := testStructs.State.DB.UpdateFollow(ctx, follow); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON := suite.statusJSON(
		ctx,
		testStructs.TypeConverter,
		status,
		receivingAccount,
	)

	// Check message in home stream.
	suite.checkStreamed(
		homeStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Check message in list stream.
	suite.checkStreamed(
		listStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Wait for a notification to appear for the status.
	var notif *gtsmodel.Notification
	if !testrig.WaitFor(func() bool {
		var err error
		notif, err = testStructs.State.DB.GetNotification(
			ctx,
			gtsmodel.NotificationStatus,
			receivingAccount.ID,
			postingAccount.ID,
			status.ID,
		)
		return err == nil
	}) {
		suite.FailNow("timed out waiting for new status notification")
	}

	apiNotif, err := testStructs.TypeConverter.NotificationToAPINotification(ctx, notif, nil, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}

	notifJSON, err := json.Marshal(apiNotif)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check message in notification stream.
	suite.checkStreamed(
		notifStream,
		true,
		string(notifJSON),
		stream.EventTypeNotification,
	)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusReply() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		testList         = suite.testLists["local_account_1_list_1"]
		streams          = suite.openStreams(ctx, testStructs.Processor, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]

		// Admin account posts a reply to turtle.
		// Since turtle is followed by zork, and
		// the default replies policy for this list
		// is to show replies to followed accounts,
		// post should also show in the list stream.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			suite.testStatuses["local_account_2_status_1"],
			nil,
		)
	)

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON := suite.statusJSON(
		ctx,
		testStructs.TypeConverter,
		status,
		receivingAccount,
	)

	// Check message in home stream.
	suite.checkStreamed(
		homeStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Check message in list stream.
	suite.checkStreamed(
		listStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusReplyMuted() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]

		// Admin account posts a reply to zork.
		// Normally zork would get a notification
		// for this, but zork mutes this thread.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			suite.testStatuses["local_account_1_status_1"],
			nil,
		)
		threadMute = &gtsmodel.ThreadMute{
			ID:        "01HD3KRMBB1M85QRWHD912QWRE",
			ThreadID:  suite.testStatuses["local_account_1_status_1"].ThreadID,
			AccountID: receivingAccount.ID,
		}
	)

	// Store the thread mute before processing new status.
	if err := testStructs.State.DB.PutThreadMute(ctx, threadMute); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Ensure no notification received.
	notif, err := testStructs.State.DB.GetNotification(
		ctx,
		gtsmodel.NotificationMention,
		receivingAccount.ID,
		postingAccount.ID,
		status.ID,
	)

	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(notif)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusBoostMuted() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]

		// Admin account boosts a status by zork.
		// Normally zork would get a notification
		// for this, but zork mutes this thread.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			nil,
			suite.testStatuses["local_account_1_status_1"],
		)
		threadMute = &gtsmodel.ThreadMute{
			ID:        "01HD3KRMBB1M85QRWHD912QWRE",
			ThreadID:  suite.testStatuses["local_account_1_status_1"].ThreadID,
			AccountID: receivingAccount.ID,
		}
	)

	// Store the thread mute before processing new status.
	if err := testStructs.State.DB.PutThreadMute(ctx, threadMute); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ActivityAnnounce,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Ensure no notification received.
	notif, err := testStructs.State.DB.GetNotification(
		ctx,
		gtsmodel.NotificationReblog,
		receivingAccount.ID,
		postingAccount.ID,
		status.ID,
	)

	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(notif)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusListRepliesPolicyListOnlyOK() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	// We're modifying the test list so take a copy.
	testList := new(gtsmodel.List)
	*testList = *suite.testLists["local_account_1_list_1"]

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		streams          = suite.openStreams(ctx, testStructs.Processor, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]

		// Admin account posts a reply to turtle.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			suite.testStatuses["local_account_2_status_1"],
			nil,
		)
	)

	// Modify replies policy of test list to show replies
	// only to other accounts in the same list. Since turtle
	// and admin are in the same list, this means the reply
	// should be shown in the list.
	testList.RepliesPolicy = gtsmodel.RepliesPolicyList
	if err := testStructs.State.DB.UpdateList(ctx, testList, "replies_policy"); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON := suite.statusJSON(
		ctx,
		testStructs.TypeConverter,
		status,
		receivingAccount,
	)

	// Check message in home stream.
	suite.checkStreamed(
		homeStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Check message in list stream.
	suite.checkStreamed(
		listStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusListRepliesPolicyListOnlyNo() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	// We're modifying the test list so take a copy.
	testList := new(gtsmodel.List)
	*testList = *suite.testLists["local_account_1_list_1"]

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		streams          = suite.openStreams(ctx, testStructs.Processor, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]

		// Admin account posts a reply to turtle.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			suite.testStatuses["local_account_2_status_1"],
			nil,
		)
	)

	// Modify replies policy of test list to show replies
	// only to other accounts in the same list. We're
	// about to remove turtle from the same list as admin,
	// so the new post should not be streamed to the list.
	testList.RepliesPolicy = gtsmodel.RepliesPolicyList
	if err := testStructs.State.DB.UpdateList(ctx, testList, "replies_policy"); err != nil {
		suite.FailNow(err.Error())
	}

	// Remove turtle from the list.
	if err := testStructs.State.DB.DeleteListEntry(ctx, suite.testListEntries["local_account_1_list_1_entry_1"].ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON := suite.statusJSON(
		ctx,
		testStructs.TypeConverter,
		status,
		receivingAccount,
	)

	// Check message in home stream.
	suite.checkStreamed(
		homeStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Check message NOT in list stream.
	suite.checkStreamed(
		listStream,
		false,
		"",
		"",
	)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusReplyListRepliesPolicyNone() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	// We're modifying the test list so take a copy.
	testList := new(gtsmodel.List)
	*testList = *suite.testLists["local_account_1_list_1"]

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		streams          = suite.openStreams(ctx, testStructs.Processor, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]

		// Admin account posts a reply to turtle.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			suite.testStatuses["local_account_2_status_1"],
			nil,
		)
	)

	// Modify replies policy of test list.
	// Since we're modifying the list to not
	// show any replies, the post should not
	// be streamed to the list.
	testList.RepliesPolicy = gtsmodel.RepliesPolicyNone
	if err := testStructs.State.DB.UpdateList(ctx, testList, "replies_policy"); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON := suite.statusJSON(
		ctx,
		testStructs.TypeConverter,
		status,
		receivingAccount,
	)

	// Check message in home stream.
	suite.checkStreamed(
		homeStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Check message NOT in list stream.
	suite.checkStreamed(
		listStream,
		false,
		"",
		"",
	)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusBoost() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		testList         = suite.testLists["local_account_1_list_1"]
		streams          = suite.openStreams(ctx, testStructs.Processor, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]

		// Admin account boosts a post by turtle.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			nil,
			suite.testStatuses["local_account_2_status_1"],
		)
	)

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ActivityAnnounce,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	statusJSON := suite.statusJSON(
		ctx,
		testStructs.TypeConverter,
		status,
		receivingAccount,
	)

	// Check message in home stream.
	suite.checkStreamed(
		homeStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)

	// Check message in list stream.
	suite.checkStreamed(
		listStream,
		true,
		statusJSON,
		stream.EventTypeUpdate,
	)
}

func (suite *FromClientAPITestSuite) TestProcessCreateStatusBoostNoReblogs() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx              = context.Background()
		postingAccount   = suite.testAccounts["admin_account"]
		receivingAccount = suite.testAccounts["local_account_1"]
		testList         = suite.testLists["local_account_1_list_1"]
		streams          = suite.openStreams(ctx, testStructs.Processor, receivingAccount, []string{testList.ID})
		homeStream       = streams[stream.TimelineHome]
		listStream       = streams[stream.TimelineList+":"+testList.ID]

		// Admin account boosts a post by turtle.
		status = suite.newStatus(
			ctx,
			testStructs.State,
			postingAccount,
			gtsmodel.VisibilityPublic,
			gtsmodel.DefaultInteractionPolicyPublic(),
			nil,
			suite.testStatuses["local_account_2_status_1"],
		)
	)

	// Update zork's follow of admin
	// to not show boosts in timeline.
	follow := new(gtsmodel.Follow)
	*follow = *suite.testFollows["local_account_1_admin_account"]
	follow.ShowReblogs = util.Ptr(false)
	if err := testStructs.State.DB.UpdateFollow(ctx, follow, "show_reblogs"); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the new status.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ActivityAnnounce,
			APActivityType: ap.ActivityCreate,
			GTSModel:       status,
			Origin:         postingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Check message NOT in home stream.
	suite.checkStreamed(
		homeStream,
		false,
		"",
		"",
	)

	// Check message NOT in list stream.
	suite.checkStreamed(
		listStream,
		false,
		"",
		"",
	)
}

func (suite *FromClientAPITestSuite) TestProcessStatusDelete() {
	testStructs := suite.SetupTestStructs()
	defer suite.TearDownTestStructs(testStructs)

	var (
		ctx                  = context.Background()
		deletingAccount      = suite.testAccounts["local_account_1"]
		receivingAccount     = suite.testAccounts["local_account_2"]
		deletedStatus        = suite.testStatuses["local_account_1_status_1"]
		boostOfDeletedStatus = suite.testStatuses["admin_account_status_4"]
		streams              = suite.openStreams(ctx, testStructs.Processor, receivingAccount, nil)
		homeStream           = streams[stream.TimelineHome]
	)

	// Delete the status from the db first, to mimic what
	// would have already happened earlier up the flow
	if err := testStructs.State.DB.DeleteStatusByID(ctx, deletedStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Process the status delete.
	if err := testStructs.Processor.Workers().ProcessFromClientAPI(
		ctx,
		&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityDelete,
			GTSModel:       deletedStatus,
			Origin:         deletingAccount,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Stream should have the delete
	// of admin's boost in it now.
	suite.checkStreamed(
		homeStream,
		true,
		boostOfDeletedStatus.ID,
		stream.EventTypeDelete,
	)

	// Stream should also have the delete
	// of the message itself in it.
	suite.checkStreamed(
		homeStream,
		true,
		deletedStatus.ID,
		stream.EventTypeDelete,
	)

	// Boost should no longer be in the database.
	if !testrig.WaitFor(func() bool {
		_, err := testStructs.State.DB.GetStatusByID(ctx, boostOfDeletedStatus.ID)
		return errors.Is(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for status delete")
	}
}

func TestFromClientAPITestSuite(t *testing.T) {
	suite.Run(t, &FromClientAPITestSuite{})
}
