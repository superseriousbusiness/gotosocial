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

package accounts_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/accounts"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type AccountVerifyTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountVerifyTestSuite) TestAccountVerifyGet() {
	testAccount := suite.testAccounts["local_account_1"]

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodGet, nil, accounts.VerifyPath, "")

	// call the handler
	suite.accountsModule.AccountVerifyGETHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	createdAt, err := time.Parse(time.RFC3339, apimodelAccount.CreatedAt)
	suite.NoError(err)

	suite.Equal(testAccount.ID, apimodelAccount.ID)
	suite.Equal(testAccount.Username, apimodelAccount.Username)
	suite.Equal(testAccount.Username, apimodelAccount.Acct)
	suite.Equal(testAccount.DisplayName, apimodelAccount.DisplayName)
	suite.Equal(*testAccount.Locked, apimodelAccount.Locked)
	suite.Equal(*testAccount.Bot, apimodelAccount.Bot)
	suite.WithinDuration(testAccount.CreatedAt, createdAt, 30*time.Second) // we lose a bit of accuracy serializing so fuzz this a bit
	suite.Equal(testAccount.URL, apimodelAccount.URL)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg", apimodelAccount.Avatar)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg", apimodelAccount.AvatarStatic)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg", apimodelAccount.Header)
	suite.Equal("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg", apimodelAccount.HeaderStatic)
	suite.Equal(2, apimodelAccount.FollowersCount)
	suite.Equal(2, apimodelAccount.FollowingCount)
	suite.Equal(5, apimodelAccount.StatusesCount)
	suite.EqualValues(gtsmodel.VisibilityPublic, apimodelAccount.Source.Privacy)
	suite.Equal(testAccount.Language, apimodelAccount.Source.Language)
	suite.Equal(testAccount.NoteRaw, apimodelAccount.Source.Note)
}

func TestAccountVerifyTestSuite(t *testing.T) {
	suite.Run(t, new(AccountVerifyTestSuite))
}
