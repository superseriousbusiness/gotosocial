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
	// standard suite interfaces
	suite.Suite
	config         *config.Config
	db             db.DB
	log            *logrus.Logger
	storage        storage.Storage
	mastoConverter mastotypes.Converter
	mediaHandler   media.MediaHandler
	oauthServer    oauth.Server

	// standard suite models
	testTokens       map[string]*oauth.Token
	testClients      map[string]*oauth.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment

	// item being tested
	mediaModule *mediaModule
}

/*
	TEST INFRASTRUCTURE
*/

func (suite *MediaCreateTestSuite) SetupSuite() {
	// setup standard items
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.mastoConverter = testrig.NewTestMastoConverter(suite.db)
	suite.mediaHandler = testrig.NewTestMediaHandler(suite.db, suite.storage)
	suite.oauthServer = testrig.NewTestOauthServer(suite.db)

	// setup module being tested
	suite.mediaModule = New(suite.db, suite.mediaHandler, suite.mastoConverter, suite.config, suite.log).(*mediaModule)
}

func (suite *MediaCreateTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

func (suite *MediaCreateTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
	testrig.StandardStorageSetup(suite.storage, "../../../testrig/media")
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
}

func (suite *MediaCreateTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

/*
	ACTUAL TESTS
*/

func (suite *MediaCreateTestSuite) TestStatusCreatePOSTImageHandlerSuccessful() {

	// set up the context for the request
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.PGTokenToOauthToken(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// see what's in storage *before* the request
	storageKeysBeforeRequest, err := suite.storage.ListKeys()
	if err != nil {
		panic(err)
	}

	// create the request
	buf, w, err := testrig.CreateMultipartFormData("file", "../../../testrig/media/test-jpeg.jpg", map[string]string{
		"description": "this is a test image -- a cool background from somewhere",
		"focus":       "-0.5,0.5",
	})
	if err != nil {
		panic(err)
	}
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/%s", basePath), bytes.NewReader(buf.Bytes())) // the endpoint we're hitting
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())

	// do the actual request
	suite.mediaModule.mediaCreatePOSTHandler(ctx)

	// check what's in storage *after* the request
	storageKeysAfterRequest, err := suite.storage.ListKeys()
	if err != nil {
		panic(err)
	}

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
	assert.Equal(suite.T(), len(storageKeysBeforeRequest) + 2, len(storageKeysAfterRequest)) // 2 images should be added to storage: the original and the thumbnail
}

func TestMediaCreateTestSuite(t *testing.T) {
	suite.Run(t, new(MediaCreateTestSuite))
}
