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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type WorkersTestSuite struct {
	// standard suite interfaces
	suite.Suite

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testFollows      map[string]*gtsmodel.Follow
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testTags         map[string]*gtsmodel.Tag
	testMentions     map[string]*gtsmodel.Mention
	testAutheds      map[string]*oauth.Auth
	testBlocks       map[string]*gtsmodel.Block
	testActivities   map[string]testrig.ActivityWithSignature
	testLists        map[string]*gtsmodel.List
	testListEntries  map[string]*gtsmodel.ListEntry
}

// TestStructs encapsulates structs needed to
// run one test in this package. Each test should
// call SetupTestStructs to get a new TestStructs,
// and defer TearDownTestStructs to close it when
// the test is complete. The reason for doing things
// this way here is to prevent the tests in this
// package from overwriting one another's processors
// and worker queues, which was causing issues
// when running all tests at once.
type TestStructs struct {
	State         *state.State
	Processor     *processing.Processor
	HTTPClient    *testrig.MockHTTPClient
	TypeConverter *typeutils.Converter
	EmailSender   email.Sender
}

func (suite *WorkersTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testFollows = testrig.NewTestFollows()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
	suite.testAutheds = map[string]*oauth.Auth{
		"local_account_1": {
			Application: suite.testApplications["local_account_1"],
			User:        suite.testUsers["local_account_1"],
			Account:     suite.testAccounts["local_account_1"],
		},
	}
	suite.testBlocks = testrig.NewTestBlocks()
	suite.testLists = testrig.NewTestLists()
	suite.testListEntries = testrig.NewTestListEntries()
}

func (suite *WorkersTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
}

func (suite *WorkersTestSuite) openStreams(ctx context.Context, processor *processing.Processor, account *gtsmodel.Account, listIDs []string) map[string]*stream.Stream {
	streams := make(map[string]*stream.Stream)

	for _, streamType := range []string{
		stream.TimelineHome,
		stream.TimelinePublic,
		stream.TimelineNotifications,
		stream.TimelineDirect,
	} {
		stream, err := processor.Stream().Open(ctx, account, streamType)
		if err != nil {
			suite.FailNow(err.Error())
		}

		streams[streamType] = stream
	}

	for _, listID := range listIDs {
		streamType := stream.TimelineList + ":" + listID

		stream, err := processor.Stream().Open(ctx, account, streamType)
		if err != nil {
			suite.FailNow(err.Error())
		}

		streams[streamType] = stream
	}

	return streams
}

func (suite *WorkersTestSuite) SetupTestStructs() *TestStructs {
	state := state.State{}

	state.Caches.Init()

	db := testrig.NewTestDB(&state)
	state.DB = db

	storage := testrig.NewInMemoryStorage()
	state.Storage = storage
	typeconverter := typeutils.NewConverter(&state)

	testrig.StartTimelines(
		&state,
		visibility.NewFilter(&state),
		typeconverter,
	)

	httpClient := testrig.NewMockHTTPClient(nil, "../../../testrig/media")
	httpClient.TestRemotePeople = testrig.NewTestFediPeople()
	httpClient.TestRemoteStatuses = testrig.NewTestFediStatuses()

	transportController := testrig.NewTestTransportController(&state, httpClient)
	mediaManager := testrig.NewTestMediaManager(&state)
	federator := testrig.NewTestFederator(&state, transportController, mediaManager)
	oauthServer := testrig.NewTestOauthServer(db)
	emailSender := testrig.NewEmailSender("../../../web/template/", nil)

	processor := processing.NewProcessor(
		cleaner.New(&state),
		typeconverter,
		federator,
		oauthServer,
		mediaManager,
		&state,
		emailSender,
		visibility.NewFilter(&state),
		interaction.NewFilter(&state),
	)

	testrig.StartWorkers(&state, processor.Workers())

	testrig.StandardDBSetup(db, suite.testAccounts)
	testrig.StandardStorageSetup(storage, "../../../testrig/media")

	return &TestStructs{
		State:         &state,
		Processor:     processor,
		HTTPClient:    httpClient,
		TypeConverter: typeconverter,
		EmailSender:   emailSender,
	}
}

func (suite *WorkersTestSuite) TearDownTestStructs(testStructs *TestStructs) {
	testrig.StandardDBTeardown(testStructs.State.DB)
	testrig.StandardStorageTeardown(testStructs.State.Storage)
	testrig.StopWorkers(testStructs.State)
}
