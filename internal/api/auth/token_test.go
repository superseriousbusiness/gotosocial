/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package auth_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: grant_type was not set in the token request form, but must be set to authorization_code or client_credentials: client_id was not set in the token request form: client_secret was not set in the token request form: redirect_uri was not set in the token request form"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveClientCredentialsOK() {
	testClient := suite.testClients["local_account_1"]

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

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	t := &apimodel.Token{}
	err = json.Unmarshal(b, t)
	suite.NoError(err)

	suite.Equal("Bearer", t.TokenType)
	suite.NotEmpty(t.AccessToken)
	suite.NotEmpty(t.CreatedAt)
	suite.WithinDuration(time.Now(), time.Unix(t.CreatedAt, 0), 1*time.Minute)

	// there should be a token in the database now too
	dbToken := &gtsmodel.Token{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "access", Value: t.AccessToken}}, dbToken)
	suite.NoError(err)
	suite.NotNil(dbToken)
}

func (suite *TokenTestSuite) TestRetrieveAuthorizationCodeOK() {
	testClient := suite.testClients["local_account_1"]
	testUserAuthorizationToken := suite.testTokens["local_account_1_user_authorization_token"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     testClient.ID,
			"client_secret": testClient.Secret,
			"redirect_uri":  "http://localhost:8080",
			"code":          testUserAuthorizationToken.Code,
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	ctx, recorder := suite.newContext(http.MethodPost, "oauth/token", bodyBytes, w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

	suite.authModule.TokenPOSTHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	t := &apimodel.Token{}
	err = json.Unmarshal(b, t)
	suite.NoError(err)

	suite.Equal("Bearer", t.TokenType)
	suite.NotEmpty(t.AccessToken)
	suite.NotEmpty(t.CreatedAt)
	suite.WithinDuration(time.Now(), time.Unix(t.CreatedAt, 0), 1*time.Minute)

	dbToken := &gtsmodel.Token{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "access", Value: t.AccessToken}}, dbToken)
	suite.NoError(err)
	suite.NotNil(dbToken)
}

func (suite *TokenTestSuite) TestRetrieveAuthorizationCodeNoCode() {
	testClient := suite.testClients["local_account_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"grant_type":    "authorization_code",
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

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: code was not set in the token request form, but must be set since grant_type is authorization_code"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveAuthorizationCodeWrongGrantType() {
	testClient := suite.testClients["local_account_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     testClient.ID,
			"client_secret": testClient.Secret,
			"redirect_uri":  "http://localhost:8080",
			"code":          "peepeepoopoo",
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

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: a code was provided in the token request form, but grant_type was not set to authorization_code"}`, string(b))
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, &TokenTestSuite{})
}
