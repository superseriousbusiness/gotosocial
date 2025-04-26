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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type RevokeTestSuite struct {
	AuthStandardTestSuite
}

func (suite *RevokeTestSuite) TestRevokeOK() {
	var (
		app   = suite.testApplications["application_1"]
		token = suite.testTokens["local_account_1"]
	)

	// Prepare request form.
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"token":         {token.Access},
			"client_id":     {app.ClientID},
			"client_secret": {app.ClientSecret},
		})
	if err != nil {
		panic(err)
	}

	// Prepare request ctx.
	ctx, recorder := suite.newContext(
		http.MethodPost,
		"/oauth/revoke",
		requestBody.Bytes(),
		w.FormDataContentType(),
	)

	// Submit the revoke request.
	suite.authModule.TokenRevokePOSTHandler(ctx)

	// Check response code.
	// We don't really care about body.
	suite.Equal(http.StatusOK, recorder.Code)
	result := recorder.Result()
	defer result.Body.Close()

	// Ensure token now gone.
	_, err = suite.state.DB.GetTokenByAccess(
		context.Background(),
		token.Access,
	)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *RevokeTestSuite) TestRevokeWrongSecret() {
	var (
		app   = suite.testApplications["application_1"]
		token = suite.testTokens["local_account_1"]
	)

	// Prepare request form.
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"token":         {token.Access},
			"client_id":     {app.ClientID},
			"client_secret": {"Not the right secret :( :( :("},
		})
	if err != nil {
		panic(err)
	}

	// Prepare request ctx.
	ctx, recorder := suite.newContext(
		http.MethodPost,
		"/oauth/revoke",
		requestBody.Bytes(),
		w.FormDataContentType(),
	)

	// Submit the revoke request.
	suite.authModule.TokenRevokePOSTHandler(ctx)

	// Check response code + body.
	suite.Equal(http.StatusForbidden, recorder.Code)
	result := recorder.Result()
	defer result.Body.Close()

	// Read json bytes.
	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Indent nicely.
	dst := bytes.Buffer{}
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "error": "unauthorized_client",
  "error_description": "Forbidden: You are not authorized to revoke this token"
}`, dst.String())

	// Ensure token still there.
	_, err = suite.state.DB.GetTokenByAccess(
		context.Background(),
		token.Access,
	)
	suite.NoError(err)
}

func (suite *RevokeTestSuite) TestRevokeNoClientID() {
	var (
		app   = suite.testApplications["application_1"]
		token = suite.testTokens["local_account_1"]
	)

	// Prepare request form.
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"token":         {token.Access},
			"client_secret": {app.ClientSecret},
		})
	if err != nil {
		panic(err)
	}

	// Prepare request ctx.
	ctx, recorder := suite.newContext(
		http.MethodPost,
		"/oauth/revoke",
		requestBody.Bytes(),
		w.FormDataContentType(),
	)

	// Submit the revoke request.
	suite.authModule.TokenRevokePOSTHandler(ctx)

	// Check response code + body.
	suite.Equal(http.StatusBadRequest, recorder.Code)
	result := recorder.Result()
	defer result.Body.Close()

	// Read json bytes.
	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Indent nicely.
	dst := bytes.Buffer{}
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{
  "error": "invalid_request",
  "error_description": "Bad Request: client_id not set"
}`, dst.String())

	// Ensure token still there.
	_, err = suite.state.DB.GetTokenByAccess(
		context.Background(),
		token.Access,
	)
	suite.NoError(err)
}

func TestRevokeTestSuite(t *testing.T) {
	suite.Run(t, new(RevokeTestSuite))
}
