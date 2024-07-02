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

package conversations_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing/conversations"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ConversationsTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	tc                  *typeutils.Converter
	storage             *storage.Driver
	state               state.State
	mediaManager        *media.Manager
	transportController transport.Controller
	federator           *federation.Federator
	emailSender         email.Sender
	sentEmails          map[string]string

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testFollows      map[string]*gtsmodel.Follow
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status

	// module being tested
	conversationsProcessor conversations.Processor

	// conversation created for test
	testAccount      *gtsmodel.Account
	testConversation *gtsmodel.Conversation
	testNow          time.Time
}

func (suite *ConversationsTestSuite) getClientMsg(timeout time.Duration) (*messages.FromClientAPI, bool) {
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, timeout)
	defer cncl()
	return suite.state.Workers.Client.Queue.PopCtx(ctx)
}

func (suite *ConversationsTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testFollows = testrig.NewTestFollows()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *ConversationsTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.tc = typeutils.NewConverter(&suite.state)

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		suite.tc,
	)

	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)

	suite.transportController = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../testrig/media"))
	suite.federator = testrig.NewTestFederator(&suite.state, suite.transportController, suite.mediaManager)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", suite.sentEmails)

	suite.conversationsProcessor = conversations.New(&suite.state, suite.tc)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")

	suite.testNow = time.Now()
	suite.testAccount = suite.testAccounts["local_account_1"]
	suite.testConversation = suite.newTestConversation(0)
}

func (suite *ConversationsTestSuite) TearDownTest() {
	conversationModels := []interface{}{
		(*gtsmodel.Conversation)(nil),
		(*gtsmodel.ConversationToStatus)(nil),
	}
	for _, model := range conversationModels {
		if err := suite.db.DropTable(context.Background(), model); err != nil {
			log.Error(context.Background(), err)
		}
	}

	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *ConversationsTestSuite) newTestStatus(threadID string, nowOffset time.Duration, inReplyToStatus *gtsmodel.Status) *gtsmodel.Status {
	statusID := id.NewULID()
	createdAt := suite.testNow.Add(nowOffset)
	status := &gtsmodel.Status{
		ID:                  statusID,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		URI:                 "http://localhost:8080/users/" + suite.testAccount.Username + "/statuses/" + statusID,
		AccountID:           suite.testAccount.ID,
		AccountURI:          suite.testAccount.URI,
		Local:               util.Ptr(true),
		ThreadID:            threadID,
		Visibility:          gtsmodel.VisibilityDirect,
		ActivityStreamsType: ap.ObjectNote,
		Federated:           util.Ptr(true),
		Boostable:           util.Ptr(true),
		Replyable:           util.Ptr(true),
		Likeable:            util.Ptr(true),
	}
	if inReplyToStatus != nil {
		status.InReplyToID = inReplyToStatus.ID
		status.InReplyToURI = inReplyToStatus.URI
		status.InReplyToAccountID = inReplyToStatus.AccountID
	}
	if err := suite.db.PutStatus(context.Background(), status); err != nil {
		suite.FailNow(err.Error())
	}
	return status
}

// newTestConversation creates a new status and adds it to a new unread conversation, returning the conversation.
func (suite *ConversationsTestSuite) newTestConversation(nowOffset time.Duration) *gtsmodel.Conversation {
	status := suite.newTestStatus(id.NewULID(), nowOffset, nil)
	conversation := &gtsmodel.Conversation{
		ID:        id.NewULID(),
		AccountID: suite.testAccount.ID,
		ThreadID:  status.ThreadID,
		Read:      util.Ptr(false),
	}
	conversation, err := suite.db.AddStatusToConversation(context.Background(), conversation, status)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return conversation
}

func TestConversationsTestSuite(t *testing.T) {
	suite.Run(t, new(ConversationsTestSuite))
}
