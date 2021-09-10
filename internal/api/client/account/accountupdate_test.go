/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package account_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountUpdateTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateCredentialsPATCHHandlerSimple() {
	// set up the request
	// we're updating the header image, the display name, and the locked status of zork
	// we're removing the note/bio
	requestBody, w, err := testrig.CreateMultipartFormData(
		"header", "../../../../testrig/media/test-jpeg.jpg",
		map[string]string{
			"display_name": "updated zork display name!!!",
			"note":         "",
			"locked":       "true",
		})
	if err != nil {
		panic(err)
	}
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, requestBody.Bytes(), account.UpdateCredentialsPath, w.FormDataContentType())

	// call the handler
	suite.accountModule.AccountUpdateCredentialsPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	fmt.Println(string(b))

	// unmarshal the returned account
	apimodelAccount := &apimodel.Account{}
	err = json.Unmarshal(b, apimodelAccount)
	suite.NoError(err)

	// check the returned api model account
	// fields should be updated
	suite.Equal("updated zork display name!!!", apimodelAccount.DisplayName)
	suite.True(apimodelAccount.Locked)
	suite.Empty(apimodelAccount.Note)

	// header values...
	// should be set
	suite.NotEmpty(apimodelAccount.Header)
	suite.NotEmpty(apimodelAccount.HeaderStatic)

	// should be different from the values set before
	suite.NotEqual("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg", apimodelAccount.Header)
	suite.NotEqual("http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg", apimodelAccount.HeaderStatic)
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
