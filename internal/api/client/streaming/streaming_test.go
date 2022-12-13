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

package streaming_test

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/streaming"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StreamingTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	tc           typeutils.TypeConverter
	mediaManager media.Manager
	federator    federation.Federator
	emailSender  email.Sender
	processor    processing.Processor
	storage      *storage.Driver

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
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
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFollows = testrig.NewTestFollows()
}

func (suite *StreamingTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.storage = testrig.NewInMemoryStorage()
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)

	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../../testrig/media"), suite.db, fedWorker), suite.storage, suite.mediaManager, fedWorker)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender, suite.mediaManager, clientWorker, fedWorker)
	suite.streamingModule = streaming.NewWithTickDuration(suite.processor, 1).(*streaming.Module)
	suite.NoError(suite.processor.Start())
}

func (suite *StreamingTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
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
	ctx.Request.Header.Set("Sec-Websocket-Key", "abcd")

	suite.streamingModule.StreamGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	_, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
}

func TestStreamingTestSuite(t *testing.T) {
	suite.Run(t, new(StreamingTestSuite))
}
