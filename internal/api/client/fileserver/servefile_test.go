/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

	"git.iim.gay/grufwub/go-store/kv"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
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
	config       *config.Config
	db           db.DB
	log          *logrus.Logger
	storage      *kv.KVStore
	federator    federation.Federator
	tc           typeutils.TypeConverter
	processor    processing.Processor
	mediaHandler media.Handler
	oauthServer  oauth.Server

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
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.mediaHandler = testrig.NewTestMediaHandler(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)

	// setup module being tested
	suite.fileServer = fileserver.New(suite.config, suite.processor, suite.log).(*fileserver.FileServer)
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
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), targetAttachment)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAttachment.URL, nil)

	// normally the router would populate these params from the path values,
	// but because we're calling the ServeFile function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   fileserver.AccountIDKey,
			Value: targetAttachment.AccountID,
		},
		gin.Param{
			Key:   fileserver.MediaTypeKey,
			Value: string(media.Attachment),
		},
		gin.Param{
			Key:   fileserver.MediaSizeKey,
			Value: string(media.Original),
		},
		gin.Param{
			Key:   fileserver.FileNameKey,
			Value: fmt.Sprintf("%s.jpeg", targetAttachment.ID),
		},
	}

	// call the function we're testing and check status code
	suite.fileServer.ServeFile(ctx)
	suite.EqualValues(http.StatusOK, recorder.Code)

	b, err := ioutil.ReadAll(recorder.Body)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), b)

	fileInStorage, err := suite.storage.Get(targetAttachment.File.Path)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), fileInStorage)
	assert.Equal(suite.T(), b, fileInStorage)
}

func TestServeFileTestSuite(t *testing.T) {
	suite.Run(t, new(ServeFileTestSuite))
}
