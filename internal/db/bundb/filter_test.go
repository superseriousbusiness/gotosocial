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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type FilterTestSuite struct {
	BunDBStandardTestSuite
}

// TestFilterCRUD tests CRUD and read-all operations on filters.
func (suite *FilterTestSuite) TestFilterCRUD() {
	t := suite.T()

	// Create new example filter with attached keyword.
	filter := &gtsmodel.Filter{
		ID:        "01HNEJNVZZVXJTRB3FX3K2B1YF",
		AccountID: "01HNEJXCPRTJVJY9MV0VVHGD47",
		Title:     "foss jail",
		Action:    gtsmodel.FilterActionWarn,
		Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextHome | gtsmodel.FilterContextPublic),
	}
	filterKeyword := &gtsmodel.FilterKeyword{
		ID:       "01HNEK4RW5QEAMG9Y4ET6ST0J4",
		FilterID: filter.ID,
		Keyword:  "GNU/Linux",
	}
	filter.Keywords = []*gtsmodel.FilterKeyword{filterKeyword}
	filter.KeywordIDs = []string{filterKeyword.ID}

	// Create new cancellable test context.
	ctx := suite.T().Context()
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	// Insert the example filter keyword into db.
	if err := suite.db.PutFilterKeyword(ctx, filterKeyword); err != nil {
		t.Fatalf("error inserting filter keyword: %v", err)
	}

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
	suite.Equal(filter.Contexts, check.Contexts)

	suite.Equal(len(filter.Keywords), len(check.Keywords))
	suite.Equal(filter.Keywords[0].ID, check.Keywords[0].ID)
	suite.Equal(filter.Keywords[0].FilterID, check.Keywords[0].FilterID)
	suite.Equal(filter.Keywords[0].Keyword, check.Keywords[0].Keyword)
	suite.Equal(filter.Keywords[0].FilterID, check.Keywords[0].FilterID)

	suite.Equal(len(filter.Statuses), len(check.Statuses))

	// Fetch all filters.
	all, err := suite.db.GetFiltersByAccountID(ctx, filter.AccountID)
	if err != nil {
		t.Fatalf("error fetching filters: %v", err)
	}

	// Ensure the result contains our example filter.
	suite.Len(all, 1)
	suite.Equal(filter.ID, all[0].ID)

	suite.Len(all[0].Keywords, 1)
	suite.Equal(filter.Keywords[0].ID, all[0].Keywords[0].ID)

	suite.Empty(all[0].Statuses)

	// Update the filter context and
	// add another keyword and a status.
	check.Contexts.SetNotifications()
	newKeyword := &gtsmodel.FilterKeyword{
		ID:       "01HNEMY810E5XKWDDMN5ZRE749",
		FilterID: filter.ID,
		Keyword:  "tux",
	}
	check.Keywords = append(check.Keywords, newKeyword)
	check.KeywordIDs = append(check.KeywordIDs, newKeyword.ID)
	newStatus := &gtsmodel.FilterStatus{
		ID:       "01HNEMYD5XE7C8HH8TNCZ76FN2",
		FilterID: filter.ID,
		StatusID: "01HNEKZW34SQZ8PSDQ0Z10NZES",
	}
	check.Statuses = append(check.Statuses, newStatus)
	check.StatusIDs = append(check.StatusIDs, newStatus.ID)

	// Insert the new filter keyword.
	if err := suite.db.PutFilterKeyword(ctx, newKeyword); err != nil {
		t.Fatalf("error inserting filter keyword: %v", err)
	}

	// Insert the new filter status.
	if err := suite.db.PutFilterStatus(ctx, newStatus); err != nil {
		t.Fatalf("error inserting filter status: %v", err)
	}

	// Now update the filter with new keyword and status.
	if err := suite.db.UpdateFilter(ctx, check); err != nil {
		t.Fatalf("error updating filter: %v", err)
	}

	// Now fetch newly updated filter.
	check, err = suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching updated filter: %v", err)
	}

	// Ensure expected fields were modified on check filter.
	suite.True(check.Contexts.Home())
	suite.True(check.Contexts.Notifications())
	suite.True(check.Contexts.Public())
	suite.False(check.Contexts.Thread())
	suite.False(check.Contexts.Account())

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

	// Update the original filter keyword.
	filterKeyword.WholeWord = util.Ptr(true)
	if err := suite.db.UpdateFilterKeyword(ctx, filterKeyword); err != nil {
		t.Fatalf("error updating filter keyword: %v", err)
	}

	// Drop most recently added filter keyword from filter.
	check.Keywords = []*gtsmodel.FilterKeyword{filterKeyword}
	check.KeywordIDs = []string{filterKeyword.ID}
	if err := suite.db.UpdateFilter(ctx, check); err != nil {
		t.Fatalf("error updating filter: %v", err)
	}

	check, err = suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching updated filter: %v", err)
	}

	// Ensure expected fields were not modified.
	suite.Equal(filter.Title, check.Title)
	suite.Equal(gtsmodel.FilterActionWarn, check.Action)
	suite.True(check.Contexts.Home())
	suite.True(check.Contexts.Notifications())
	suite.True(check.Contexts.Public())
	suite.False(check.Contexts.Thread())
	suite.False(check.Contexts.Account())

	// Ensure only changed field of keyword was
	// modified, and other keyword was deleted.
	suite.Len(check.Keywords, 1)
	suite.Equal(filterKeyword.ID, check.Keywords[0].ID)
	suite.Equal("GNU/Linux", check.Keywords[0].Keyword)
	if suite.NotNil(check.Keywords[0].WholeWord) {
		suite.True(*check.Keywords[0].WholeWord)
	}

	// Ensure status entry was not deleted.
	suite.Len(check.Statuses, 1)
	suite.Equal(newStatus.ID, check.Statuses[0].ID)

	// Now delete the filter from the DB.
	if err := suite.db.DeleteFilter(ctx, filter); err != nil {
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
		ctx      = suite.T().Context()
		account1 = "01HNEJXCPRTJVJY9MV0VVHGD47"
		account2 = "01JAG5BRJPJYA0FSA5HR2MMFJH"
	)

	// Create an empty filter for account 1.
	account1filter1 := &gtsmodel.Filter{
		ID:        "01HNEJNVZZVXJTRB3FX3K2B1YF",
		AccountID: account1,
		Title:     "my filter",
		Action:    gtsmodel.FilterActionWarn,
		Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextHome),
	}
	if err := suite.db.PutFilter(ctx, account1filter1); err != nil {
		suite.FailNow("", "error putting account1filter1: %s", err)
	}

	// Create a filter for account 2 with
	// the same title, should be no issue.
	account2filter1 := &gtsmodel.Filter{
		ID:        "01JAG5GPXG7H5Y4ZP78GV1F2ET",
		AccountID: account2,
		Title:     "my filter",
		Action:    gtsmodel.FilterActionWarn,
		Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextHome),
	}
	if err := suite.db.PutFilter(ctx, account2filter1); err != nil {
		suite.FailNow("", "error putting account2filter1: %s", err)
	}

	// Try to create another filter for
	// account 1 with the same name as
	// an existing filter of theirs.
	account1filter2 := &gtsmodel.Filter{
		ID:        "01JAG5J8NYKQE2KYCD28Y4P05V",
		AccountID: account1,
		Title:     "my filter",
		Action:    gtsmodel.FilterActionWarn,
		Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextHome),
	}
	err := suite.db.PutFilter(ctx, account1filter2)
	if !errors.Is(err, db.ErrAlreadyExists) {
		suite.FailNow("", "wanted ErrAlreadyExists, got %s", err)
	}
}

func TestFilterTestSuite(t *testing.T) {
	suite.Run(t, new(FilterTestSuite))
}
