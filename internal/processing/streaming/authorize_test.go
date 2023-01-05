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

package streaming_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthorizeTestSuite struct {
	StreamingTestSuite
}

func (suite *AuthorizeTestSuite) TestAuthorize() {
	account1, err := suite.streamingProcessor.AuthorizeStreamingRequest(context.Background(), suite.testTokens["local_account_1"].Access)
	suite.NoError(err)
	suite.Equal(suite.testAccounts["local_account_1"].ID, account1.ID)

	account2, err := suite.streamingProcessor.AuthorizeStreamingRequest(context.Background(), suite.testTokens["local_account_2"].Access)
	suite.NoError(err)
	suite.Equal(suite.testAccounts["local_account_2"].ID, account2.ID)

	noAccount, err := suite.streamingProcessor.AuthorizeStreamingRequest(context.Background(), "aaaaaaaaaaaaaaaaaaaaa!!")
	suite.EqualError(err, "could not load access token: no entries")
	suite.Nil(noAccount)
}

func TestAuthorizeTestSuite(t *testing.T) {
	suite.Run(t, &AuthorizeTestSuite{})
}
