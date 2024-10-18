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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type FilterTestSuite struct {
	BunDBStandardTestSuite
}

// TestFilterCRUD tests CRUD and read-all operations on filters.
func (suite *FilterTestSuite) TestFilterCRUD() {
	t := suite.T()

	// Create new example filter with attached keyword.
	filter := &gtsmodel.Filter{
		ID:            "01HNEJNVZZVXJTRB3FX3K2B1YF",
		AccountID:     "01HNEJXCPRTJVJY9MV0VVHGD47",
		Title:         "foss jail",
		Action:        gtsmodel.FilterActionWarn,
		ContextHome:   util.Ptr(true),
		ContextPublic: util.Ptr(true),
	}
	filterKeyword := &gtsmodel.FilterKeyword{
		ID:        "01HNEK4RW5QEAMG9Y4ET6ST0J4",
		AccountID: filter.AccountID,
		FilterID:  filter.ID,
		Keyword:   "GNU/Linux",
	}
	filter.Keywords = []*gtsmodel.FilterKeyword{filterKeyword}

	// Create new cancellable test context.
	ctx := context.Background()
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	// Insert the example filter into db.
	if err := suite.db.PutFilter(ctx, filter); err != nil {
		t.Fatalf("error inserting filter: %v", err)
	}

	// Now fetch newly created filter.
	check, err := suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching filter: %v", err)
	}

	// Check all expected fields match.
	suite.Equal(filter.ID, check.ID)
	suite.Equal(filter.AccountID, check.AccountID)
	suite.Equal(filter.Title, check.Title)
	suite.Equal(filter.Action, check.Action)
	suite.Equal(filter.ContextHome, check.ContextHome)
	suite.Equal(filter.ContextNotifications, check.ContextNotifications)
	suite.Equal(filter.ContextPublic, check.ContextPublic)
	suite.Equal(filter.ContextThread, check.ContextThread)
	suite.Equal(filter.ContextAccount, check.ContextAccount)
	suite.NotZero(check.CreatedAt)
	suite.NotZero(check.UpdatedAt)

	suite.Equal(len(filter.Keywords), len(check.Keywords))
	suite.Equal(filter.Keywords[0].ID, check.Keywords[0].ID)
	suite.Equal(filter.Keywords[0].AccountID, check.Keywords[0].AccountID)
	suite.Equal(filter.Keywords[0].FilterID, check.Keywords[0].FilterID)
	suite.Equal(filter.Keywords[0].Keyword, check.Keywords[0].Keyword)
	suite.Equal(filter.Keywords[0].FilterID, check.Keywords[0].FilterID)
	suite.NotZero(check.Keywords[0].CreatedAt)
	suite.NotZero(check.Keywords[0].UpdatedAt)

	suite.Equal(len(filter.Statuses), len(check.Statuses))

	// Fetch all filters.
	all, err := suite.db.GetFiltersForAccountID(ctx, filter.AccountID)
	if err != nil {
		t.Fatalf("error fetching filters: %v", err)
	}

	// Ensure the result contains our example filter.
	suite.Len(all, 1)
	suite.Equal(filter.ID, all[0].ID)

	suite.Len(all[0].Keywords, 1)
	suite.Equal(filter.Keywords[0].ID, all[0].Keywords[0].ID)

	suite.Empty(all[0].Statuses)

	// Update the filter context and add another keyword and a status.
	check.ContextNotifications = util.Ptr(true)

	newKeyword := &gtsmodel.FilterKeyword{
		ID:        "01HNEMY810E5XKWDDMN5ZRE749",
		FilterID:  filter.ID,
		AccountID: filter.AccountID,
		Keyword:   "tux",
	}
	check.Keywords = append(check.Keywords, newKeyword)

	newStatus := &gtsmodel.FilterStatus{
		ID:        "01HNEMYD5XE7C8HH8TNCZ76FN2",
		FilterID:  filter.ID,
		AccountID: filter.AccountID,
		StatusID:  "01HNEKZW34SQZ8PSDQ0Z10NZES",
	}
	check.Statuses = append(check.Statuses, newStatus)

	if err := suite.db.UpdateFilter(ctx, check, nil, [][]string{nil, nil}, nil, nil); err != nil {
		t.Fatalf("error updating filter: %v", err)
	}
	// Now fetch newly updated filter.
	check, err = suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching updated filter: %v", err)
	}

	// Ensure expected fields were modified on check filter.
	suite.True(check.UpdatedAt.After(filter.UpdatedAt))
	if suite.NotNil(check.ContextHome) {
		suite.True(*check.ContextHome)
	}
	if suite.NotNil(check.ContextNotifications) {
		suite.True(*check.ContextNotifications)
	}
	if suite.NotNil(check.ContextPublic) {
		suite.True(*check.ContextPublic)
	}
	if suite.NotNil(check.ContextThread) {
		suite.False(*check.ContextThread)
	}
	if suite.NotNil(check.ContextAccount) {
		suite.False(*check.ContextAccount)
	}

	// Ensure keyword entries were added.
	suite.Len(check.Keywords, 2)
	checkFilterKeywordIDs := make([]string, 0, 2)
	for _, checkFilterKeyword := range check.Keywords {
		checkFilterKeywordIDs = append(checkFilterKeywordIDs, checkFilterKeyword.ID)
	}
	suite.ElementsMatch([]string{filterKeyword.ID, newKeyword.ID}, checkFilterKeywordIDs)

	// Ensure status entry was added.
	suite.Len(check.Statuses, 1)
	checkFilterStatusIDs := make([]string, 0, 1)
	for _, checkFilterStatus := range check.Statuses {
		checkFilterStatusIDs = append(checkFilterStatusIDs, checkFilterStatus.ID)
	}
	suite.ElementsMatch([]string{newStatus.ID}, checkFilterStatusIDs)

	// Update one filter keyword and delete another. Don't change the filter or the filter status.
	filterKeyword.WholeWord = util.Ptr(true)
	check.Keywords = []*gtsmodel.FilterKeyword{filterKeyword}
	check.Statuses = nil

	if err := suite.db.UpdateFilter(ctx, check, nil, [][]string{{"whole_word"}}, []string{newKeyword.ID}, nil); err != nil {
		t.Fatalf("error updating filter: %v", err)
	}
	check, err = suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching updated filter: %v", err)
	}

	// Ensure expected fields were not modified.
	suite.Equal(filter.Title, check.Title)
	suite.Equal(gtsmodel.FilterActionWarn, check.Action)
	if suite.NotNil(check.ContextHome) {
		suite.True(*check.ContextHome)
	}
	if suite.NotNil(check.ContextNotifications) {
		suite.True(*check.ContextNotifications)
	}
	if suite.NotNil(check.ContextPublic) {
		suite.True(*check.ContextPublic)
	}
	if suite.NotNil(check.ContextThread) {
		suite.False(*check.ContextThread)
	}
	if suite.NotNil(check.ContextAccount) {
		suite.False(*check.ContextAccount)
	}

	// Ensure only changed field of keyword was modified, and other keyword was deleted.
	suite.Len(check.Keywords, 1)
	suite.Equal(filterKeyword.ID, check.Keywords[0].ID)
	suite.Equal("GNU/Linux", check.Keywords[0].Keyword)
	if suite.NotNil(check.Keywords[0].WholeWord) {
		suite.True(*check.Keywords[0].WholeWord)
	}

	// Ensure status entry was not deleted.
	suite.Len(check.Statuses, 1)
	suite.Equal(newStatus.ID, check.Statuses[0].ID)

	// Add another status entry for the same status ID. It should be ignored without problems.
	redundantStatus := &gtsmodel.FilterStatus{
		ID:        "01HQXJ5Y405XZSQ67C2BSQ6HJ0",
		FilterID:  filter.ID,
		AccountID: filter.AccountID,
		StatusID:  newStatus.StatusID,
	}
	check.Statuses = []*gtsmodel.FilterStatus{redundantStatus}
	if err := suite.db.UpdateFilter(ctx, check, nil, [][]string{nil}, nil, nil); err != nil {
		t.Fatalf("error updating filter: %v", err)
	}
	check, err = suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching updated filter: %v", err)
	}

	// Ensure status entry was not deleted, updated, or duplicated.
	suite.Len(check.Statuses, 1)
	suite.Equal(newStatus.ID, check.Statuses[0].ID)
	suite.Equal(newStatus.StatusID, check.Statuses[0].StatusID)

	// Now delete the filter from the DB.
	if err := suite.db.DeleteFilterByID(ctx, filter.ID); err != nil {
		t.Fatalf("error deleting filter: %v", err)
	}

	// Ensure we can't refetch it.
	_, err = suite.db.GetFilterByID(ctx, filter.ID)
	if !errors.Is(err, db.ErrNoEntries) {
		t.Fatalf("fetching deleted filter returned unexpected error: %v", err)
	}
}

func (suite *FilterTestSuite) TestFilterTitleOverlap() {
	var (
		ctx      = context.Background()
		account1 = "01HNEJXCPRTJVJY9MV0VVHGD47"
		account2 = "01JAG5BRJPJYA0FSA5HR2MMFJH"
	)

	// Create an empty filter for account 1.
	account1filter1 := &gtsmodel.Filter{
		ID:          "01HNEJNVZZVXJTRB3FX3K2B1YF",
		AccountID:   account1,
		Title:       "my filter",
		Action:      gtsmodel.FilterActionWarn,
		ContextHome: util.Ptr(true),
	}
	if err := suite.db.PutFilter(ctx, account1filter1); err != nil {
		suite.FailNow("", "error putting account1filter1: %s", err)
	}

	// Create a filter for account 2 with
	// the same title, should be no issue.
	account2filter1 := &gtsmodel.Filter{
		ID:          "01JAG5GPXG7H5Y4ZP78GV1F2ET",
		AccountID:   account2,
		Title:       "my filter",
		Action:      gtsmodel.FilterActionWarn,
		ContextHome: util.Ptr(true),
	}
	if err := suite.db.PutFilter(ctx, account2filter1); err != nil {
		suite.FailNow("", "error putting account2filter1: %s", err)
	}

	// Try to create another filter for
	// account 1 with the same name as
	// an existing filter of theirs.
	account1filter2 := &gtsmodel.Filter{
		ID:          "01JAG5J8NYKQE2KYCD28Y4P05V",
		AccountID:   account1,
		Title:       "my filter",
		Action:      gtsmodel.FilterActionWarn,
		ContextHome: util.Ptr(true),
	}
	err := suite.db.PutFilter(ctx, account1filter2)
	if !errors.Is(err, db.ErrAlreadyExists) {
		suite.FailNow("", "wanted ErrAlreadyExists, got %s", err)
	}
}

func TestFilterTestSuite(t *testing.T) {
	suite.Run(t, new(FilterTestSuite))
}
