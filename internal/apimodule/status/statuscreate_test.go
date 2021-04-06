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
	"context"
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
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusCreateTestSuite struct {
	suite.Suite
	config           *config.Config
	mockOauthServer  *oauth.MockServer
	mockStorage      *storage.MockStorage
	mediaHandler     media.MediaHandler
	distributor      *distributor.MockDistributor
	testTokens       map[string]*oauth.Token
	testClients      map[string]*oauth.Client
	testApplications map[string]*model.Application
	testUsers        map[string]*model.User
	testAccounts     map[string]*model.Account
	log              *logrus.Logger
	db               db.DB
	statusModule     *statusModule
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *StatusCreateTestSuite) SetupSuite() {
	// some of our subsequent entities need a log so create this here
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log

	// Direct config to local postgres instance
	c := config.Empty()
	c.Protocol = "http"
	c.Host = "localhost"
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	c.MediaConfig = &config.MediaConfig{
		MaxImageSize: 2 << 20,
	}
	c.StorageConfig = &config.StorageConfig{
		Backend:       "local",
		BasePath:      "/tmp",
		ServeProtocol: "http",
		ServeHost:     "localhost",
		ServeBasePath: "/fileserver/media",
	}
	c.StatusesConfig = &config.StatusesConfig{
		MaxChars:           500,
		CWMaxChars:         50,
		PollMaxOptions:     4,
		PollOptionMaxChars: 50,
		MaxMediaFiles:      4,
	}
	suite.config = c

	// use an actual database for this, because it's just easier than mocking one out
	database, err := db.New(context.Background(), c, log)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.db = database

	suite.mockOauthServer = &oauth.MockServer{}
	suite.mockStorage = &storage.MockStorage{}
	suite.mediaHandler = media.New(suite.config, suite.db, suite.mockStorage, log)
	suite.distributor = &distributor.MockDistributor{}
	suite.distributor.On("FromClientAPI").Return(make(chan distributor.FromClientAPI, 100))

	suite.statusModule = New(suite.config, suite.db, suite.mockOauthServer, suite.mediaHandler, suite.distributor, suite.log).(*statusModule)
}

func (suite *StatusCreateTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

func (suite *StatusCreateTestSuite) SetupTest() {
	if err := testrig.StandardDBSetup(suite.db); err != nil {
		panic(err)
	}
	suite.testTokens = testrig.TestTokens()
	suite.testClients = testrig.TestClients()
	suite.testApplications = testrig.TestApplications()
	suite.testUsers = testrig.TestUsers()
	suite.testAccounts = testrig.TestAccounts()
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *StatusCreateTestSuite) TearDownTest() {
	if err := testrig.StandardDBTeardown(suite.db); err != nil {
		panic(err)
	}
}

/*
	ACTUAL TESTS
*/

/*
	TESTING: StatusCreatePOSTHandler
*/

func (suite *StatusCreateTestSuite) TestStatusCreatePOSTHandlerSuccessful() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.PGTokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = url.Values{
		"status":       {"this is a brand new status!"},
		"spoiler_text": {"hello hello"},
		"sensitive":    {"true"},
		"visibility":   {"public"},
		// 	Status string `form:"status"`
		// // Array of Attachment ids to be attached as media. If provided, status becomes optional, and poll cannot be used.
		// MediaIDs []string `form:"media_ids"`
		// // Poll to include with this status.
		// Poll *PollRequest `form:"poll"`
		// // ID of the status being replied to, if status is a reply
		// InReplyToID string `form:"in_reply_to_id"`
		// // Mark status and attached media as sensitive?
		// Sensitive bool `form:"sensitive"`
		// // Text to be shown as a warning or subject before the actual content. Statuses are generally collapsed behind this field.
		// SpoilerText string `form:"spoiler_text"`
		// // Visibility of the posted status. Enumerable oneOf public, unlisted, private, direct.
		// Visibility Visibility `form:"visibility"`
		// // ISO 8601 Datetime at which to schedule a status. Providing this paramter will cause ScheduledStatus to be returned instead of Status. Must be at least 5 minutes in the future.
		// ScheduledAt string `form:"scheduled_at"`
		// // ISO 639 language code for this status.
		// Language string `form:"language"`
	}
	suite.statusModule.statusCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &mastotypes.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "hello hello", statusReply.SpoilerText)
	assert.Equal(suite.T(), "this is a brand new status!", statusReply.Content)
	assert.True(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), mastotypes.VisibilityPublic, statusReply.Visibility)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}
