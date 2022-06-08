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

package status_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

type StatusBoostTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusBoostTestSuite) TestPostBoost() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	targetStatus := suite.testStatuses["admin_account_status_1"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.ReblogPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusBoostPOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), model.VisibilityPublic, statusReply.Visibility)

	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.SpoilerText)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.Equal(suite.T(), "the_mighty_zork", statusReply.Account.Username)
	assert.Len(suite.T(), statusReply.MediaAttachments, 0)
	assert.Len(suite.T(), statusReply.Mentions, 0)
	assert.Len(suite.T(), statusReply.Emojis, 0)
	assert.Len(suite.T(), statusReply.Tags, 0)

	assert.NotNil(suite.T(), statusReply.Application)
	assert.Equal(suite.T(), "really cool gts application", statusReply.Application.Name)

	assert.NotNil(suite.T(), statusReply.Reblog)
	assert.Equal(suite.T(), 1, statusReply.Reblog.ReblogsCount)
	assert.Equal(suite.T(), 1, statusReply.Reblog.FavouritesCount)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Reblog.Content)
	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.Reblog.SpoilerText)
	assert.Equal(suite.T(), targetStatus.AccountID, statusReply.Reblog.Account.ID)
	assert.Len(suite.T(), statusReply.Reblog.MediaAttachments, 1)
	assert.Len(suite.T(), statusReply.Reblog.Tags, 1)
	assert.Len(suite.T(), statusReply.Reblog.Emojis, 1)
	assert.Equal(suite.T(), "superseriousbusiness", statusReply.Reblog.Application.Name)
}

// try to boost a status that's not boostable
func (suite *StatusBoostTestSuite) TestPostUnboostable() {

	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	targetStatus := suite.testStatuses["local_account_2_status_4"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.ReblogPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusBoostPOSTHandler(ctx)

	// check response
	suite.Equal(http.StatusForbidden, recorder.Code) // we 403 unboostable statuses

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{"error":"Forbidden"}`, string(b))
}

// try to boost a status that's not visible to the user
func (suite *StatusBoostTestSuite) TestPostNotVisible() {

	// stop local_account_2 following zork
	err := suite.db.DeleteByID(context.Background(), suite.testFollows["local_account_2_local_account_1"].ID, &gtsmodel.Follow{})
	suite.NoError(err)

	t := suite.testTokens["local_account_2"]
	oauthToken := oauth.DBTokenToToken(t)

	targetStatus := suite.testStatuses["local_account_1_status_3"] // this is a mutual only status and these accounts aren't mutuals

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_2"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_2"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(status.ReblogPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   status.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusBoostPOSTHandler(ctx)

	// check response
	suite.Equal(http.StatusNotFound, recorder.Code) // we 404 statuses that aren't visible
}

func TestStatusBoostTestSuite(t *testing.T) {
	suite.Run(t, new(StatusBoostTestSuite))
}
