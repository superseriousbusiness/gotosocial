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

package status_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/status"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	mastomodel "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusCreateTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config         *config.Config
	db             db.DB
	log            *logrus.Logger
	storage        storage.Storage
	mastoConverter mastotypes.Converter
	mediaHandler   media.Handler
	oauthServer    oauth.Server
	distributor    distributor.Distributor

	// standard suite models
	testTokens       map[string]*oauth.Token
	testClients      map[string]*oauth.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment

	// module being tested
	statusModule *status.Module
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *StatusCreateTestSuite) SetupSuite() {
	// setup standard items
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.mastoConverter = testrig.NewTestMastoConverter(suite.db)
	suite.mediaHandler = testrig.NewTestMediaHandler(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	suite.distributor = testrig.NewTestDistributor()

	// setup module being tested
	suite.statusModule = status.New(suite.config, suite.db, suite.mediaHandler, suite.mastoConverter, suite.distributor, suite.log).(*status.Module)
}

func (suite *StatusCreateTestSuite) TearDownSuite() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *StatusCreateTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *StatusCreateTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

/*
	ACTUAL TESTS
*/

/*
	TESTING: StatusCreatePOSTHandler
*/

// Post a new status with some custom visibility settings
func (suite *StatusCreateTestSuite) TestPostNewStatus() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":              {"this is a brand new status! #helloworld"},
		"spoiler_text":        {"hello hello"},
		"sensitive":           {"true"},
		"visibility_advanced": {"mutuals_only"},
		"likeable":            {"false"},
		"replyable":           {"false"},
		"federated":           {"false"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &mastomodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "hello hello", statusReply.SpoilerText)
	assert.Equal(suite.T(), "this is a brand new status! #helloworld", statusReply.Content)
	assert.True(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), mastomodel.VisibilityPrivate, statusReply.Visibility)
	assert.Len(suite.T(), statusReply.Tags, 1)
	assert.Equal(suite.T(), mastomodel.Tag{
		Name: "helloworld",
		URL:  "http://localhost:8080/tags/helloworld",
	}, statusReply.Tags[0])

	gtsTag := &gtsmodel.Tag{}
	err = suite.db.GetWhere("name", "helloworld", gtsTag)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), statusReply.Account.ID, gtsTag.FirstSeenFromAccountID)
}

func (suite *StatusCreateTestSuite) TestPostNewStatusWithEmoji() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status": {"here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: \n here's an emoji that isn't in the db: :test_emoji: "},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &mastomodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "", statusReply.SpoilerText)
	assert.Equal(suite.T(), "here is a rainbow emoji a few times! :rainbow: :rainbow: :rainbow: \n here's an emoji that isn't in the db: :test_emoji: ", statusReply.Content)

	assert.Len(suite.T(), statusReply.Emojis, 1)
	mastoEmoji := statusReply.Emojis[0]
	gtsEmoji := testrig.NewTestEmojis()["rainbow"]

	assert.Equal(suite.T(), gtsEmoji.Shortcode, mastoEmoji.Shortcode)
	assert.Equal(suite.T(), gtsEmoji.ImageURL, mastoEmoji.URL)
	assert.Equal(suite.T(), gtsEmoji.ImageStaticURL, mastoEmoji.StaticURL)
}

// Try to reply to a status that doesn't exist
func (suite *StatusCreateTestSuite) TestReplyToNonexistentStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":         {"this is a reply to a status that doesn't exist"},
		"spoiler_text":   {"don't open cuz it won't work"},
		"in_reply_to_id": {"3759e7ef-8ee1-4c0c-86f6-8b70b9ad3d50"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response

	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"status with id 3759e7ef-8ee1-4c0c-86f6-8b70b9ad3d50 not replyable because it doesn't exist"}`, string(b))
}

// Post a reply to the status of a local user that allows replies.
func (suite *StatusCreateTestSuite) TestReplyToLocalStatus() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":         {fmt.Sprintf("hello @%s this reply should work!", testrig.NewTestAccounts()["local_account_2"].Username)},
		"in_reply_to_id": {testrig.NewTestStatuses()["local_account_2_status_1"].ID},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &mastomodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "", statusReply.SpoilerText)
	assert.Equal(suite.T(), fmt.Sprintf("hello @%s this reply should work!", testrig.NewTestAccounts()["local_account_2"].Username), statusReply.Content)
	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), mastomodel.VisibilityPublic, statusReply.Visibility)
	assert.Equal(suite.T(), testrig.NewTestStatuses()["local_account_2_status_1"].ID, statusReply.InReplyToID)
	assert.Equal(suite.T(), testrig.NewTestAccounts()["local_account_2"].ID, statusReply.InReplyToAccountID)
	assert.Len(suite.T(), statusReply.Mentions, 1)
}

// Take a media file which is currently not associated with a status, and attach it to a new status.
func (suite *StatusCreateTestSuite) TestAttachNewMediaSuccess() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", status.BasePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":    {"here's an image attachment"},
		"media_ids": {"7a3b9f77-ab30-461e-bdd8-e64bd1db3008"},
	}
	suite.statusModule.StatusCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	fmt.Println(string(b))

	statusReply := &mastomodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "", statusReply.SpoilerText)
	assert.Equal(suite.T(), "here's an image attachment", statusReply.Content)
	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), mastomodel.VisibilityPublic, statusReply.Visibility)

	// there should be one media attachment
	assert.Len(suite.T(), statusReply.MediaAttachments, 1)

	// get the updated media attachment from the database
	gtsAttachment := &gtsmodel.MediaAttachment{}
	err = suite.db.GetByID(statusReply.MediaAttachments[0].ID, gtsAttachment)
	assert.NoError(suite.T(), err)

	// convert it to a masto attachment
	gtsAttachmentAsMasto, err := suite.mastoConverter.AttachmentToMasto(gtsAttachment)
	assert.NoError(suite.T(), err)

	// compare it with what we have now
	assert.EqualValues(suite.T(), statusReply.MediaAttachments[0], gtsAttachmentAsMasto)

	// the status id of the attachment should now be set to the id of the status we just created
	assert.Equal(suite.T(), statusReply.ID, gtsAttachment.StatusID)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}
