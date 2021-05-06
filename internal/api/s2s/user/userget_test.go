package user_test

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type UserGetTestSuite struct {
	UserStandardTestSuite
}

func (suite *UserGetTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
}

func (suite *UserGetTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.storage = testrig.NewTestStorage()
	suite.log = testrig.NewTestLog()
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil)))
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator)
	suite.userModule = user.New(suite.config, suite.processor, suite.log).(*user.Module)
	testrig.StandardDBSetup(suite.db)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *UserGetTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *UserGetTestSuite) TestGetUser() {
	// the dereference we're gonna use
	signedRequest := testrig.NewTestDereferenceRequests(suite.testAccounts)["foss_satan_dereference_zork"]

	requestingAccount := suite.testAccounts["remote_account_1"]
	targetAccount := suite.testAccounts["local_account_1"]

	encodedPublicKey, err := x509.MarshalPKIXPublicKey(requestingAccount.PublicKey)
	assert.NoError(suite.T(), err)
	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: encodedPublicKey,
	})
	publicKeyString := strings.ReplaceAll(string(publicKeyBytes), "\n", "\\n")

	// for this test we need the client to return the public key of the requester on the 'remote' instance
	responseBodyString := fmt.Sprintf(`
	{
		"@context": [
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1"
		],

		"id": "%s",
		"type": "Person",
		"preferredUsername": "%s",
		"inbox": "%s",

		"publicKey": {
			"id": "%s",
			"owner": "%s",
			"publicKeyPem": "%s"
		}
	}`, requestingAccount.URI, requestingAccount.Username, requestingAccount.InboxURI, requestingAccount.PublicKeyURI, requestingAccount.URI, publicKeyString)

	// create a transport controller whose client will just return the response body string we specified above
	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(bytes.NewReader([]byte(responseBodyString)))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}))
	// get this transport controller embedded right in the user module we're testing
	federator := testrig.NewTestFederator(suite.db, tc)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator)
	userModule := user.New(suite.config, processor, suite.log).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080%s", strings.Replace(user.UsersBasePathWithUsername, ":username", targetAccount.Username, 1)), nil) // the endpoint we're hitting

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
			Value: targetAccount.Username,
		},
	}

	// we need these headers for the request to be validated
	ctx.Request.Header.Set("Signature", signedRequest.SignatureHeader)
	ctx.Request.Header.Set("Date", signedRequest.DateHeader)
	ctx.Request.Header.Set("Digest", signedRequest.DigestHeader)

	// trigger the function being tested
	userModule.UsersGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// should be a Person
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	person, ok := t.(vocab.ActivityStreamsPerson)
	assert.True(suite.T(), ok)

	// convert person to account
	// since this account is already known, we should get a pretty full model of it from the conversion
	a, err := suite.tc.ASPersonToAccount(person)
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), targetAccount.Username, a.Username)
}

func TestUserGetTestSuite(t *testing.T) {
	suite.Run(t, new(UserGetTestSuite))
}
