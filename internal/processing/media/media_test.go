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

package media_test

import (
	"bytes"
	"io"
	"net/http"

	"codeberg.org/gruf/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	mediaprocessing "github.com/superseriousbusiness/gotosocial/internal/processing/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/worker"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db                  db.DB
	tc                  typeutils.TypeConverter
	storage             *kv.KVStore
	mediaManager        media.Manager
	transportController transport.Controller

	// standard suite models
	testTokens            map[string]*gtsmodel.Token
	testClients           map[string]*gtsmodel.Client
	testApplications      map[string]*gtsmodel.Application
	testUsers             map[string]*gtsmodel.User
	testAccounts          map[string]*gtsmodel.Account
	testAttachments       map[string]*gtsmodel.MediaAttachment
	testStatuses          map[string]*gtsmodel.Status
	testRemoteAttachments map[string]testrig.RemoteAttachmentFile

	// module being tested
	mediaProcessor mediaprocessing.Processor
}

func (suite *MediaStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testRemoteAttachments = testrig.NewTestFediAttachments("../../../testrig/media")
}

func (suite *MediaStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.storage = testrig.NewTestStorage()
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.transportController = suite.mockTransportController()
	suite.mediaProcessor = mediaprocessing.New(suite.db, suite.tc, suite.mediaManager, suite.transportController, suite.storage)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
}

func (suite *MediaStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *MediaStandardTestSuite) mockTransportController() transport.Controller {
	do := func(req *http.Request) (*http.Response, error) {
		logrus.Debugf("received request for %s", req.URL)

		responseBytes := []byte{}
		responseType := ""
		responseLength := 0

		if attachment, ok := suite.testRemoteAttachments[req.URL.String()]; ok {
			responseBytes = attachment.Data
			responseType = attachment.ContentType
		}

		if len(responseBytes) != 0 {
			// we found something, so print what we're going to return
			logrus.Debugf("returning response %s", string(responseBytes))
		}
		responseLength = len(responseBytes)

		reader := bytes.NewReader(responseBytes)
		readCloser := io.NopCloser(reader)
		response := &http.Response{
			StatusCode:    200,
			Body:          readCloser,
			ContentLength: int64(responseLength),
			Header: http.Header{
				"content-type": {responseType},
			},
		}

		return response, nil
	}
	fedWorker := worker.New[messages.FromFederator](-1, -1)
	mockClient := testrig.NewMockHTTPClient(do)
	return testrig.NewTestTransportController(mockClient, suite.db, fedWorker)
}
