package rollback

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
)

var Rollback action.GTSAction = func(ctx context.Context) (err error) {
	return bundb.DoRollback(ctx)
}
