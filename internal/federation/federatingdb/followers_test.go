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

package federatingdb_test

import (
	"context"
	"encoding/json"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type FollowersTestSuite struct {
	FederatingDBTestSuite
}

func (suite *FollowersTestSuite) TestGetFollowers() {
	testAccount := suite.testAccounts["local_account_2"]

	f, err := suite.federatingDB.Followers(context.Background(), testrig.URLMustParse(testAccount.URI))
	suite.NoError(err)

	fi, err := ap.Serialize(f)
	suite.NoError(err)

	fJson, err := json.MarshalIndent(fi, "", "  ")
	suite.NoError(err)

	// zork follows local_account_2 so this should be reflected in the response
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "items": "http://localhost:8080/users/the_mighty_zork",
  "type": "Collection"
}`, string(fJson))
}

func TestFollowersTestSuite(t *testing.T) {
	suite.Run(t, &FollowersTestSuite{})
}
