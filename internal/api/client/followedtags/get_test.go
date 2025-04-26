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

package followedtags_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/followedtags"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

func (suite *FollowedTagsTestSuite) getFollowedTags(
	accountFixtureName string,
	expectedHTTPStatus int,
	expectedBody string,
) ([]apimodel.Tag, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts[accountFixtureName])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens[accountFixtureName]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers[accountFixtureName])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+followedtags.BasePath, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.followedTagsModule.FollowedTagsGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
		if expectedBody == "" {
			return nil, errs.Combine()
		}
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := []apimodel.Tag{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// Test that we can list a user's followed tags.
func (suite *FollowedTagsTestSuite) TestGet() {
	accountFixtureName := "local_account_2"
	testAccount := suite.testAccounts[accountFixtureName]
	testTag := suite.testTags["welcome"]

	// Follow an existing tag.
	if err := suite.db.PutFollowedTag(context.Background(), testAccount.ID, testTag.ID); err != nil {
		suite.FailNow(err.Error())
	}

	followedTags, err := suite.getFollowedTags(accountFixtureName, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	if suite.Len(followedTags, 1) {
		followedTag := followedTags[0]
		suite.Equal(testTag.Name, followedTag.Name)
		if suite.NotNil(followedTag.Following) {
			suite.True(*followedTag.Following)
		}
	}
}

// Test that we can list a user's followed tags even if they don't have any.
func (suite *FollowedTagsTestSuite) TestGetEmpty() {
	accountFixtureName := "local_account_1"

	followedTags, err := suite.getFollowedTags(accountFixtureName, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(followedTags, 0)
}
