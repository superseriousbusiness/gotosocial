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

package federatingdb_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FollowersTestSuite struct {
	FederatingDBTestSuite
}

func (suite *FollowersTestSuite) TestGetFollowers() {
	testAccount := suite.testAccounts["local_account_2"]

	f, err := suite.federatingDB.Followers(context.Background(), testrig.URLMustParse(testAccount.URI))
	suite.NoError(err)

	fi, err := streams.Serialize(f)
	suite.NoError(err)

	fJson, err := json.Marshal(fi)
	suite.NoError(err)

	// zork follows local_account_2 so this should be reflected in the response
	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","items":"http://localhost:8080/users/the_mighty_zork","type":"Collection"}`, string(fJson))
}

func TestFollowersTestSuite(t *testing.T) {
	suite.Run(t, &FollowersTestSuite{})
}
