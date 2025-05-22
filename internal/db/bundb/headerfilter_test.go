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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type HeaderFilterTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *HeaderFilterTestSuite) TestAllowHeaderFilterGetPutUpdateDelete() {
	suite.testHeaderFilterGetPutUpdateDelete(
		suite.db.GetAllowHeaderFilter,
		suite.db.GetAllowHeaderFilters,
		suite.db.PutAllowHeaderFilter,
		suite.db.UpdateAllowHeaderFilter,
		suite.db.DeleteAllowHeaderFilter,
	)
}

func (suite *HeaderFilterTestSuite) TestBlockHeaderFilterGetPutUpdateDelete() {
	suite.testHeaderFilterGetPutUpdateDelete(
		suite.db.GetBlockHeaderFilter,
		suite.db.GetBlockHeaderFilters,
		suite.db.PutBlockHeaderFilter,
		suite.db.UpdateBlockHeaderFilter,
		suite.db.DeleteBlockHeaderFilter,
	)
}

func (suite *HeaderFilterTestSuite) testHeaderFilterGetPutUpdateDelete(
	get func(context.Context, string) (*gtsmodel.HeaderFilter, error),
	getAll func(context.Context) ([]*gtsmodel.HeaderFilter, error),
	put func(context.Context, *gtsmodel.HeaderFilter) error,
	update func(context.Context, *gtsmodel.HeaderFilter, ...string) error,
	delete func(context.Context, string) error,
) {
	t := suite.T()

	// Create new example header filter.
	filter := gtsmodel.HeaderFilter{
		ID:       "some unique id",
		Header:   "Http-Header-Key",
		Regex:    ".*",
		AuthorID: "some unique author id",
	}

	// Create new cancellable test context.
	ctx := suite.T().Context()
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	// Insert the example header filter into db.
	if err := put(ctx, &filter); err != nil {
		t.Fatalf("error inserting header filter: %v", err)
	}

	// Now fetch newly created filter.
	check, err := get(ctx, filter.ID)
	if err != nil {
		t.Fatalf("error fetching header filter: %v", err)
	}

	// Check all expected fields match.
	suite.Equal(filter.ID, check.ID)
	suite.Equal(filter.Header, check.Header)
	suite.Equal(filter.Regex, check.Regex)
	suite.Equal(filter.AuthorID, check.AuthorID)

	// Fetch all header filters.
	all, err := getAll(ctx)
	if err != nil {
		t.Fatalf("error fetching header filters: %v", err)
	}

	// Ensure contains example.
	suite.Equal(len(all), 1)
	suite.Equal(all[0].ID, filter.ID)

	// Update the header filter regex value.
	check.Regex = "new regex value"
	if err := update(ctx, check); err != nil {
		t.Fatalf("error updating header filter: %v", err)
	}

	// Ensure 'updated_at' was updated on check model.
	suite.True(check.UpdatedAt.After(filter.UpdatedAt))

	// Now delete the header filter from db.
	if err := delete(ctx, filter.ID); err != nil {
		t.Fatalf("error deleting header filter: %v", err)
	}

	// Ensure we can't refetch it.
	_, err = get(ctx, filter.ID)
	if err != db.ErrNoEntries {
		t.Fatalf("deleted header filter returned unexpected error: %v", err)
	}
}

func TestHeaderFilterTestSuite(t *testing.T) {
	suite.Run(t, new(HeaderFilterTestSuite))
}
