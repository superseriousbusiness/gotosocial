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

package status_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"codeberg.org/gruf/go-store/kv"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/worker"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	tc           typeutils.TypeConverter
	mediaManager media.Manager
	federator    federation.Federator
	emailSender  email.Sender
	processor    processing.Processor
	storage      *kv.KVStore

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
	statusModule *status.Module
}

func (suite *StatusStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFollows = testrig.NewTestFollows()
}

func (suite *StatusStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.storage = testrig.NewTestStorage()
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	fedWorker := worker.New[messages.FromFederator](-1, -1)
	clientWorker := worker.New[messages.FromClientAPI](-1, -1)

	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(suite.testHttpClient(), suite.db, fedWorker), suite.storage, suite.mediaManager, fedWorker)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender, suite.mediaManager, clientWorker, fedWorker)
	suite.statusModule = status.New(suite.processor).(*status.Module)
}

func (suite *StatusStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *StatusStandardTestSuite) testHttpClient() pub.HttpClient {
	remoteAccount := suite.testAccounts["remote_account_1"]
	remoteAccountNamestring := fmt.Sprintf("acct:%s@%s", remoteAccount.Username, remoteAccount.Domain)
	remoteAccountWebfingerURI := fmt.Sprintf("https://%s/.well-known/webfinger?resource=%s", remoteAccount.Domain, remoteAccountNamestring)

	fmt.Println(remoteAccountWebfingerURI)

	httpClient := testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		// respond correctly to a webfinger lookup
		if req.URL.String() == remoteAccountWebfingerURI {
			responseJson := fmt.Sprintf(`
			{
				"subject": "%s",
				"aliases": [
				  "%s",
				  "%s"
				],
				"links": [
				  {
					"rel": "http://webfinger.net/rel/profile-page",
					"type": "text/html",
					"href": "%s"
				  },
				  {
					"rel": "self",
					"type": "application/activity+json",
					"href": "%s"
				  }
				]
			}`, remoteAccountNamestring, remoteAccount.URI, remoteAccount.URL, remoteAccount.URL, remoteAccount.URI)
			responseType := "application/json"

			reader := bytes.NewReader([]byte(responseJson))
			readCloser := io.NopCloser(reader)
			response := &http.Response{
				StatusCode:    200,
				Body:          readCloser,
				ContentLength: int64(len(responseJson)),
				Header: http.Header{
					"content-type": {responseType},
				},
			}
			return response, nil
		}

		// respond correctly to an account dereference
		if req.URL.String() == remoteAccount.URI {
			satanAS, err := suite.tc.AccountToAS(context.Background(), remoteAccount)
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

		r := ioutil.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})

	return httpClient
}
