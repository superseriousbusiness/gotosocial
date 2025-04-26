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

	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	dbtest "code.superseriousbusiness.org/gotosocial/internal/db/test"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/processing/conversations"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
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
	filter              *visibility.Filter

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testFollows      map[string]*gtsmodel.Follow
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status

	// module being tested
	conversationsProcessor conversations.Processor

	// Owner of test conversations
	testAccount *gtsmodel.Account

	// Mixin for conversation tests
	dbtest.ConversationFactory
}

func (suite *ConversationsTestSuite) getClientMsg(timeout time.Duration) (*messages.FromClientAPI, bool) {
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, timeout)
	defer cncl()
	return suite.state.Workers.Client.Queue.PopCtx(ctx)
}

func (suite *ConversationsTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testFollows = testrig.NewTestFollows()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()

	suite.ConversationFactory.SetupSuite(suite)
}

func (suite *ConversationsTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.tc = typeutils.NewConverter(&suite.state)
	suite.filter = visibility.NewFilter(&suite.state)

	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)

	suite.transportController = testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../testrig/media"))
	suite.federator = testrig.NewTestFederator(&suite.state, suite.transportController, suite.mediaManager)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", suite.sentEmails)

	suite.conversationsProcessor = conversations.New(&suite.state, suite.tc, suite.filter)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")

	suite.ConversationFactory.SetupTest(suite.db)

	suite.testAccount = suite.testAccounts["local_account_1"]
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

func TestConversationsTestSuite(t *testing.T) {
	suite.Run(t, new(ConversationsTestSuite))
}
