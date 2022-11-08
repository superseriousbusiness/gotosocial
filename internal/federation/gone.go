package federation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// CheckGone checks if a tombstone exists in the database for AP Actor or Object with the given uri.
func (f *federator) CheckGone(ctx context.Context, uri *url.URL) (bool, error) {
	return f.db.TombstoneExistsWithURI(ctx, uri.String())
}

// HandleGone puts a tombstone in the database, which marks an AP Actor or Object with the given uri as gone.
func (f *federator) HandleGone(ctx context.Context, uri *url.URL) error {
	tombstoneID, err := id.NewULID()
	if err != nil {
		err = fmt.Errorf("HandleGone: error generating id for new tombstone %s: %s", uri, err)
		log.Error(err)
		return err
	}

	tombstone := &gtsmodel.Tombstone{
		ID:     tombstoneID,
		Domain: uri.Host,
		URI:    uri.String(),
	}

	return f.db.PutTombstone(ctx, tombstone)
}
