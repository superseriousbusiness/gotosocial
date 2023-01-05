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

package statuses_test

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
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
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
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.ReblogPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   statuses.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusBoostPOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &apimodel.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.False(statusReply.Sensitive)
	suite.Equal(apimodel.VisibilityPublic, statusReply.Visibility)

	suite.Equal(targetStatus.ContentWarning, statusReply.SpoilerText)
	suite.Equal(targetStatus.Content, statusReply.Content)
	suite.Equal("the_mighty_zork", statusReply.Account.Username)
	suite.Len(statusReply.MediaAttachments, 0)
	suite.Len(statusReply.Mentions, 0)
	suite.Len(statusReply.Emojis, 0)
	suite.Len(statusReply.Tags, 0)

	suite.NotNil(statusReply.Application)
	suite.Equal("really cool gts application", statusReply.Application.Name)

	suite.NotNil(statusReply.Reblog)
	suite.Equal(1, statusReply.Reblog.ReblogsCount)
	suite.Equal(1, statusReply.Reblog.FavouritesCount)
	suite.Equal(targetStatus.Content, statusReply.Reblog.Content)
	suite.Equal(targetStatus.ContentWarning, statusReply.Reblog.SpoilerText)
	suite.Equal(targetStatus.AccountID, statusReply.Reblog.Account.ID)
	suite.Len(statusReply.Reblog.MediaAttachments, 1)
	suite.Len(statusReply.Reblog.Tags, 1)
	suite.Len(statusReply.Reblog.Emojis, 1)
	suite.Equal("superseriousbusiness", statusReply.Reblog.Application.Name)
}

func (suite *StatusBoostTestSuite) TestPostBoostOwnFollowersOnly() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	testStatus := suite.testStatuses["local_account_1_status_5"]
	testAccount := suite.testAccounts["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, testUser)
	ctx.Set(oauth.SessionAuthorizedAccount, testAccount)
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.ReblogPath, ":id", testStatus.ID, 1)), nil)
	ctx.Request.Header.Set("accept", "application/json")

	ctx.Params = gin.Params{
		gin.Param{
			Key:   statuses.IDKey,
			Value: testStatus.ID,
		},
	}

	suite.statusModule.StatusBoostPOSTHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	responseStatus := &apimodel.Status{}
	err = json.Unmarshal(b, responseStatus)
	suite.NoError(err)

	suite.False(responseStatus.Sensitive)
	suite.Equal(suite.tc.VisToAPIVis(context.Background(), testStatus.Visibility), responseStatus.Visibility)

	suite.Equal(testStatus.ContentWarning, responseStatus.SpoilerText)
	suite.Equal(testStatus.Content, responseStatus.Content)
	suite.Equal("the_mighty_zork", responseStatus.Account.Username)
	suite.Len(responseStatus.MediaAttachments, 0)
	suite.Len(responseStatus.Mentions, 0)
	suite.Len(responseStatus.Emojis, 0)
	suite.Len(responseStatus.Tags, 0)

	suite.NotNil(responseStatus.Application)
	suite.Equal("really cool gts application", responseStatus.Application.Name)

	suite.NotNil(responseStatus.Reblog)
	suite.Equal(1, responseStatus.Reblog.ReblogsCount)
	suite.Equal(0, responseStatus.Reblog.FavouritesCount)
	suite.Equal(testStatus.Content, responseStatus.Reblog.Content)
	suite.Equal(testStatus.ContentWarning, responseStatus.Reblog.SpoilerText)
	suite.Equal(testStatus.AccountID, responseStatus.Reblog.Account.ID)
	suite.Equal(suite.tc.VisToAPIVis(context.Background(), testStatus.Visibility), responseStatus.Reblog.Visibility)
	suite.Empty(responseStatus.Reblog.MediaAttachments)
	suite.Empty(responseStatus.Reblog.Tags)
	suite.Empty(responseStatus.Reblog.Emojis)
	suite.Equal("really cool gts application", responseStatus.Reblog.Application.Name)
}

// try to boost a status that's not boostable
func (suite *StatusBoostTestSuite) TestPostUnboostable() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	targetStatus := suite.testStatuses["local_account_2_status_4"]

	// setup
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.ReblogPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   statuses.IDKey,
			Value: targetStatus.ID,
		},
	}

	suite.statusModule.StatusBoostPOSTHandler(ctx)

	// check response
	suite.Equal(http.StatusForbidden, recorder.Code) // we 403 unboostable statuses

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Equal(`{"error":"Forbidden"}`, string(b))
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
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_2"])
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_2"])
	ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.ReblogPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   statuses.IDKey,
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
