package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type HeaderFilterTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *HeaderFilterTestSuite) TestAllowHeaderFilterGetPutUpdateDelete() {
	suite.testHeaderFilterGetPutUpdateDelete(
		suite.db.GetAllowHeaderFilter,
		suite.db.GetAllowHeaderFilters,
		suite.db.PutAllowHeaderFilter,
		suite.db.DeleteAllowHeaderFilter,
	)
}

func (suite *HeaderFilterTestSuite) TestBlockHeaderFilterGetPutUpdateDelete() {
	suite.testHeaderFilterGetPutUpdateDelete(
		suite.db.GetBlockHeaderFilter,
		suite.db.GetBlockHeaderFilters,
		suite.db.PutBlockHeaderFilter,
		suite.db.DeleteBlockHeaderFilter,
	)
}

func (suite *HeaderFilterTestSuite) testHeaderFilterGetPutUpdateDelete(
	get func(context.Context, string) (*gtsmodel.HeaderFilter, error),
	getAll func(context.Context) ([]*gtsmodel.HeaderFilter, error),
	put func(context.Context, *gtsmodel.HeaderFilter) error,
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
	ctx := context.Background()
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
