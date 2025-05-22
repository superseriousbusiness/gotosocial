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

package tags_test

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/tags"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
)

func (suite *TagsTestSuite) unfollow(
	accountFixtureName string,
	tagName string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Tag, error) {
	return suite.tagAction(
		accountFixtureName,
		tagName,
		http.MethodPost,
		tags.UnfollowPath,
		suite.tagsModule.UnfollowTagPOSTHandler,
		expectedHTTPStatus,
		expectedBody,
	)
}

// Unfollow a tag that we follow.
func (suite *TagsTestSuite) TestUnfollow() {
	accountFixtureName := "local_account_2"
	testAccount := suite.testAccounts[accountFixtureName]
	testTag := suite.testTags["welcome"]

	// Setup: follow an existing tag.
	if err := suite.db.PutFollowedTag(suite.T().Context(), testAccount.ID, testTag.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Unfollow it through the API.
	apiTag, err := suite.unfollow(accountFixtureName, testTag.Name, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testTag.Name, apiTag.Name)
	if suite.NotNil(apiTag.Following) {
		suite.False(*apiTag.Following)
	}
}

// When we unfollow a tag not followed by the account, it should succeed.
func (suite *TagsTestSuite) TestUnfollowIdempotent() {
	accountFixtureName := "local_account_2"
	testTag := suite.testTags["Hashtag"]

	apiTag, err := suite.unfollow(accountFixtureName, testTag.Name, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testTag.Name, apiTag.Name)
	if suite.NotNil(apiTag.Following) {
		suite.False(*apiTag.Following)
	}
}
