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

package fileserver_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/gruf/go-store/kv"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ServeFileTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	storage      *kv.KVStore
	federator    federation.Federator
	tc           typeutils.TypeConverter
	processor    processing.Processor
	mediaManager media.Manager
	oauthServer  oauth.Server
	emailSender  email.Sender

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment

	// item being tested
	fileServer *fileserver.FileServer
}

/*
	TEST INFRASTRUCTURE
*/

func (suite *ServeFileTestSuite) SetupSuite() {
	// setup standard items
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewTestStorage()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage, testrig.NewTestMediaManager(suite.db, suite.storage))
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)

	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender, testrig.NewTestMediaManager(suite.db, suite.storage))
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)

	// setup module being tested
	suite.fileServer = fileserver.New(suite.processor).(*fileserver.FileServer)
}

func (suite *ServeFileTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

func (suite *ServeFileTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

func (suite *ServeFileTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

/*
	ACTUAL TESTS
*/

func (suite *ServeFileTestSuite) TestServeOriginalFileSuccessful() {
	targetAttachment, ok := suite.testAttachments["admin_account_status_1_attachment_1"]
	suite.True(ok)
	suite.NotNil(targetAttachment)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAttachment.URL, nil)
	ctx.Request.Header.Set("accept", "*/*")

	// normally the router would populate these params from the path values,
	// but because we're calling the ServeFile function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   fileserver.AccountIDKey,
			Value: targetAttachment.AccountID,
		},
		gin.Param{
			Key:   fileserver.MediaTypeKey,
			Value: string(media.TypeAttachment),
		},
		gin.Param{
			Key:   fileserver.MediaSizeKey,
			Value: string(media.SizeOriginal),
		},
		gin.Param{
			Key:   fileserver.FileNameKey,
			Value: fmt.Sprintf("%s.jpeg", targetAttachment.ID),
		},
	}

	// call the function we're testing and check status code
	suite.fileServer.ServeFile(ctx)
	suite.EqualValues(http.StatusOK, recorder.Code)
	suite.EqualValues("image/jpeg", recorder.Header().Get("content-type"))

	b, err := ioutil.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)

	fileInStorage, err := suite.storage.Get(targetAttachment.File.Path)
	suite.NoError(err)
	suite.NotNil(fileInStorage)
	suite.Equal(b, fileInStorage)
}

func (suite *ServeFileTestSuite) TestServeSmallFileSuccessful() {
	targetAttachment, ok := suite.testAttachments["admin_account_status_1_attachment_1"]
	suite.True(ok)
	suite.NotNil(targetAttachment)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAttachment.Thumbnail.URL, nil)
	ctx.Request.Header.Set("accept", "*/*")

	// normally the router would populate these params from the path values,
	// but because we're calling the ServeFile function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   fileserver.AccountIDKey,
			Value: targetAttachment.AccountID,
		},
		gin.Param{
			Key:   fileserver.MediaTypeKey,
			Value: string(media.TypeAttachment),
		},
		gin.Param{
			Key:   fileserver.MediaSizeKey,
			Value: string(media.SizeSmall),
		},
		gin.Param{
			Key:   fileserver.FileNameKey,
			Value: fmt.Sprintf("%s.jpeg", targetAttachment.ID),
		},
	}

	// call the function we're testing and check status code
	suite.fileServer.ServeFile(ctx)
	suite.EqualValues(http.StatusOK, recorder.Code)
	suite.EqualValues("image/jpeg", recorder.Header().Get("content-type"))

	b, err := ioutil.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)

	fileInStorage, err := suite.storage.Get(targetAttachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotNil(fileInStorage)
	suite.Equal(b, fileInStorage)
}

func TestServeFileTestSuite(t *testing.T) {
	suite.Run(t, new(ServeFileTestSuite))
}
