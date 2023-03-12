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

package bundb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type StatusBookmarkTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *StatusBookmarkTestSuite) TestGetStatusBookmarkIDOK() {
	testBookmark := suite.testBookmarks["local_account_1_admin_account_status_1"]

	id, err := suite.db.GetStatusBookmarkID(context.Background(), testBookmark.AccountID, testBookmark.StatusID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(testBookmark.ID, id)
}

func (suite *StatusBookmarkTestSuite) TestGetStatusBookmarkIDNonexisting() {
	id, err := suite.db.GetStatusBookmarkID(context.Background(), "01GVAVGD06YJ2FSB5GJSMF8M2K", "01GVAVGKGR1MK9ZN7JCJFYSFZV")
	suite.Empty(id)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *StatusBookmarkTestSuite) TestDeleteStatusBookmarksOriginatingFromAccount() {
	testAccount := suite.testAccounts["local_account_1"]

	if err := suite.db.DeleteStatusBookmarks(context.Background(), "", testAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	bookmarks := []*gtsmodel.StatusBookmark{}
	if err := suite.db.GetAll(context.Background(), &bookmarks); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, b := range bookmarks {
		if b.AccountID == testAccount.ID {
			suite.FailNowf("", "no StatusBookmarks with account id %s should remain", testAccount.ID)
		}
	}
}

func (suite *StatusBookmarkTestSuite) TestDeleteStatusBookmarksTargetingAccount() {
	testAccount := suite.testAccounts["local_account_1"]

	if err := suite.db.DeleteStatusBookmarks(context.Background(), testAccount.ID, ""); err != nil {
		suite.FailNow(err.Error())
	}

	bookmarks := []*gtsmodel.StatusBookmark{}
	if err := suite.db.GetAll(context.Background(), &bookmarks); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, b := range bookmarks {
		if b.TargetAccountID == testAccount.ID {
			suite.FailNowf("", "no StatusBookmarks with target account id %s should remain", testAccount.ID)
		}
	}
}

func (suite *StatusBookmarkTestSuite) TestDeleteStatusBookmarksTargetingStatus() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	if err := suite.db.DeleteStatusBookmarksForStatus(context.Background(), testStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	bookmarks := []*gtsmodel.StatusBookmark{}
	if err := suite.db.GetAll(context.Background(), &bookmarks); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, b := range bookmarks {
		if b.StatusID == testStatus.ID {
			suite.FailNowf("", "no StatusBookmarks with status id %s should remain", testStatus.ID)
		}
	}
}

func (suite *StatusBookmarkTestSuite) TestDeleteStatusBookmark() {
	testBookmark := suite.testBookmarks["local_account_1_admin_account_status_1"]
	ctx := context.Background()

	if err := suite.db.DeleteStatusBookmark(ctx, testBookmark.ID); err != nil {
		suite.FailNow(err.Error())
	}

	bookmark, err := suite.db.GetStatusBookmark(ctx, testBookmark.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(bookmark)
}

func (suite *StatusBookmarkTestSuite) TestDeleteStatusBookmarkNonExisting() {
	err := suite.db.DeleteStatusBookmark(context.Background(), "01GVAV715K6Y2SG9ZKS9ZA8G7G")
	suite.NoError(err)
}

func TestStatusBookmarkTestSuite(t *testing.T) {
	suite.Run(t, new(StatusBookmarkTestSuite))
}
