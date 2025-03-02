// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package auth_test

import (
	"context"
	"encoding/json"
	"io"
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

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: grant_type was not set in the token request form, but must be set to authorization_code or client_credentials: client_id was not set in the token request form: client_secret was not set in the token request form: redirect_uri was not set in the token request form"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveClientCredentialsOK() {
	testApp := suite.testApplications["application_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"grant_type":    {"client_credentials"},
			"client_id":     {testApp.ClientID},
			"client_secret": {testApp.ClientSecret},
			"redirect_uri":  {"http://localhost:8080"},
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

	b, err := io.ReadAll(result.Body)
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

func (suite *TokenTestSuite) TestRetrieveClientCredentialsBadScope() {
	testApp := suite.testApplications["application_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"grant_type":    {"client_credentials"},
			"client_id":     {testApp.ClientID},
			"client_secret": {testApp.ClientSecret},
			"redirect_uri":  {"http://localhost:8080"},
			"scope":         {"admin"},
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	ctx, recorder := suite.newContext(http.MethodPost, "oauth/token", bodyBytes, w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

	suite.authModule.TokenPOSTHandler(ctx)

	suite.Equal(http.StatusForbidden, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid_scope","error_description":"Forbidden: requested scope admin was not covered by client scope: If you arrived at this error during a sign in/oauth flow, please try clearing your session cookies and signing in again; if problems persist, make sure you're using the correct credentials"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveClientCredentialsDifferentRedirectURI() {
	testApp := suite.testApplications["application_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"grant_type":    {"client_credentials"},
			"client_id":     {testApp.ClientID},
			"client_secret": {testApp.ClientSecret},
			"redirect_uri":  {"http://somewhere.else.example.org"},
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	ctx, recorder := suite.newContext(http.MethodPost, "oauth/token", bodyBytes, w.FormDataContentType())
	ctx.Request.Header.Set("accept", "application/json")

	suite.authModule.TokenPOSTHandler(ctx)

	suite.Equal(http.StatusForbidden, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid redirect uri","error_description":"Forbidden: requested redirect URI http://somewhere.else.example.org was not covered by client redirect URIs: If you arrived at this error during a sign in/oauth flow, please try clearing your session cookies and signing in again; if problems persist, make sure you're using the correct credentials"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveAuthorizationCodeOK() {
	testApp := suite.testApplications["application_1"]
	testUserAuthorizationToken := suite.testTokens["local_account_1_user_authorization_token"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"grant_type":    {"authorization_code"},
			"client_id":     {testApp.ClientID},
			"client_secret": {testApp.ClientSecret},
			"redirect_uri":  {"http://localhost:8080"},
			"code":          {testUserAuthorizationToken.Code},
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

	b, err := io.ReadAll(result.Body)
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
	testApp := suite.testApplications["application_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"grant_type":    {"authorization_code"},
			"client_id":     {testApp.ClientID},
			"client_secret": {testApp.ClientSecret},
			"redirect_uri":  {"http://localhost:8080"},
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

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: code was not set in the token request form, but must be set since grant_type is authorization_code"}`, string(b))
}

func (suite *TokenTestSuite) TestRetrieveAuthorizationCodeWrongGrantType() {
	testApplication := suite.testApplications["application_1"]

	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"grant_type":    {"client_credentials"},
			"client_id":     {testApplication.ClientID},
			"client_secret": {testApplication.ClientSecret},
			"redirect_uri":  {"http://localhost:8080"},
			"code":          {"peepeepoopoo"},
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

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"invalid_request","error_description":"Bad Request: a code was provided in the token request form, but grant_type was not set to authorization_code"}`, string(b))
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, &TokenTestSuite{})
}
