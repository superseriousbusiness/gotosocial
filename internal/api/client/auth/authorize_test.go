package auth_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type AuthAuthorizeTestSuite struct {
	AuthStandardTestSuite
}

func (suite *AuthAuthorizeTestSuite) TestAccountAuthorizeGETHandler() {

	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodGet, auth.OauthAuthorizePath)

	// call the handler
	suite.authModule.AuthorizeGETHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("<p>this is my new bio read it and weep</p>", apimodelAccount.Note)
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AuthAuthorizeTestSuite))
}
