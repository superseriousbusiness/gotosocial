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

package webpush_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/webpush"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type RealSenderStandardTestSuite struct {
	suite.Suite
	db                  db.DB
	storage             *storage.Driver
	state               state.State
	mediaManager        *media.Manager
	typeconverter       *typeutils.Converter
	httpClient          *testrig.MockHTTPClient
	transportController transport.Controller
	federator           *federation.Federator
	oauthServer         oauth.Server
	emailSender         email.Sender
	webPushSender       webpush.Sender

	// standard suite models
	testTokens               map[string]*gtsmodel.Token
	testClients              map[string]*gtsmodel.Client
	testApplications         map[string]*gtsmodel.Application
	testUsers                map[string]*gtsmodel.User
	testAccounts             map[string]*gtsmodel.Account
	testAttachments          map[string]*gtsmodel.MediaAttachment
	testStatuses             map[string]*gtsmodel.Status
	testTags                 map[string]*gtsmodel.Tag
	testMentions             map[string]*gtsmodel.Mention
	testEmojis               map[string]*gtsmodel.Emoji
	testNotifications        map[string]*gtsmodel.Notification
	testWebPushSubscriptions map[string]*gtsmodel.WebPushSubscription

	processor *processing.Processor

	webPushHttpClientDo func(request *http.Request) (*http.Response, error)
}

func (suite *RealSenderStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
	suite.testEmojis = testrig.NewTestEmojis()
	suite.testNotifications = testrig.NewTestNotifications()
	suite.testWebPushSubscriptions = testrig.NewTestWebPushSubscriptions()
}

func (suite *RealSenderStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.typeconverter = typeutils.NewConverter(&suite.state)

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		suite.typeconverter,
	)

	suite.httpClient = testrig.NewMockHTTPClient(nil, "../../testrig/media")
	suite.httpClient.TestRemotePeople = testrig.NewTestFediPeople()
	suite.httpClient.TestRemoteStatuses = testrig.NewTestFediStatuses()

	suite.transportController = testrig.NewTestTransportController(&suite.state, suite.httpClient)
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, suite.transportController, suite.mediaManager)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.emailSender = testrig.NewEmailSender("../../web/template/", nil)

	suite.webPushSender = webpush.NewRealSender(
		&http.Client{
			Transport: suite,
		},
		&suite.state,
	)

	suite.processor = processing.NewProcessor(
		cleaner.New(&suite.state),
		suite.typeconverter,
		suite.federator,
		suite.oauthServer,
		suite.mediaManager,
		&suite.state,
		suite.emailSender,
		suite.webPushSender,
		visibility.NewFilter(&suite.state),
		interaction.NewFilter(&suite.state),
	)
	testrig.StartWorkers(&suite.state, suite.processor.Workers())

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
}

func (suite *RealSenderStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
	suite.webPushHttpClientDo = nil
}

// RoundTrip implements http.RoundTripper with a closure stored in the test suite.
func (suite *RealSenderStandardTestSuite) RoundTrip(request *http.Request) (*http.Response, error) {
	return suite.webPushHttpClientDo(request)
}

// notifyingReadCloser is a zero-length io.ReadCloser that can tell us when it's been closed,
// indicating the simulated Web Push server response has been sent, received, read, and closed.
type notifyingReadCloser struct {
	bodyClosed chan struct{}
}

func (rc *notifyingReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (rc *notifyingReadCloser) Close() error {
	rc.bodyClosed <- struct{}{}
	close(rc.bodyClosed)
	return nil
}

func (suite *RealSenderStandardTestSuite) TestSendSuccess() {
	// Set a timeout on the whole test. If it fails due to the timeout,
	// the push notification was not sent for some reason.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notification, err := suite.state.DB.GetNotificationByID(ctx, suite.testNotifications["local_account_1_like"].ID)
	if !suite.NoError(err) {
		suite.FailNow("Couldn't fetch notification to send")
	}

	rc := &notifyingReadCloser{
		bodyClosed: make(chan struct{}, 1),
	}

	// Simulate a successful response from the Web Push server.
	suite.webPushHttpClientDo = func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       rc,
		}, nil
	}

	// Send the push notification.
	suite.NoError(suite.webPushSender.Send(ctx, notification, nil, nil))

	// Wait for it to be sent or for the context to time out.
	bodyClosed := false
	contextExpired := false
	select {
	case <-rc.bodyClosed:
		bodyClosed = true
	case <-ctx.Done():
		contextExpired = true
	}
	suite.True(bodyClosed)
	suite.False(contextExpired)
}

func TestRealSenderStandardTestSuite(t *testing.T) {
	suite.Run(t, &RealSenderStandardTestSuite{})
}
