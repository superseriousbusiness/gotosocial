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

package account_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type AccountStatusesTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountStatusesTestSuite) TestGetStatusesMediaOnly() {
	// set up the request
	// we're getting statuses of admin
	targetAccount := suite.testAccounts["admin_account"]
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodGet, nil, fmt.Sprintf("/api/v1/accounts/%s/statuses?limit=20&only_media=true&only_public=true", targetAccount.ID), "")
	ctx.Params = gin.Params{
		gin.Param{
			Key:   account.IDKey,
			Value: targetAccount.ID,
		},
	}

	// call the handler
	suite.accountModule.AccountStatusesGETHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// unmarshal the returned statuses
	apimodelStatuses := []*apimodel.Status{}
	err = json.Unmarshal(b, &apimodelStatuses)
	suite.NoError(err)
	suite.NotEmpty(apimodelStatuses)

	for _, s := range apimodelStatuses {
		suite.NotEmpty(s.MediaAttachments)
		suite.Equal(apimodel.VisibilityPublic, s.Visibility)
	}
}

func TestAccountStatusesTestSuite(t *testing.T) {
	suite.Run(t, new(AccountStatusesTestSuite))
}
