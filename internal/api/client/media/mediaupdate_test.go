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

package media_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mediamodule "code.superseriousbusiness.org/gotosocial/internal/api/client/media"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
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
	"github.com/stretchr/testify/suite"
)

type MediaUpdateTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	storage      *storage.Driver
	federator    *federation.Federator
	tc           *typeutils.Converter
	mediaManager *media.Manager
	oauthServer  oauth.Server
	emailSender  email.Sender
	processor    *processing.Processor
	state        state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
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

func (suite *MediaUpdateTestSuite) SetupTest() {
	testrig.StartNoopWorkers(&suite.state)

	// setup standard items
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.Caches.Init()

	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	suite.db = testrig.NewTestDB(&suite.state)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	suite.tc = typeutils.NewConverter(&suite.state)

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.oauthServer = testrig.NewTestOauthServer(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../../testrig/media")), suite.mediaManager)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)

	// setup module being tested
	suite.mediaModule = mediamodule.New(suite.processor)

	// setup test data
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

func (suite *MediaUpdateTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

/*
	ACTUAL TESTS
*/

func (suite *MediaUpdateTestSuite) TestUpdateImage() {
	toUpdate := suite.testAttachments["local_account_1_unattached_1"]

	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// create the request
	buf, w, err := testrig.CreateMultipartFormData(nil, map[string][]string{
		"id":          {toUpdate.ID},
		"description": {"new description!"},
		"focus":       {"-0.1,0.3"},
	})
	if err != nil {
		panic(err)
	}
	ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:8080/api/v1/media/%s", toUpdate.ID), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")
	ctx.AddParam(apiutil.APIVersionKey, apiutil.APIv1)
	ctx.AddParam(mediamodule.IDKey, toUpdate.ID)

	// do the actual request
	suite.mediaModule.MediaPUTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// reply should be an attachment
	attachmentReply := &apimodel.Attachment{}
	err = json.Unmarshal(b, attachmentReply)
	suite.NoError(err)

	// the reply should contain the updated fields
	suite.Equal("new description!", *attachmentReply.Description)
	suite.EqualValues("image", attachmentReply.Type)
	suite.EqualValues(apimodel.MediaMeta{
		Original: apimodel.MediaDimensions{Width: 800, Height: 450, FrameRate: "", Duration: 0, Bitrate: 0, Size: "800x450", Aspect: 1.7777778},
		Small:    apimodel.MediaDimensions{Width: 512, Height: 288, FrameRate: "", Duration: 0, Bitrate: 0, Size: "512x288", Aspect: 1.7777778},
		Focus:    &apimodel.MediaFocus{X: -0.1, Y: 0.3},
	}, *attachmentReply.Meta)
	suite.Equal(toUpdate.Blurhash, *attachmentReply.Blurhash)
	suite.Equal(toUpdate.ID, attachmentReply.ID)
	suite.Equal(toUpdate.URL, *attachmentReply.URL)
	suite.NotEmpty(toUpdate.Thumbnail.URL, attachmentReply.PreviewURL)
}

func (suite *MediaUpdateTestSuite) TestUpdateImageShortDescription() {
	// set the min description length
	config.SetMediaDescriptionMinChars(50)

	toUpdate := suite.testAttachments["local_account_1_unattached_1"]

	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// create the request
	buf, w, err := testrig.CreateMultipartFormData(nil, map[string][]string{
		"id":          {toUpdate.ID},
		"description": {"new description!"},
		"focus":       {"-0.1,0.3"},
	})
	if err != nil {
		panic(err)
	}
	ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:8080/api/v1/media/%s", toUpdate.ID), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")
	ctx.AddParam(apiutil.APIVersionKey, apiutil.APIv1)
	ctx.AddParam(mediamodule.IDKey, toUpdate.ID)

	// do the actual request
	suite.mediaModule.MediaPUTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// reply should be an error message
	suite.Equal(`{"error":"Bad Request: image description length must be between 50 and 500 characters (inclusive), but provided image description was 16 chars"}`, string(b))
}

func TestMediaUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(MediaUpdateTestSuite))
}
