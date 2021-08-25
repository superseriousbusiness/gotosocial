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

package account_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountUpdateTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountUpdateTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *AccountUpdateTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewTestStorage()
	suite.log = testrig.NewTestLog()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)
	suite.accountModule = account.New(suite.config, suite.processor, suite.log).(*account.Module)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *AccountUpdateTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandler() {

	requestBody, w, err := testrig.CreateMultipartFormData("header", "../../../../testrig/media/test-jpeg.jpg", map[string]string{
		"display_name": "updated zork display name!!!",
		"locked":       "true",
	})
	if err != nil {
		panic(err)
	}

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:8080/%s", account.UpdateCredentialsPath), bytes.NewReader(requestBody.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	suite.accountModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// check response

	// 1. we should have OK because our request was valid
	suite.EqualValues(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	fmt.Println(string(b))

	// TODO write more assertions allee
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
