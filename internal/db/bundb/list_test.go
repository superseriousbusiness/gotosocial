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

package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"golang.org/x/exp/slices"
)

type ListTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *ListTestSuite) testStructs() (*gtsmodel.List, *gtsmodel.Account) {
	testList := &gtsmodel.List{}
	*testList = *suite.testLists["local_account_1_list_1"]

	// Populate entries on this list as we'd expect them back from the db.
	entries := make([]*gtsmodel.ListEntry, 0, len(suite.testListEntries))
	for _, entry := range suite.testListEntries {
		entries = append(entries, entry)
	}

	// Sort by ID descending (again, as we'd expect from the db).
	slices.SortFunc(entries, func(a, b *gtsmodel.ListEntry) bool {
		return b.ID < a.ID
	})

	testList.ListEntries = entries

	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]

	return testList, testAccount
}

func (suite *ListTestSuite) checkList(expected *gtsmodel.List, actual *gtsmodel.List) {
	suite.Equal(expected.ID, actual.ID)
	suite.Equal(expected.Title, actual.Title)
	suite.Equal(expected.AccountID, actual.AccountID)
	suite.Equal(expected.RepliesPolicy, actual.RepliesPolicy)
	suite.NotNil(actual.Account)
}

func (suite *ListTestSuite) checkListEntry(expected *gtsmodel.ListEntry, actual *gtsmodel.ListEntry) {
	suite.Equal(expected.ID, actual.ID)
	suite.Equal(expected.ListID, actual.ListID)
	suite.Equal(expected.FollowID, actual.FollowID)
	suite.NotNil(actual.Follow)
}

func (suite *ListTestSuite) checkListEntries(expected []*gtsmodel.ListEntry, actual []*gtsmodel.ListEntry) {
	var (
		lExpected = len(expected)
		lActual   = len(actual)
	)

	if lExpected != lActual {
		suite.FailNow("", "expected %d list entries, got %d", lExpected, lActual)
	}

	var topID string
	for i, expectedEntry := range expected {
		actualEntry := actual[i]

		// Ensure ID descending.
		if topID == "" {
			topID = actualEntry.ID
		} else {
			suite.Less(actualEntry.ID, topID)
		}

		suite.checkListEntry(expectedEntry, actualEntry)
	}
}

func (suite *ListTestSuite) TestGetListByID() {
	testList, _ := suite.testStructs()

	dbList, err := suite.db.GetListByID(context.Background(), testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkList(testList, dbList)
	suite.checkListEntries(testList.ListEntries, dbList.ListEntries)
}

func (suite *ListTestSuite) TestGetListsForAccountID() {
	testList, testAccount := suite.testStructs()

	dbLists, err := suite.db.GetListsForAccountID(context.Background(), testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(dbLists); l != 1 {
		suite.FailNow("", "expected %d lists, got %d", 1, l)
	}

	suite.checkList(testList, dbLists[0])
	suite.checkListEntries(testList.ListEntries, dbLists[0].ListEntries)
}

func (suite *ListTestSuite) TestPutList() {
	ctx := context.Background()
	_, testAccount := suite.testStructs()

	testList := &gtsmodel.List{
		ID:        "01H0J2PMYM54618VCV8Y8QYAT4",
		Title:     "Test List!",
		AccountID: testAccount.ID,
	}

	if err := suite.db.PutList(ctx, testList); err != nil {
		suite.FailNow(err.Error())
	}

	dbList, err := suite.db.GetListByID(ctx, testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Bodge testlist as though default had been set.
	testList.RepliesPolicy = gtsmodel.RepliesPolicyFollowed
	suite.checkList(testList, dbList)
}

func (suite *ListTestSuite) TestUpdateList() {
	ctx := context.Background()
	testList, _ := suite.testStructs()

	// Get List in the cache first.
	dbList, err := suite.db.GetListByID(ctx, testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Now do the update.
	testList.Title = "New Title!"
	if err := suite.db.UpdateList(ctx, testList, "title"); err != nil {
		suite.FailNow(err.Error())
	}

	// Cache should be invalidated
	// + we should have updated list.
	dbList, err = suite.db.GetListByID(ctx, testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkList(testList, dbList)
}

func (suite *ListTestSuite) TestDeleteList() {
	ctx := context.Background()
	testList, _ := suite.testStructs()

	// Get List in the cache first.
	if _, err := suite.db.GetListByID(ctx, testList.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Now do the delete.
	if err := suite.db.DeleteListByID(ctx, testList.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Cache should be invalidated
	// + we should have no list.
	_, err := suite.db.GetListByID(ctx, testList.ID)
	suite.ErrorIs(err, db.ErrNoEntries)

	// All entries belonging to this
	// list should now be deleted.
	listEntries, err := suite.db.GetListEntries(ctx, testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Empty(listEntries)
}

func (suite *ListTestSuite) TestDeleteListEntry() {
	ctx := context.Background()
	testList, _ := suite.testStructs()

	// Get List in the cache first.
	if _, err := suite.db.GetListByID(ctx, testList.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Delete the first entry.
	if err := suite.db.DeleteListEntry(ctx, testList.ListEntries[0].ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Get list from the db again.
	dbList, err := suite.db.GetListByID(ctx, testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Bodge the testlist as though
	// we'd removed the first entry.
	testList.ListEntries = testList.ListEntries[1:]
	suite.checkList(testList, dbList)
}

func (suite *ListTestSuite) TestDeleteListEntriesForFollowID() {
	ctx := context.Background()
	testList, _ := suite.testStructs()

	// Get List in the cache first.
	if _, err := suite.db.GetListByID(ctx, testList.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Delete the first entry.
	if err := suite.db.DeleteListEntriesForFollowID(ctx, testList.ListEntries[0].FollowID); err != nil {
		suite.FailNow(err.Error())
	}

	// Get list from the db again.
	dbList, err := suite.db.GetListByID(ctx, testList.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Bodge the testlist as though
	// we'd removed the first entry.
	testList.ListEntries = testList.ListEntries[1:]
	suite.checkList(testList, dbList)
}

func TestListTestSuite(t *testing.T) {
	suite.Run(t, new(ListTestSuite))
}
