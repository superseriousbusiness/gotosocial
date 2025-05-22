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
	"net/http"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/tokens"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"github.com/stretchr/testify/suite"
)

type TokenInvalidateTestSuite struct {
	TokensStandardTestSuite
}

func (suite *TokenInvalidateTestSuite) TestTokenInvalidate() {
	var (
		testToken = suite.testTokens["local_account_1"]
		testPath  = "/api" + tokens.BasePath + "/" + testToken.ID + "/invalidate"
	)

	out, code := suite.req(
		http.MethodPost,
		testPath,
		suite.tokens.TokenInvalidatePOSTHandler,
		map[string]string{"id": testToken.ID},
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal(`{
  "id": "01F8MGTQW4DKTDF8SW5CT9HYGA",
  "created_at": "2021-06-20T10:53:00.164Z",
  "scope": "read write push",
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  }
}`, out)

	// Check database for token we
	// just invalidated, should be gone.
	_, err := suite.testStructs.State.DB.GetTokenByID(
		suite.T().Context(), testToken.ID,
	)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *TokenInvalidateTestSuite) TestTokenInvalidateNotOurs() {
	var (
		testToken = suite.testTokens["admin_account"]
		testPath  = "/api" + tokens.BasePath + "/" + testToken.ID + "/invalidate"
	)

	out, code := suite.req(
		http.MethodGet,
		testPath,
		suite.tokens.TokenInfoGETHandler,
		map[string]string{"id": testToken.ID},
	)

	suite.Equal(http.StatusNotFound, code)
	suite.Equal(`{
  "error": "Not Found"
}`, out)
}

func TestTokenInvalidateTestSuite(t *testing.T) {
	suite.Run(t, new(TokenInvalidateTestSuite))
}
