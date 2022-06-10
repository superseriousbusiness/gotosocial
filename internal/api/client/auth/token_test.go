package auth_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type TokenTestSuite struct {
	AuthStandardTestSuite
}

func (suite *TokenTestSuite) TestPOSTTokenEmptyForm() {
	ctx, recorder := suite.newContext(http.MethodPost, "oauth/token", []byte{}, "")
	ctx.Request.Header.Set("accept", "application/json")

	suite.authModule.TokenPOSTHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: grant_type was not set in the token request, but must be set to authorization_code or client_credentials: code was not set in the token request body: redirect_uri was not set in the token request body: client_id was not set in the token request body"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveApplicationToken() {
	testClient := suite.testClients["local_account_1"]
	// testApplication := suite.testApplications["application_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     testClient.ID,
			"client_secret": testClient.Secret,
			"redirect_uri":  "http://localhost:8080",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	ctx, recorder := suite.newContext(http.MethodPost, "oauth/token", bodyBytes, w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

	suite.authModule.TokenPOSTHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: could not validate token request: invalid_request"}`, string(b))
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, &TokenTestSuite{})
}
