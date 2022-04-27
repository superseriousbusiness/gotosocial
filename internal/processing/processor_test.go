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

package processing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"codeberg.org/gruf/go-store/kv"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/worker"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ProcessingStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	storage             *kv.KVStore
	mediaManager        media.Manager
	typeconverter       typeutils.TypeConverter
	transportController transport.Controller
	federator           federation.Federator
	oauthServer         oauth.Server
	timelineManager     timeline.Manager
	emailSender         email.Sender

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testTags         map[string]*gtsmodel.Tag
	testMentions     map[string]*gtsmodel.Mention
	testAutheds      map[string]*oauth.Auth
	testBlocks       map[string]*gtsmodel.Block
	testActivities   map[string]testrig.ActivityWithSignature

	sentHTTPRequests map[string][]byte

	processor processing.Processor
}

func (suite *ProcessingStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
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
}

func (suite *ProcessingStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB()
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	suite.storage = testrig.NewTestStorage()
	suite.typeconverter = testrig.NewTestTypeConverter(suite.db)

	// make an http client that stores POST requests it receives into a map,
	// and also responds to correctly to dereference requests
	suite.sentHTTPRequests = make(map[string][]byte)
	httpClient := testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPost && req.Body != nil {
			requestBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}
			if err := req.Body.Close(); err != nil {
				panic(err)
			}
			suite.sentHTTPRequests[req.URL.String()] = requestBytes
		}

		if req.URL.String() == suite.testAccounts["remote_account_1"].URI {
			// the request is for remote account 1
			satan := suite.testAccounts["remote_account_1"]

			satanAS, err := suite.typeconverter.AccountToAS(context.Background(), satan)
			if err != nil {
				panic(err)
			}

			satanI, err := streams.Serialize(satanAS)
			if err != nil {
				panic(err)
			}
			satanJson, err := json.Marshal(satanI)
			if err != nil {
				panic(err)
			}
			responseType := "application/activity+json"

			reader := bytes.NewReader(satanJson)
			readCloser := io.NopCloser(reader)
			response := &http.Response{
				StatusCode:    200,
				Body:          readCloser,
				ContentLength: int64(len(satanJson)),
				Header: http.Header{
					"content-type": {responseType},
				},
			}
			return response, nil
		}

		if req.URL.String() == suite.testAccounts["remote_account_2"].URI {
			// the request is for remote account 2
			someAccount := suite.testAccounts["remote_account_2"]

			someAccountAS, err := suite.typeconverter.AccountToAS(context.Background(), someAccount)
			if err != nil {
				panic(err)
			}

			someAccountI, err := streams.Serialize(someAccountAS)
			if err != nil {
				panic(err)
			}
			someAccountJson, err := json.Marshal(someAccountI)
			if err != nil {
				panic(err)
			}
			responseType := "application/activity+json"

			reader := bytes.NewReader(someAccountJson)
			readCloser := io.NopCloser(reader)
			response := &http.Response{
				StatusCode:    200,
				Body:          readCloser,
				ContentLength: int64(len(someAccountJson)),
				Header: http.Header{
					"content-type": {responseType},
				},
			}
			return response, nil
		}

		if req.URL.String() == "http://example.org/users/some_user/statuses/afaba698-5740-4e32-a702-af61aa543bc1" {
			// the request is for the forwarded message
			message := suite.testActivities["forwarded_message"].Activity.GetActivityStreamsObject().At(0).GetActivityStreamsNote()
			messageI, err := streams.Serialize(message)
			if err != nil {
				panic(err)
			}
			messageJson, err := json.Marshal(messageI)
			if err != nil {
				panic(err)
			}
			responseType := "application/activity+json"

			reader := bytes.NewReader(messageJson)
			readCloser := io.NopCloser(reader)
			response := &http.Response{
				StatusCode:    200,
				Body:          readCloser,
				ContentLength: int64(len(messageJson)),
				Header: http.Header{
					"content-type": {responseType},
				},
			}
			return response, nil
		}

		r := ioutil.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})

	clientWorker := worker.New[messages.FromClientAPI](-1, -1)
	fedWorker := worker.New[messages.FromFederator](-1, -1)

	suite.transportController = testrig.NewTestTransportController(httpClient, suite.db, fedWorker)
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.federator = testrig.NewTestFederator(suite.db, suite.transportController, suite.storage, suite.mediaManager, fedWorker)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.emailSender = testrig.NewEmailSender("../../web/template/", nil)

	suite.processor = processing.NewProcessor(suite.typeconverter, suite.federator, suite.oauthServer, suite.mediaManager, suite.storage, suite.db, suite.emailSender, clientWorker, fedWorker)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
	if err := suite.processor.Start(); err != nil {
		panic(err)
	}
}

func (suite *ProcessingStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	if err := suite.processor.Stop(); err != nil {
		panic(err)
	}
}
