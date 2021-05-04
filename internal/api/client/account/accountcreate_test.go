// /*
//    GoToSocial
//    Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.

//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.
// */

package account_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"mime/multipart"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/suite"
// 	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
// 	"github.com/superseriousbusiness/gotosocial/internal/api/model"
// 	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
// 	"github.com/superseriousbusiness/gotosocial/testrig"

// 	"github.com/superseriousbusiness/gotosocial/internal/oauth"
// 	"golang.org/x/crypto/bcrypt"
// )

// type AccountCreateTestSuite struct {
// 	AccountStandardTestSuite
// }

// func (suite *AccountCreateTestSuite) SetupSuite() {
// 	testrig.StandardDBSetup(suite.db)
// 	suite.testTokens = testrig.NewTestTokens()
// 	suite.testClients = testrig.NewTestClients()
// 	suite.testApplications = testrig.NewTestApplications()
// 	suite.testUsers = testrig.NewTestUsers()
// 	suite.testAccounts = testrig.NewTestAccounts()
// 	suite.testAttachments = testrig.NewTestAttachments()
// 	suite.testStatuses = testrig.NewTestStatuses()
// }

// func (suite *AccountCreateTestSuite) SetupTest() {
// 	suite.config = testrig.NewTestConfig()
// 	suite.db = testrig.NewTestDB()
// 	suite.log = testrig.NewTestLog()
// 	suite.processor = testrig.NewTestProcessor(suite.db)
// 	suite.accountModule = account.New(suite.config, suite.processor, suite.log).(*account.Module)
// }

// func (suite *AccountCreateTestSuite) TearDownTest() {
// 	testrig.StandardDBTeardown(suite.db)
// }

// // TestAccountCreatePOSTHandlerSuccessful checks the happy path for an account creation request: all the fields provided are valid,
// // and at the end of it a new user and account should be added into the database.
// //
// // This is the handler served at /api/v1/accounts as POST
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerSuccessful() {

// 	t := suite.testTokens["local_account_1"]
// 	oauthToken := oauth.TokenToOauthToken(t)

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
// 	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath
// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response

// 	// 1. we should have OK from our call to the function
// 	suite.EqualValues(http.StatusOK, recorder.Code)

// 	// 2. we should have a token in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	t := &model.Token{}
// 	err = json.Unmarshal(b, t)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "we're authorized now!", t.AccessToken)

// 	// check new account

// 	// 1. we should be able to get the new account from the db
// 	acct := &gtsmodel.Account{}
// 	err = suite.db.GetLocalAccountByUsername("test_user", acct)
// 	assert.NoError(suite.T(), err)
// 	assert.NotNil(suite.T(), acct)
// 	// 2. reason should be set
// 	assert.Equal(suite.T(), suite.newUserFormHappyPath.Get("reason"), acct.Reason)
// 	// 3. display name should be equal to username by default
// 	assert.Equal(suite.T(), suite.newUserFormHappyPath.Get("username"), acct.DisplayName)
// 	// 4. domain should be nil because this is a local account
// 	assert.Nil(suite.T(), nil, acct.Domain)
// 	// 5. id should be set and parseable as a uuid
// 	assert.NotNil(suite.T(), acct.ID)
// 	_, err = uuid.Parse(acct.ID)
// 	assert.Nil(suite.T(), err)
// 	// 6. private and public key should be set
// 	assert.NotNil(suite.T(), acct.PrivateKey)
// 	assert.NotNil(suite.T(), acct.PublicKey)

// 	// check new user

// 	// 1. we should be able to get the new user from the db
// 	usr := &gtsmodel.User{}
// 	err = suite.db.GetWhere("unconfirmed_email", suite.newUserFormHappyPath.Get("email"), usr)
// 	assert.Nil(suite.T(), err)
// 	assert.NotNil(suite.T(), usr)

// 	// 2. user should have account id set to account we got above
// 	assert.Equal(suite.T(), acct.ID, usr.AccountID)

// 	// 3. id should be set and parseable as a uuid
// 	assert.NotNil(suite.T(), usr.ID)
// 	_, err = uuid.Parse(usr.ID)
// 	assert.Nil(suite.T(), err)

// 	// 4. locale should be equal to what we requested
// 	assert.Equal(suite.T(), suite.newUserFormHappyPath.Get("locale"), usr.Locale)

// 	// 5. created by application id should be equal to the app id
// 	assert.Equal(suite.T(), suite.testApplication.ID, usr.CreatedByApplicationID)

// 	// 6. password should be matcheable to what we set above
// 	err = bcrypt.CompareHashAndPassword([]byte(usr.EncryptedPassword), []byte(suite.newUserFormHappyPath.Get("password")))
// 	assert.Nil(suite.T(), err)
// }

// // TestAccountCreatePOSTHandlerNoAuth makes sure that the handler fails when no authorization is provided:
// // only registered applications can create accounts, and we don't provide one here.
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerNoAuth() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath
// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response

// 	// 1. we should have forbidden from our call to the function because we didn't auth
// 	suite.EqualValues(http.StatusForbidden, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"not authorized"}`, string(b))
// }

// // TestAccountCreatePOSTHandlerNoAuth makes sure that the handler fails when no form is provided at all.
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerNoForm() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response
// 	suite.EqualValues(http.StatusBadRequest, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"missing one or more required form values"}`, string(b))
// }

// // TestAccountCreatePOSTHandlerWeakPassword makes sure that the handler fails when a weak password is provided
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerWeakPassword() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath
// 	// set a weak password
// 	ctx.Request.Form.Set("password", "weak")
// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response
// 	suite.EqualValues(http.StatusBadRequest, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"insecure password, try including more special characters, using uppercase letters, using numbers or using a longer password"}`, string(b))
// }

// // TestAccountCreatePOSTHandlerWeirdLocale makes sure that the handler fails when a weird locale is provided
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerWeirdLocale() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath
// 	// set an invalid locale
// 	ctx.Request.Form.Set("locale", "neverneverland")
// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response
// 	suite.EqualValues(http.StatusBadRequest, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"language: tag is not well-formed"}`, string(b))
// }

// // TestAccountCreatePOSTHandlerRegistrationsClosed makes sure that the handler fails when registrations are closed
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerRegistrationsClosed() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath

// 	// close registrations
// 	suite.config.AccountsConfig.OpenRegistration = false
// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response
// 	suite.EqualValues(http.StatusBadRequest, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"registration is not open for this server"}`, string(b))
// }

// // TestAccountCreatePOSTHandlerReasonNotProvided makes sure that the handler fails when no reason is provided but one is required
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerReasonNotProvided() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath

// 	// remove reason
// 	ctx.Request.Form.Set("reason", "")

// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response
// 	suite.EqualValues(http.StatusBadRequest, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"no reason provided"}`, string(b))
// }

// // TestAccountCreatePOSTHandlerReasonNotProvided makes sure that the handler fails when a crappy reason is presented but a good one is required
// func (suite *AccountCreateTestSuite) TestAccountCreatePOSTHandlerInsufficientReason() {

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", account.BasePath), nil) // the endpoint we're hitting
// 	ctx.Request.Form = suite.newUserFormHappyPath

// 	// remove reason
// 	ctx.Request.Form.Set("reason", "just cuz")

// 	suite.accountModule.AccountCreatePOSTHandler(ctx)

// 	// check response
// 	suite.EqualValues(http.StatusBadRequest, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	b, err := ioutil.ReadAll(result.Body)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), `{"error":"reason should be at least 40 chars but 'just cuz' was 8"}`, string(b))
// }

// /*
// 	TESTING: AccountUpdateCredentialsPATCHHandler
// */

// func (suite *AccountCreateTestSuite) TestAccountUpdateCredentialsPATCHHandler() {

// 	// put test local account in db
// 	err := suite.db.Put(suite.testAccountLocal)
// 	assert.NoError(suite.T(), err)

// 	// attach avatar to request
// 	aviFile, err := os.Open("../../media/test/test-jpeg.jpg")
// 	assert.NoError(suite.T(), err)
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)

// 	part, err := writer.CreateFormFile("avatar", "test-jpeg.jpg")
// 	assert.NoError(suite.T(), err)

// 	_, err = io.Copy(part, aviFile)
// 	assert.NoError(suite.T(), err)

// 	err = aviFile.Close()
// 	assert.NoError(suite.T(), err)

// 	err = writer.Close()
// 	assert.NoError(suite.T(), err)

// 	// setup
// 	recorder := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(recorder)
// 	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccountLocal)
// 	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
// 	ctx.Request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:8080/%s", account.UpdateCredentialsPath), body) // the endpoint we're hitting
// 	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
// 	suite.accountModule.AccountUpdateCredentialsPATCHHandler(ctx)

// 	// check response

// 	// 1. we should have OK because our request was valid
// 	suite.EqualValues(http.StatusOK, recorder.Code)

// 	// 2. we should have an error message in the result body
// 	result := recorder.Result()
// 	defer result.Body.Close()
// 	// TODO: implement proper checks here
// 	//
// 	// b, err := ioutil.ReadAll(result.Body)
// 	// assert.NoError(suite.T(), err)
// 	// assert.Equal(suite.T(), `{"error":"not authorized"}`, string(b))
// }

// func TestAccountCreateTestSuite(t *testing.T) {
// 	suite.Run(t, new(AccountCreateTestSuite))
// }
