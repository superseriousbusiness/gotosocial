package user_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func (suite *UserGetTestSuite) TestGetUser() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork"]
	targetAccount := suite.testAccounts["local_account_1"]

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator)
	userModule := user.New(suite.config, processor, suite.log).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.URI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("Signature", signedRequest.SignatureHeader)
	ctx.Request.Header.Set("Date", signedRequest.DateHeader)

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
			Value: targetAccount.Username,
		},
	}

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
	a, err := suite.tc.ASRepresentationToAccount(context.Background(), person, false)
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), targetAccount.Username, a.Username)
}

func TestUserGetTestSuite(t *testing.T) {
	suite.Run(t, new(UserGetTestSuite))
}
