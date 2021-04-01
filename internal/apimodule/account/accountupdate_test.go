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
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/models"
	oauthmodels "github.com/superseriousbusiness/oauth2/v4/models"
)

type AccountUpdateTestSuite struct {
	suite.Suite
	config               *config.Config
	log                  *logrus.Logger
	testAccountLocal     *model.Account
	testApplication      *model.Application
	testToken            oauth2.TokenInfo
	mockOauthServer      *oauth.MockServer
	mockStorage          *storage.MockStorage
	mediaHandler         media.MediaHandler
	db                   db.DB
	accountModule        *accountModule
	newUserFormHappyPath url.Values
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *AccountUpdateTestSuite) SetupSuite() {
	// some of our subsequent entities need a log so create this here
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log

	suite.testAccountLocal = &model.Account{
		ID:       uuid.NewString(),
		Username: "test_user",
	}

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

	suite.mockStorage = &storage.MockStorage{}
	// We don't need storage to do anything for these tests, so just simulate a success and do nothing -- we won't need to return anything from storage
	suite.mockStorage.On("StoreFileAt", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	// set a media handler because some handlers (eg update credentials) need to upload media (new header/avatar)
	suite.mediaHandler = media.New(suite.config, suite.db, suite.mockStorage, log)

	// and finally here's the thing we're actually testing!
	suite.accountModule = New(suite.config, suite.db, suite.mockOauthServer, suite.mediaHandler, suite.log).(*accountModule)
}

func (suite *AccountUpdateTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

// SetupTest creates a db connection and creates necessary tables before each test
func (suite *AccountUpdateTestSuite) SetupTest() {
	// create all the tables we might need in thie suite
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.FollowRequest{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
		&model.MediaAttachment{},
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
func (suite *AccountUpdateTestSuite) TearDownTest() {

	// remove all the tables we might have used so it's clear for the next test
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.FollowRequest{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
		&model.MediaAttachment{},
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

/*
	TESTING: AccountUpdateCredentialsPATCHHandler
*/

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandler() {

	// put test local account in db
	err := suite.db.Put(suite.testAccountLocal)
	assert.NoError(suite.T(), err)

	// attach avatar to request form
	avatarFile, err := os.Open("../../media/test/test-jpeg.jpg")
	assert.NoError(suite.T(), err)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	avatarPart, err := writer.CreateFormFile("avatar", "test-jpeg.jpg")
	assert.NoError(suite.T(), err)

	_, err = io.Copy(avatarPart, avatarFile)
	assert.NoError(suite.T(), err)

	err = avatarFile.Close()
	assert.NoError(suite.T(), err)

	// set display name to a new value
	displayNamePart, err := writer.CreateFormField("display_name")
	assert.NoError(suite.T(), err)

	_, err = io.Copy(displayNamePart, bytes.NewBufferString("test_user_wohoah"))
	assert.NoError(suite.T(), err)

	// set locked to true
	lockedPart, err := writer.CreateFormField("locked")
	assert.NoError(suite.T(), err)

	_, err = io.Copy(lockedPart, bytes.NewBufferString("true"))
	assert.NoError(suite.T(), err)

	// close the request writer, the form is now prepared
	err = writer.Close()
	assert.NoError(suite.T(), err)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccountLocal)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:8080/%s", updateCredentialsPath), body) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	suite.accountModule.accountUpdateCredentialsPATCHHandler(ctx)

	// check response

	// 1. we should have OK because our request was valid
	suite.EqualValues(http.StatusOK, recorder.Code)

	// 2. we should have an error message in the result body
	result := recorder.Result()
	defer result.Body.Close()
	// TODO: implement proper checks here
	//
	// b, err := ioutil.ReadAll(result.Body)
	// assert.NoError(suite.T(), err)
	// assert.Equal(suite.T(), `{"error":"not authorized"}`, string(b))
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
