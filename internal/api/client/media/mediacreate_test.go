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
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/gruf/go-store/kv"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	mediamodule "github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaCreateTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	storage      *kv.KVStore
	mediaManager media.Manager
	federator    federation.Federator
	tc           typeutils.TypeConverter
	oauthServer  oauth.Server
	emailSender  email.Sender
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
	testrig.InitTestConfig()
	testrig.InitTestLog()

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)

	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewTestStorage()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db, fedWorker), suite.storage, suite.mediaManager, fedWorker)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender, suite.mediaManager, clientWorker, fedWorker)

	// setup module being tested
	suite.mediaModule = mediamodule.New(suite.processor).(*mediamodule.Module)
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

func (suite *MediaCreateTestSuite) TestMediaCreateSuccessful() {
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
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", mediamodule.BasePathV1), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

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
	suite.NoError(err)
	fmt.Println(string(b))

	attachmentReply := &model.Attachment{}
	err = json.Unmarshal(b, attachmentReply)
	suite.NoError(err)

	suite.Equal("this is a test image -- a cool background from somewhere", attachmentReply.Description)
	suite.Equal("image", attachmentReply.Type)
	suite.EqualValues(model.MediaMeta{
		Original: model.MediaDimensions{
			Width:  1920,
			Height: 1080,
			Size:   "1920x1080",
			Aspect: 1.7777778,
		},
		Small: model.MediaDimensions{
			Width:  512,
			Height: 288,
			Size:   "512x288",
			Aspect: 1.7777778,
		},
		Focus: model.MediaFocus{
			X: -0.5,
			Y: 0.5,
		},
	}, attachmentReply.Meta)
	suite.Equal("LjBzUo#6RQR._NvzRjWF?urqV@a$", attachmentReply.Blurhash)
	suite.NotEmpty(attachmentReply.ID)
	suite.NotEmpty(attachmentReply.URL)
	suite.NotEmpty(attachmentReply.PreviewURL)
	suite.Equal(len(storageKeysBeforeRequest)+2, len(storageKeysAfterRequest)) // 2 images should be added to storage: the original and the thumbnail
}

func (suite *MediaCreateTestSuite) TestMediaCreateLongDescription() {
	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// read a random string of a really long description
	descriptionBytes := make([]byte, 5000)
	if _, err := rand.Read(descriptionBytes); err != nil {
		panic(err)
	}
	description := base64.RawStdEncoding.EncodeToString(descriptionBytes)

	// create the request
	buf, w, err := testrig.CreateMultipartFormData("file", "../../../../testrig/media/test-jpeg.jpg", map[string]string{
		"description": description,
		"focus":       "-0.5,0.5",
	})
	if err != nil {
		panic(err)
	}
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", mediamodule.BasePathV1), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

	// do the actual request
	suite.mediaModule.MediaCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: image description length must be between 0 and 500 characters (inclusive), but provided image description was 6667 chars"}`, string(b))
}

func (suite *MediaCreateTestSuite) TestMediaCreateTooShortDescription() {
	// set the min description length
	config.SetMediaDescriptionMinChars(500)

	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// create the request
	buf, w, err := testrig.CreateMultipartFormData("file", "../../../../testrig/media/test-jpeg.jpg", map[string]string{
		"description": "", // provide an empty description
		"focus":       "-0.5,0.5",
	})
	if err != nil {
		panic(err)
	}
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", mediamodule.BasePathV1), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

	// do the actual request
	suite.mediaModule.MediaCreatePOSTHandler(ctx)

	// check response -- there should be no error because minimum description length is checked on *UPDATE*, not initial upload
	suite.EqualValues(http.StatusOK, recorder.Code)
}

func TestMediaCreateTestSuite(t *testing.T) {
	suite.Run(t, new(MediaCreateTestSuite))
}
