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

package streaming_test

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/streaming"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type StreamingTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	tc           *typeutils.Converter
	mediaManager *media.Manager
	federator    *federation.Federator
	emailSender  email.Sender
	processor    *processing.Processor
	storage      *storage.Driver
	state        state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testFollows      map[string]*gtsmodel.Follow

	// module being tested
	streamingModule *streaming.Module
}

func (suite *StreamingTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFollows = testrig.NewTestFollows()
}

func (suite *StreamingTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	suite.tc = typeutils.NewConverter(&suite.state)

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../../testrig/media")), suite.mediaManager)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.streamingModule = streaming.New(suite.processor, 1, 4096)
}

func (suite *StreamingTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

// Addr is a fake network interface which implements the net.Addr interface
type Addr struct {
	NetworkString string
	AddrString    string
}

func (a Addr) Network() string {
	return a.NetworkString
}

func (a Addr) String() string {
	return a.AddrString
}

type connTester struct {
	deadline time.Time
	writes   int
}

func (c *connTester) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c *connTester) SetDeadline(t time.Time) error {
	c.deadline = t
	return nil
}

func (c *connTester) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *connTester) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *connTester) Write(p []byte) (int, error) {
	c.writes++
	if c.writes > 1 {
		return 0, errors.New("timeout")
	}
	return 0, nil
}

func (c *connTester) Close() error {
	return nil
}

func (c *connTester) LocalAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

func (c *connTester) RemoteAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	w            gin.ResponseWriter
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) closeClient() {
	r.closeChannel <- true
}

func (r *TestResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn := &connTester{
		writes: 0,
	}
	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return conn, brw, nil
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	w := new(gin.ResponseWriter)
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		*w,
		make(chan bool, 1),
	}
}

func (suite *StreamingTestSuite) TestSecurityHeader() {
	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	recorder := CreateTestResponseRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/%s?stream=user", streaming.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Header.Set(streaming.AccessTokenHeader, oauthToken.Access)
	ctx.Request.Header.Set("Connection", "upgrade")
	ctx.Request.Header.Set("Upgrade", "websocket")
	ctx.Request.Header.Set("Sec-Websocket-Version", "13")
	key := [16]byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'}
	key64 := base64.StdEncoding.EncodeToString(key[:]) // sec-websocket-key must be base64 encoded and 16 bytes long
	ctx.Request.Header.Set("Sec-Websocket-Key", key64)

	suite.streamingModule.StreamGETHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	// check response
	if !suite.EqualValues(http.StatusOK, recorder.Code) {
		suite.T().Logf("%s", b)
	}
}

func TestStreamingTestSuite(t *testing.T) {
	suite.Run(t, new(StreamingTestSuite))
}
