package federation

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// CheckGone checks if a tombstone exists in the database for AP Actor or Object with the given uri.
func (f *federator) CheckGone(ctx context.Context, uri *url.URL) (bool, error) {
	return f.db.TombstoneExistsWithURI(ctx, uri.String())
}

// HandleGone puts a tombstone in the database, which marks an AP Actor or Object with the given uri as gone.
func (f *federator) HandleGone(ctx context.Context, uri *url.URL) error {
	tombstone := &gtsmodel.Tombstone{
		ID:     id.NewULID(),
		Domain: uri.Host,
		URI:    uri.String(),
	}
	return f.db.PutTombstone(ctx, tombstone)
}
