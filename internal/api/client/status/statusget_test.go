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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusGetTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusGetTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *StatusGetTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewTestStorage()
	suite.log = testrig.NewTestLog()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)
	suite.statusModule = status.New(suite.config, suite.processor, suite.log).(*status.Module)
	testrig.StandardDBSetup(suite.db)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *StatusGetTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

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

	// statusReply := &mastotypes.Status{}
	// err = json.Unmarshal(b, statusReply)
	// assert.NoError(suite.T(), err)

	// assert.Equal(suite.T(), "hello hello", statusReply.SpoilerText)
	// assert.Equal(suite.T(), "this is a brand new status! #helloworld", statusReply.Content)
	// assert.True(suite.T(), statusReply.Sensitive)
	// assert.Equal(suite.T(), mastotypes.VisibilityPrivate, statusReply.Visibility)
	// assert.Len(suite.T(), statusReply.Tags, 1)
	// assert.Equal(suite.T(), mastotypes.Tag{
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
