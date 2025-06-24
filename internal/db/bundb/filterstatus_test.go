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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// TestFilterStatusCRD tests CRD (no U) and read-all operations on filter statuses.
func (suite *FilterTestSuite) TestFilterStatusCRD() {
	t := suite.T()

	// Create new filter.
	filter := &gtsmodel.Filter{
		ID:        "01HNEJNVZZVXJTRB3FX3K2B1YF",
		AccountID: "01HNEJXCPRTJVJY9MV0VVHGD47",
		Title:     "foss jail",
		Action:    gtsmodel.FilterActionWarn,
		Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextHome | gtsmodel.FilterContextPublic),
	}

	// Create new cancellable test context.
	ctx := suite.T().Context()
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	// Insert the new filter into the DB.
	err := suite.db.PutFilter(ctx, filter)
	if err != nil {
		t.Fatalf("error inserting filter: %v", err)
	}

	// Add a filter status to it.
	filterStatus := &gtsmodel.FilterStatus{
		ID:       "01HNEK4RW5QEAMG9Y4ET6ST0J4",
		FilterID: filter.ID,
		StatusID: "01HQXGMQ3QFXRT4GX9WNQ8KC0X",
	}

	// Insert the new filter status into the DB.
	err = suite.db.PutFilterStatus(ctx, filterStatus)
	if err != nil {
		t.Fatalf("error inserting filter status: %v", err)
	}

	// Try to find it again and ensure it has the fields we expect.
	check, err := suite.db.GetFilterStatusByID(ctx, filterStatus.ID)
	if err != nil {
		t.Fatalf("error fetching filter status: %v", err)
	}
	suite.Equal(filterStatus.ID, check.ID)
	suite.Equal(filterStatus.FilterID, check.FilterID)
	suite.Equal(filterStatus.StatusID, check.StatusID)

	// Check that fetching multiple filter statuses by IDs works.
	checks, err := suite.db.GetFilterStatusesByIDs(ctx, []string{filterStatus.ID})
	if err != nil {
		t.Fatalf("error fetching filter statuses: %v", err)
	}
	suite.Len(checks, 1)
	suite.Equal(filterStatus.ID, checks[0].ID)

	// Delete the filter status from the DB.
	err = suite.db.DeleteFilterStatusesByIDs(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error deleting filter status: %v", err)
	}

	// Ensure we can't refetch it.
	check, err = suite.db.GetFilterStatusByID(ctx, filter.ID)
	if !errors.Is(err, db.ErrNoEntries) {
		t.Fatalf("fetching deleted filter status returned unexpected error: %v", err)
	}
	suite.Nil(check)

	// Ensure the filter itself is still there.
	checkFilter, err := suite.db.GetFilterByID(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching filter: %v", err)
	}
	suite.Equal(filter.ID, checkFilter.ID)
}
