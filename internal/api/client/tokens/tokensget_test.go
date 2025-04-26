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
	"github.com/stretchr/testify/suite"
)

type TokensGetTestSuite struct {
	TokensStandardTestSuite
}

func (suite *TokensGetTestSuite) TestTokensGet() {
	var (
		testPath = "/api" + tokens.BasePath
	)

	out, code := suite.req(
		http.MethodGet,
		testPath,
		suite.tokens.TokensInfoGETHandler,
		nil,
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal(`[
  {
    "id": "01JN0X2D9GJTZQ5KYPYFWN16QW",
    "created_at": "2025-02-26T10:33:04.560Z",
    "scope": "push",
    "application": {
      "name": "really cool gts application",
      "website": "https://reallycool.app"
    }
  },
  {
    "id": "01F8MGTQW4DKTDF8SW5CT9HYGA",
    "created_at": "2021-06-20T10:53:00.164Z",
    "scope": "read write push",
    "application": {
      "name": "really cool gts application",
      "website": "https://reallycool.app"
    }
  }
]`, out)
}

func TestTokensGetTestSuite(t *testing.T) {
	suite.Run(t, new(TokensGetTestSuite))
}
