/*
   GoToSocial
   Copyright (C) GoToSocial Authors admin@gotosocial.org
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

package statuses_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type StatusDeleteTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusDeleteTestSuite) TestPostDelete() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.BasePathWithID, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   statuses.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusDELETEHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &apimodel.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)
	suite.NotNil(statusReply)

	// Check that text and content type are returned for delete and redraft
	suite.Equal("hello everyone!", statusReply.Text)
	suite.Equal(apimodel.StatusContentTypePlain, statusReply.ContentType)

	if !testrig.WaitFor(func() bool {
		_, err := suite.db.GetStatusByID(ctx, targetStatus.ID)
		return errors.Is(err, db.ErrNoEntries)
	}) {
		suite.FailNow("time out waiting for status to be deleted")
	}

}

func TestStatusDeleteTestSuite(t *testing.T) {
	suite.Run(t, new(StatusDeleteTestSuite))
}
