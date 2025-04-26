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

package tokens_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/tokens"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type TokensStandardTestSuite struct {
	suite.Suite

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testStructs      *testrig.TestStructs

	// module being tested
	tokens *tokens.Module
}

func (suite *TokensStandardTestSuite) req(
	httpMethod string,
	requestPath string,
	handler gin.HandlerFunc,
	pathParams map[string]string,
) (string, int) {
	var (
		recorder = httptest.NewRecorder()
		ctx, _   = testrig.CreateGinTestContext(recorder, nil)
	)

	// Prepare test context.
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// Prepare test context request.
	request := httptest.NewRequest(httpMethod, requestPath, nil)
	request.Header.Set("accept", "application/json")
	ctx.Request = request

	// Inject path parameters.
	if pathParams != nil {
		for k, v := range pathParams {
			ctx.AddParam(k, v)
		}
	}

	// Trigger the handler
	handler(ctx)

	// Read the response
	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Format as nice indented json.
	dst := &bytes.Buffer{}
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	return dst.String(), recorder.Code
}

func (suite *TokensStandardTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *TokensStandardTestSuite) SetupTest() {
	suite.testStructs = testrig.SetupTestStructs(
		"../../../../testrig/media",
		"../../../../web/template",
	)
	suite.tokens = tokens.New(suite.testStructs.Processor)
}

func (suite *TokensStandardTestSuite) TearDownTest() {
	testrig.TearDownTestStructs(suite.testStructs)
}
