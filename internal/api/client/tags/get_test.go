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

	"github.com/superseriousbusiness/gotosocial/internal/api/client/tags"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

// tagAction follows or unfollows a tag.
func (suite *TagsTestSuite) get(
	accountFixtureName string,
	tagName string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Tag, error) {
	return suite.tagAction(
		accountFixtureName,
		tagName,
		http.MethodGet,
		tags.TagPath,
		suite.tagsModule.TagGETHandler,
		expectedHTTPStatus,
		expectedBody,
	)
}

// Get a tag followed by the account.
func (suite *TagsTestSuite) TestGetFollowed() {
	accountFixtureName := "local_account_2"
	testAccount := suite.testAccounts[accountFixtureName]
	testTag := suite.testTags["welcome"]

	// Setup: follow an existing tag.
	if err := suite.db.PutFollowedTag(context.Background(), testAccount.ID, testTag.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Get it through the API.
	apiTag, err := suite.get(accountFixtureName, testTag.Name, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testTag.Name, apiTag.Name)
	if suite.NotNil(apiTag.Following) {
		suite.True(*apiTag.Following)
	}
}

// Get a tag not followed by the account.
func (suite *TagsTestSuite) TestGetUnfollowed() {
	accountFixtureName := "local_account_2"
	testTag := suite.testTags["Hashtag"]

	apiTag, err := suite.get(accountFixtureName, testTag.Name, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testTag.Name, apiTag.Name)
	if suite.NotNil(apiTag.Following) {
		suite.False(*apiTag.Following)
	}
}

// Get a tag that does not exist, which should result in a 404.
func (suite *TagsTestSuite) TestGetNotFound() {
	accountFixtureName := "local_account_2"

	_, err := suite.get(accountFixtureName, "THIS_TAG_DOES_NOT_EXIST", http.StatusNotFound, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
}
