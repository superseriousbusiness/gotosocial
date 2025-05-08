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

	// for go:linkname
	_ "unsafe"

	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/interaction"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/subscriptions"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/internal/webpush"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
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
	testAccounts             map[string]*gtsmodel.Account
	testNotifications        map[string]*gtsmodel.Notification
	testWebPushSubscriptions map[string]*gtsmodel.WebPushSubscription

	processor *processing.Processor

	webPushHttpClientDo func(request *http.Request) (*http.Response, error)
}

func (suite *RealSenderStandardTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
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

	suite.httpClient = testrig.NewMockHTTPClient(nil, "../../testrig/media")
	suite.httpClient.TestRemotePeople = testrig.NewTestFediPeople()
	suite.httpClient.TestRemoteStatuses = testrig.NewTestFediStatuses()

	suite.transportController = testrig.NewTestTransportController(&suite.state, suite.httpClient)
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, suite.transportController, suite.mediaManager)
	suite.oauthServer = testrig.NewTestOauthServer(&suite.state)
	suite.emailSender = testrig.NewEmailSender("../../web/template/", nil)

	suite.webPushSender = newSenderWith(
		&http.Client{
			Transport: suite,
		},
		&suite.state,
		suite.typeconverter,
	)

	suite.processor = processing.NewProcessor(
		cleaner.New(&suite.state),
		subscriptions.New(
			&suite.state,
			suite.transportController,
			suite.typeconverter,
		),
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

func (rc *notifyingReadCloser) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (rc *notifyingReadCloser) Close() error {
	rc.bodyClosed <- struct{}{}
	close(rc.bodyClosed)
	return nil
}

// Simulate sending a push notification with the suite's fake web client.
func (suite *RealSenderStandardTestSuite) simulatePushNotification(
	notificationID string,
	statusCode int,
	expectSend bool,
	expectDeletedSubscription bool,
) error {
	// Don't let the test run forever if the push notification was not sent for some reason.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	notification, err := suite.state.DB.GetNotificationByID(ctx, notificationID)
	if !suite.NoError(err) {
		suite.FailNow("Couldn't fetch notification to send")
	}

	rc := &notifyingReadCloser{
		bodyClosed: make(chan struct{}, 1),
	}

	// Simulate a response from the Web Push server.
	suite.webPushHttpClientDo = func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(statusCode),
			StatusCode: statusCode,
			Body:       rc,
		}, nil
	}

	// Send the push notification.
	sendError := suite.webPushSender.Send(ctx, notification, nil, nil)

	// Wait for it to be sent or for the context to time out.
	bodyClosed := false
	contextExpired := false
	select {
	case <-rc.bodyClosed:
		bodyClosed = true
	case <-ctx.Done():
		contextExpired = true
	}

	// In some cases we expect the notification *not* to be sent.
	if !expectSend {
		suite.False(bodyClosed)
		suite.True(contextExpired)
		return nil
	}

	suite.True(bodyClosed)
	suite.False(contextExpired)

	// Look for the associated Web Push subscription. Some server responses should delete it.
	subscription, err := suite.state.DB.GetWebPushSubscriptionByTokenID(
		ctx,
		suite.testWebPushSubscriptions["local_account_1_token_1"].TokenID,
	)
	if expectDeletedSubscription {
		suite.ErrorIs(err, db.ErrNoEntries)
	} else {
		suite.NotNil(subscription)
	}

	return sendError
}

// Test a successful response to sending a push notification.
func (suite *RealSenderStandardTestSuite) TestSendSuccess() {
	notificationID := suite.testNotifications["local_account_1_like"].ID
	suite.NoError(suite.simulatePushNotification(notificationID, http.StatusOK, true, false))
}

// Test a rate-limiting response to sending a push notification.
// This should not delete the subscription.
func (suite *RealSenderStandardTestSuite) TestRateLimited() {
	notificationID := suite.testNotifications["local_account_1_like"].ID
	suite.NoError(suite.simulatePushNotification(notificationID, http.StatusTooManyRequests, true, false))
}

// Test a non-special-cased client error response to sending a push notification.
// This should delete the subscription.
func (suite *RealSenderStandardTestSuite) TestClientError() {
	notificationID := suite.testNotifications["local_account_1_like"].ID
	suite.NoError(suite.simulatePushNotification(notificationID, http.StatusBadRequest, true, true))
}

// Test a server error response to sending a push notification.
// This should not delete the subscription.
func (suite *RealSenderStandardTestSuite) TestServerError() {
	notificationID := suite.testNotifications["local_account_1_like"].ID
	suite.NoError(suite.simulatePushNotification(notificationID, http.StatusInternalServerError, true, false))
}

// Don't send a push notification if it doesn't match policy.
func (suite *RealSenderStandardTestSuite) TestSendPolicyMismatch() {
	// Setup: create a new notification from an account that the subscribed account doesn't follow.
	notification := &gtsmodel.Notification{
		ID:               "01JJZ2Y9Z8E1XKT90EHZ5KZBDW",
		NotificationType: gtsmodel.NotificationFavourite,
		TargetAccountID:  suite.testAccounts["local_account_1"].ID,
		OriginAccountID:  suite.testAccounts["remote_account_1"].ID,
		StatusOrEditID:   "01F8MHAMCHF6Y650WCRSCP4WMY",
		Read:             util.Ptr(false),
	}
	if err := suite.db.PutNotification(context.Background(), notification); !suite.NoError(err) {
		suite.FailNow(err.Error())
		return
	}

	suite.NoError(suite.simulatePushNotification(notification.ID, 0, false, false))
}

func TestRealSenderStandardTestSuite(t *testing.T) {
	suite.Run(t, &RealSenderStandardTestSuite{})
}

//go:linkname newSenderWith code.superseriousbusiness.org/gotosocial/internal/webpush.newSenderWith
func newSenderWith(*http.Client, *state.State, *typeutils.Converter) webpush.Sender
