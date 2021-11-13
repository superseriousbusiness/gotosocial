package account_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"

	"codeberg.org/gruf/go-store/kv"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	config      *config.Config
	db          db.DB
	tc          typeutils.TypeConverter
	storage     *kv.KVStore
	federator   federation.Federator
	processor   processing.Processor
	emailSender email.Sender
	sentEmails  map[string]string

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status

	// module being tested
	accountModule *account.Module
}

func (suite *AccountStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *AccountStandardTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewTestStorage()
	testrig.InitTestLog()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db), suite.storage)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", suite.sentEmails)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender)
	suite.accountModule = account.New(suite.config, suite.processor).(*account.Module)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *AccountStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *AccountStandardTestSuite) newContext(recorder *httptest.ResponseRecorder, requestMethod string, requestBody []byte, requestPath string, bodyContentType string) *gin.Context {
	ctx, _ := gin.CreateTestContext(recorder)

	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	baseURI := fmt.Sprintf("%s://%s", suite.config.Protocol, suite.config.Host)
	requestURI := fmt.Sprintf("%s/%s", baseURI, requestPath)

	ctx.Request = httptest.NewRequest(http.MethodPatch, requestURI, bytes.NewReader(requestBody)) // the endpoint we're hitting

	if bodyContentType != "" {
		ctx.Request.Header.Set("Content-Type", bodyContentType)
	}

	return ctx
}
