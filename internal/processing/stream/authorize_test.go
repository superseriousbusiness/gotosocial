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

package stream_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"github.com/stretchr/testify/suite"
)

type AuthorizeTestSuite struct {
	StreamTestSuite
}

func (suite *AuthorizeTestSuite) TestAuthorize() {
	account1, err := suite.streamProcessor.Authorize(suite.T().Context(), suite.testTokens["local_account_1"].Access)
	suite.NoError(err)
	suite.Equal(suite.testAccounts["local_account_1"].ID, account1.ID)

	account2, err := suite.streamProcessor.Authorize(suite.T().Context(), suite.testTokens["local_account_2"].Access)
	suite.NoError(err)
	suite.Equal(suite.testAccounts["local_account_2"].ID, account2.ID)

	noAccount, err := suite.streamProcessor.Authorize(suite.T().Context(), "aaaaaaaaaaaaaaaaaaaaa!!")
	suite.EqualError(err, "could not load access token: "+db.ErrNoEntries.Error())
	suite.Nil(noAccount)
}

func TestAuthorizeTestSuite(t *testing.T) {
	suite.Run(t, &AuthorizeTestSuite{})
}
