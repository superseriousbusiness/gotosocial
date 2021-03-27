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

package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/models"
	oauthmodels "github.com/superseriousbusiness/oauth2/v4/models"
	"golang.org/x/crypto/bcrypt"
)

type AccountTestSuite struct {
	suite.Suite
	config               *config.Config
	log                  *logrus.Logger
	testAccountLocal     *model.Account
	testAccountRemote    *model.Account
	testUser             *model.User
	testApplication      *model.Application
	testToken            oauth2.TokenInfo
	mockOauthServer      *oauth.MockServer
	db                   db.DB
	accountModule        *accountModule
	newUserFormHappyPath url.Values
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *AccountTestSuite) SetupSuite() {
	// some of our subsequent entities need a log so create this here
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log

	// can use this test application throughout
	suite.testApplication = &model.Application{
		ID:           "weeweeeeeeeeeeeeee",
		Name:         "a test application",
		Website:      "https://some-application-website.com",
		RedirectURI:  "http://localhost:8080",
		ClientID:     "a-known-client-id",
		ClientSecret: "some-secret",
		Scopes:       "read",
		VapidKey:     "aaaaaa-aaaaaaaa-aaaaaaaaaaa",
	}

	// can use this test token throughout
	suite.testToken = &oauthmodels.Token{
		ClientID:      "a-known-client-id",
		RedirectURI:   "http://localhost:8080",
		Scope:         "read",
		Code:          "123456789",
		CodeCreateAt:  time.Now(),
		CodeExpiresIn: time.Duration(10 * time.Minute),
	}

	// Direct config to local postgres instance
	c := config.Empty()
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	suite.config = c

	// use an actual database for this, because it's just easier than mocking one out
	database, err := db.New(context.Background(), c, log)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.db = database

	// we need to mock the oauth server because account creation needs it to create a new token
	suite.mockOauthServer = &oauth.MockServer{}
	suite.mockOauthServer.On("GenerateUserAccessToken", suite.testToken, suite.testApplication.ClientSecret, mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		l := suite.log.WithField("func", "GenerateUserAccessToken")
		token := args.Get(0).(oauth2.TokenInfo)
		l.Infof("received token %+v", token)
		clientSecret := args.Get(1).(string)
		l.Infof("received clientSecret %+v", clientSecret)
		userID := args.Get(2).(string)
		l.Infof("received userID %+v", userID)
	}).Return(&models.Token{
		Code: "we're authorized now!",
	}, nil)

	// and finally here's the thing we're actually testing!
	suite.accountModule = New(suite.config, suite.db, suite.mockOauthServer, suite.log).(*accountModule)
}

func (suite *AccountTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

// SetupTest creates a db connection and creates necessary tables before each test
func (suite *AccountTestSuite) SetupTest() {
	// create all the tables we might need in thie suite
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
	}
	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}

	// form to submit for happy path account create requests -- this will be changed inside tests so it's better to set it before each test
	suite.newUserFormHappyPath = url.Values{
		"reason":    []string{"a very good reason that's at least 40 characters i swear"},
		"username":  []string{"test_user"},
		"email":     []string{"user@example.org"},
		"password":  []string{"very-strong-password"},
		"agreement": []string{"true"},
		"locale":    []string{"en"},
	}

	// same with accounts config
	suite.config.AccountsConfig = &config.AccountsConfig{
		OpenRegistration: true,
		RequireApproval:  true,
		ReasonRequired:   true,
	}
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *AccountTestSuite) TearDownTest() {

	// remove all the tables we might have used so it's clear for the next test
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
}

/*
	ACTUAL TESTS
*/

// TestAccountCreatePOSTHandlerSuccessful checks the happy path for an account creation request: all the fields provided are valid,
// and at the end of it a new user and account should be added into the database.
//
// This is the handler served at /api/v1/accounts as POST
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerSuccessful() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath
	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response

	// 1. we should have OK from our call to the function
	suite.EqualValues(http.StatusOK, recorder.Code)

	// 2. we should have a token in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	t := &mastotypes.Token{}
	err = json.Unmarshal(b, t)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "we're authorized now!", t.AccessToken)

	// check new account

	// 1. we should be able to get the new account from the db
	acct := &model.Account{}
	err = suite.db.GetWhere("username", "test_user", acct)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), acct)
	// 2. reason should be set
	assert.Equal(suite.T(), suite.newUserFormHappyPath.Get("reason"), acct.Reason)
	// 3. display name should be equal to username by default
	assert.Equal(suite.T(), suite.newUserFormHappyPath.Get("username"), acct.DisplayName)
	// 4. domain should be nil because this is a local account
	assert.Nil(suite.T(), nil, acct.Domain)
	// 5. id should be set and parseable as a uuid
	assert.NotNil(suite.T(), acct.ID)
	_, err = uuid.Parse(acct.ID)
	assert.Nil(suite.T(), err)
	// 6. private and public key should be set
	assert.NotNil(suite.T(), acct.PrivateKey)
	assert.NotNil(suite.T(), acct.PublicKey)

	// check new user

	// 1. we should be able to get the new user from the db
	usr := &model.User{}
	err = suite.db.GetWhere("unconfirmed_email", suite.newUserFormHappyPath.Get("email"), usr)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), usr)

	// 2. user should have account id set to account we got above
	assert.Equal(suite.T(), acct.ID, usr.AccountID)

	// 3. id should be set and parseable as a uuid
	assert.NotNil(suite.T(), usr.ID)
	_, err = uuid.Parse(usr.ID)
	assert.Nil(suite.T(), err)

	// 4. locale should be equal to what we requested
	assert.Equal(suite.T(), suite.newUserFormHappyPath.Get("locale"), usr.Locale)

	// 5. created by application id should be equal to the app id
	assert.Equal(suite.T(), suite.testApplication.ID, usr.CreatedByApplicationID)

	// 6. password should be matcheable to what we set above
	err = bcrypt.CompareHashAndPassword([]byte(usr.EncryptedPassword), []byte(suite.newUserFormHappyPath.Get("password")))
	assert.Nil(suite.T(), err)
}

// TestAccountCreatePOSTHandlerNoAuth makes sure that the handler fails when no authorization is provided:
// only registered applications can create accounts, and we don't provide one here.
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerNoAuth() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath
	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response

	// 1. we should have forbidden from our call to the function because we didn't auth
	suite.EqualValues(http.StatusForbidden, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"not authorized"}`, string(b))
}

// TestAccountCreatePOSTHandlerNoAuth makes sure that the handler fails when no form is provided at all.
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerNoForm() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"missing one or more required form values"}`, string(b))
}

// TestAccountCreatePOSTHandlerWeakPassword makes sure that the handler fails when a weak password is provided
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerWeakPassword() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath
	// set a weak password
	ctx.Request.Form.Set("password", "weak")
	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"insecure password, try including more special characters, using uppercase letters, using numbers or using a longer password"}`, string(b))
}

// TestAccountCreatePOSTHandlerWeirdLocale makes sure that the handler fails when a weird locale is provided
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerWeirdLocale() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath
	// set an invalid locale
	ctx.Request.Form.Set("locale", "neverneverland")
	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"language: tag is not well-formed"}`, string(b))
}

// TestAccountCreatePOSTHandlerRegistrationsClosed makes sure that the handler fails when registrations are closed
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerRegistrationsClosed() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath

	// close registrations
	suite.config.AccountsConfig.OpenRegistration = false
	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"registration is not open for this server"}`, string(b))
}

// TestAccountCreatePOSTHandlerReasonNotProvided makes sure that the handler fails when no reason is provided but one is required
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerReasonNotProvided() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath

	// remove reason
	ctx.Request.Form.Set("reason", "")

	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"no reason provided"}`, string(b))
}

// TestAccountCreatePOSTHandlerReasonNotProvided makes sure that the handler fails when a crappy reason is presented but a good one is required
func (suite *AccountTestSuite) TestAccountCreatePOSTHandlerInsufficientReason() {

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), nil) // the endpoint we're hitting
	ctx.Request.Form = suite.newUserFormHappyPath

	// remove reason
	ctx.Request.Form.Set("reason", "just cuz")

	suite.accountModule.accountCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusBadRequest, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"reason should be at least 40 chars but 'just cuz' was 8"}`, string(b))
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
