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

package user_test

import (
	"context"
	"net"
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"github.com/stretchr/testify/suite"
)

type CreateTestSuite struct {
	UserStandardTestSuite
}

func (suite *CreateTestSuite) TestCreateOK() {
	var (
		ctx      = context.Background()
		app      = suite.testApps["application_1"]
		appToken = suite.testTokens["local_account_1_client_application_token"]
		form     = &apimodel.AccountCreateRequest{
			Reason:    "a long enough explanation of why I am doing api calls",
			Username:  "someone_new",
			Email:     "someone_new@example.org",
			Password:  "a long enough password for this endpoint",
			Agreement: true,
			Locale:    "en-us",
			IP:        net.ParseIP("192.0.2.128"),
		}
	)

	// Create user via the API endpoint.
	user, errWithCode := suite.user.Create(ctx, app, form)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Load the app-level access token that was just used.
	appAccessToken, err := suite.oauthServer.LoadAccessToken(ctx, appToken.Access)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Create a user-level access token for the new user.
	userAccessToken, err := suite.user.TokenForNewUser(ctx, appAccessToken, app, user)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check returned user-level access token.
	suite.NotEmpty(userAccessToken.AccessToken)
	suite.Equal("Bearer", userAccessToken.TokenType)
}

func TestCreateTestSuite(t *testing.T) {
	suite.Run(t, &CreateTestSuite{})
}
