package prune

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
)

// All performs all media clean actions
var All action.GTSAction = func(ctx context.Context) error {
	prune, err := setupPrune(ctx)
	if err != nil {
		return err
	}

	if config.GetAdminMediaPruneDryRun() {
		ctx = gtscontext.SetDryRun(ctx)
	}

	days := config.GetMediaRemoteCacheDays()
	prune.cleaner.Media().All(ctx, days)
	prune.cleaner.Emoji().All(ctx)

	return prune.shutdown(ctx)
}
