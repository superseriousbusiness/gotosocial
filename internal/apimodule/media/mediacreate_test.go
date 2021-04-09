package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	mastomodel "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaCreateTestSuite struct {
	suite.Suite
	config           *config.Config
	mockOauthServer  *oauth.MockServer
	mockStorage      *storage.MockStorage
	mediaHandler     media.MediaHandler
	mastoConverter   mastotypes.Converter
	testTokens       map[string]*oauth.Token
	testClients      map[string]*oauth.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	log              *logrus.Logger
	db               db.DB
	mediaModule      *mediaModule
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *MediaCreateTestSuite) SetupSuite() {
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
	suite.mockStorage.On("StoreFileAt", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil) // just pretend to store
	suite.mediaHandler = media.New(suite.config, suite.db, suite.mockStorage, log)
	suite.mastoConverter = mastotypes.New(suite.config, suite.db)
	suite.mediaModule = New(suite.db, suite.mediaHandler, suite.mastoConverter, suite.config, suite.log).(*mediaModule)
}

func (suite *MediaCreateTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

func (suite *MediaCreateTestSuite) SetupTest() {
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
func (suite *MediaCreateTestSuite) TearDownTest() {
	if err := testrig.StandardDBTeardown(suite.db); err != nil {
		panic(err)
	}
}

/*
	ACTUAL TESTS
*/

func (suite *MediaCreateTestSuite) TestStatusCreatePOSTImageHandlerSuccessful() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.PGTokenToOauthToken(t)

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	buf, w, err := testrig.CreateMultipartFormData("file", "../../media/test/test-jpeg.jpg", map[string]string{
		"description": "this is a test image -- a cool background from somewhere",
		"focus":       "-0.5,0.5",
	})
	if err != nil {
		panic(err)
	}

	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
	suite.mediaModule.mediaCreatePOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusAccepted, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	fmt.Println(string(b))

	attachmentReply := &mastomodel.Attachment{}
	err = json.Unmarshal(b, attachmentReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "this is a test image -- a cool background from somewhere", attachmentReply.Description)
	assert.Equal(suite.T(), "image", attachmentReply.Type)
	assert.EqualValues(suite.T(), mastomodel.MediaMeta{
		Original: mastomodel.MediaDimensions{
			Width:  1920,
			Height: 1080,
			Size:   "1920x1080",
			Aspect: 1.7777778,
		},
		Small: mastomodel.MediaDimensions{
			Width:  256,
			Height: 144,
			Size:   "256x144",
			Aspect: 1.7777778,
		},
		Focus: mastomodel.MediaFocus{
			X: -0.5,
			Y: 0.5,
		},
	}, attachmentReply.Meta)
	assert.Equal(suite.T(), "LjCZnlvyRkRn_NvzRjWF?urqV@f9", attachmentReply.Blurhash)
	assert.NotEmpty(suite.T(), attachmentReply.ID)
	assert.NotEmpty(suite.T(), attachmentReply.URL)
	assert.NotEmpty(suite.T(), attachmentReply.PreviewURL)
}

func TestMediaCreateTestSuite(t *testing.T) {
	suite.Run(t, new(MediaCreateTestSuite))
}
