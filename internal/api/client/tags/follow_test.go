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
	"context"
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/tags"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
)

func (suite *TagsTestSuite) follow(
	accountFixtureName string,
	tagName string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Tag, error) {
	return suite.tagAction(
		accountFixtureName,
		tagName,
		http.MethodPost,
		tags.FollowPath,
		suite.tagsModule.FollowTagPOSTHandler,
		expectedHTTPStatus,
		expectedBody,
	)
}

// Follow a tag we don't already follow.
func (suite *TagsTestSuite) TestFollow() {
	accountFixtureName := "local_account_2"
	testTag := suite.testTags["welcome"]

	apiTag, err := suite.follow(accountFixtureName, testTag.Name, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testTag.Name, apiTag.Name)
	if suite.NotNil(apiTag.Following) {
		suite.True(*apiTag.Following)
	}
}

// When we follow a tag already followed by the account, it should succeed.
func (suite *TagsTestSuite) TestFollowIdempotent() {
	accountFixtureName := "local_account_2"
	testAccount := suite.testAccounts[accountFixtureName]
	testTag := suite.testTags["welcome"]

	// Setup: follow an existing tag.
	if err := suite.db.PutFollowedTag(context.Background(), testAccount.ID, testTag.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Follow it again through the API.
	apiTag, err := suite.follow(accountFixtureName, testTag.Name, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testTag.Name, apiTag.Name)
	if suite.NotNil(apiTag.Following) {
		suite.True(*apiTag.Following)
	}
}
