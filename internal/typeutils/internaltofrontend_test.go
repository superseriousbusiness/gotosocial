/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package typeutils_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type InternalToFrontendTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToFrontendTestSuite) TestAccountToFrontend() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test
	apiAccount, err := suite.typeconverter.AccountToAPIAccountPublic(context.Background(), testAccount)
	suite.NoError(err)
	suite.NotNil(apiAccount)

	b, err := json.Marshal(apiAccount)
	suite.NoError(err)
	suite.Equal(`{"id":"01F8MH1H7YV1Z7D2C8K2730QBF","username":"the_mighty_zork","acct":"the_mighty_zork","display_name":"original zork (he/they)","locked":false,"bot":false,"created_at":"2022-05-20T11:09:18Z","note":"\u003cp\u003ehey yo this is my profile!\u003c/p\u003e","url":"http://localhost:8080/@the_mighty_zork","avatar":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg","avatar_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg","header":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","header_static":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg","followers_count":2,"following_count":2,"statuses_count":5,"last_status_at":"2022-05-20T11:37:55Z","emojis":[],"fields":[]}`, string(b))
}

func TestInternalToFrontendTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToFrontendTestSuite))
}
