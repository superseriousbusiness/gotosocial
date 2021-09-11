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

package media_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	mediamodule "github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
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

type MediaCreateTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config       *config.Config
	db           db.DB
	log          *logrus.Logger
	storage      *kv.KVStore
	federator    federation.Federator
	tc           typeutils.TypeConverter
	mediaHandler media.Handler
	oauthServer  oauth.Server
	processor    processing.Processor

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment

	// item being tested
	mediaModule *mediamodule.Module
}

/*
	TEST INFRASTRUCTURE
*/

func (suite *MediaCreateTestSuite) SetupSuite() {
	// setup standard items
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.mediaHandler = testrig.NewTestMediaHandler(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)

	// setup module being tested
	suite.mediaModule = mediamodule.New(suite.config, suite.processor, suite.log).(*mediamodule.Module)
}

func (suite *MediaCreateTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

func (suite *MediaCreateTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

func (suite *MediaCreateTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

/*
	ACTUAL TESTS
*/

func (suite *MediaCreateTestSuite) TestStatusCreatePOSTImageHandlerSuccessful() {
	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// see what's in storage *before* the request
	storageKeysBeforeRequest := []string{}
	iter, err := suite.storage.Iterator(nil)
	if err != nil {
		panic(err)
	}
	for iter.Next() {
		storageKeysBeforeRequest = append(storageKeysBeforeRequest, iter.Key())
	}
	iter.Release()

	// create the request
	buf, w, err := testrig.CreateMultipartFormData("file", "../../../../testrig/media/test-jpeg.jpg", map[string]string{
		"description": "this is a test image -- a cool background from somewhere",
		"focus":       "-0.5,0.5",
	})
	if err != nil {
		panic(err)
	}
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", mediamodule.BasePath), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())

	// do the actual request
	suite.mediaModule.MediaCreatePOSTHandler(ctx)

	// check what's in storage *after* the request
	storageKeysAfterRequest := []string{}
	iter, err = suite.storage.Iterator(nil)
	if err != nil {
		panic(err)
	}
	for iter.Next() {
		storageKeysAfterRequest = append(storageKeysAfterRequest, iter.Key())
	}
	iter.Release()

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	fmt.Println(string(b))

	attachmentReply := &model.Attachment{}
	err = json.Unmarshal(b, attachmentReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "this is a test image -- a cool background from somewhere", attachmentReply.Description)
	assert.Equal(suite.T(), "image", attachmentReply.Type)
	assert.EqualValues(suite.T(), model.MediaMeta{
		Original: model.MediaDimensions{
			Width:  1920,
			Height: 1080,
			Size:   "1920x1080",
			Aspect: 1.7777778,
		},
		Small: model.MediaDimensions{
			Width:  256,
			Height: 144,
			Size:   "256x144",
			Aspect: 1.7777778,
		},
		Focus: model.MediaFocus{
			X: -0.5,
			Y: 0.5,
		},
	}, attachmentReply.Meta)
	assert.Equal(suite.T(), "LjCZnlvyRkRn_NvzRjWF?urqV@f9", attachmentReply.Blurhash)
	assert.NotEmpty(suite.T(), attachmentReply.ID)
	assert.NotEmpty(suite.T(), attachmentReply.URL)
	assert.NotEmpty(suite.T(), attachmentReply.PreviewURL)
	assert.Equal(suite.T(), len(storageKeysBeforeRequest)+2, len(storageKeysAfterRequest)) // 2 images should be added to storage: the original and the thumbnail
}

func TestMediaCreateTestSuite(t *testing.T) {
	suite.Run(t, new(MediaCreateTestSuite))
}
