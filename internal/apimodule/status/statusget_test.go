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

package status

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusGetTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config         *config.Config
	db             db.DB
	log            *logrus.Logger
	storage        storage.Storage
	mastoConverter mastotypes.Converter
	mediaHandler   media.MediaHandler
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
	statusModule *statusModule
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *StatusGetTestSuite) SetupSuite() {
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
	suite.statusModule = New(suite.config, suite.db, suite.oauthServer, suite.mediaHandler, suite.mastoConverter, suite.distributor, suite.log).(*statusModule)
}

func (suite *StatusGetTestSuite) TearDownSuite() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *StatusGetTestSuite) SetupTest() {
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
func (suite *StatusGetTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

/*
	ACTUAL TESTS
*/

/*
	TESTING: StatusGetPOSTHandler
*/

// Post a new status with some custom visibility settings
func (suite *StatusGetTestSuite) TestPostNewStatus() {

	// t := suite.testTokens["local_account_1"]
	// oauthToken := oauth.PGTokenToOauthToken(t)

	// // setup
	// recorder := httptest.NewRecorder()
	// ctx, _ := gin.CreateTestContext(recorder)
	// ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	// ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	// ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	// ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	// ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	// ctx.Request.Form = url.Values{
	// 	"status":              {"this is a brand new status! #helloworld"},
	// 	"spoiler_text":        {"hello hello"},
	// 	"sensitive":           {"true"},
	// 	"visibility_advanced": {"mutuals_only"},
	// 	"likeable":            {"false"},
	// 	"replyable":           {"false"},
	// 	"federated":           {"false"},
	// }
	// suite.statusModule.statusGETHandler(ctx)

	// // check response

	// // 1. we should have OK from our call to the function
	// suite.EqualValues(http.StatusOK, recorder.Code)

	// result := recorder.Result()
	// defer result.Body.Close()
	// b, err := ioutil.ReadAll(result.Body)
	// assert.NoError(suite.T(), err)

	// statusReply := &mastomodel.Status{}
	// err = json.Unmarshal(b, statusReply)
	// assert.NoError(suite.T(), err)

	// assert.Equal(suite.T(), "hello hello", statusReply.SpoilerText)
	// assert.Equal(suite.T(), "this is a brand new status! #helloworld", statusReply.Content)
	// assert.True(suite.T(), statusReply.Sensitive)
	// assert.Equal(suite.T(), mastomodel.VisibilityPrivate, statusReply.Visibility)
	// assert.Len(suite.T(), statusReply.Tags, 1)
	// assert.Equal(suite.T(), mastomodel.Tag{
	// 	Name: "helloworld",
	// 	URL:  "http://localhost:8080/tags/helloworld",
	// }, statusReply.Tags[0])

	// gtsTag := &gtsmodel.Tag{}
	// err = suite.db.GetWhere("name", "helloworld", gtsTag)
	// assert.NoError(suite.T(), err)
	// assert.Equal(suite.T(), statusReply.Account.ID, gtsTag.FirstSeenFromAccountID)
}

func TestStatusGetTestSuite(t *testing.T) {
	suite.Run(t, new(StatusGetTestSuite))
}
