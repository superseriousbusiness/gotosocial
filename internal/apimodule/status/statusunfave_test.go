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
	"strings"
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

type StatusUnfaveTestSuite struct {
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
	testStatuses     map[string]*gtsmodel.Status

	// module being tested
	statusModule *status.Module
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *StatusUnfaveTestSuite) SetupSuite() {
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

func (suite *StatusUnfaveTestSuite) TearDownSuite() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *StatusUnfaveTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *StatusUnfaveTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

/*
	ACTUAL TESTS
*/

// unfave a status
func (suite *StatusUnfaveTestSuite) TestPostUnfave() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// this is the status we wanna unfave: in the testrig it's already faved by this account
	targetStatus := suite.testStatuses["admin_account_status_1"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.UnfavouritePath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusUnfavePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &mastomodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.SpoilerText)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), mastomodel.VisibilityPublic, statusReply.Visibility)
	assert.False(suite.T(), statusReply.Favourited)
	assert.Equal(suite.T(), 0, statusReply.FavouritesCount)
}

// try to unfave a status that's already not faved
func (suite *StatusUnfaveTestSuite) TestPostAlreadyNotFaved() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.TokenToOauthToken(t)

	// this is the status we wanna unfave: in the testrig it's not faved by this account
	targetStatus := suite.testStatuses["admin_account_status_2"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.UnfavouritePath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusUnfavePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &mastomodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.SpoilerText)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.True(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), mastomodel.VisibilityPublic, statusReply.Visibility)
	assert.False(suite.T(), statusReply.Favourited)
	assert.Equal(suite.T(), 0, statusReply.FavouritesCount)
}

func TestStatusUnfaveTestSuite(t *testing.T) {
	suite.Run(t, new(StatusUnfaveTestSuite))
}
