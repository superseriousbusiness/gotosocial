package cleaner

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Emoji encompasses a set of
// emoji cleanup / admin utils.
type Emoji struct {
	*Cleaner
}

// All will execute all cleaner.Emoji utilities synchronously, including output logging.
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (e *Emoji) All(ctx context.Context) {
	e.LogPruneMissing(ctx)
	e.LogFixBroken(ctx)
}

// LogPruneMissing performs emoji.PruneMissing(...), logging the start and outcome.
func (e *Emoji) LogPruneMissing(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := e.PruneMissing(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// LogFixBroken performs emoji.FixBroken(...), logging the start and outcome.
func (e *Emoji) LogFixBroken(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := e.FixBroken(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "fixed: %d", n)
	}
}

// PruneMissing will delete emoji with missing files from the database and storage driver.
// Context will be checked for `gtscontext.DryRun()` to perform the action. NOTE: this function
// should be updated to match media.FixCacheStat() if we ever support emoji uncaching.
func (e *Emoji) PruneMissing(ctx context.Context) (int, error) {
	var (
		total int
		maxID string
	)

	for {
		// Fetch the next batch of emoji media up to next ID.
		emojis, err := e.state.DB.GetEmojis(ctx, maxID, selectLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting emojis: %w", err)
		}

		if len(emojis) == 0 {
			// reached end.
			break
		}

		// Use last as the next 'maxID' value.
		maxID = emojis[len(emojis)-1].ID

		for _, emoji := range emojis {
			// Check / fix missing emoji media.
			fixed, err := e.pruneMissing(ctx, emoji)
			if err != nil {
				return total, err
			}

			if fixed {
				// Update
				// count.
				total++
			}
		}
	}

	return total, nil
}

// FixBroken will check all emojis for valid related models (e.g. category).
// Broken media will be automatically updated to remove now-missing models.
// Context will be checked for `gtscontext.DryRun()` to perform the action.
func (e *Emoji) FixBroken(ctx context.Context) (int, error) {
	var (
		total int
		maxID string
	)

	for {
		// Fetch the next batch of emoji media up to next ID.
		emojis, err := e.state.DB.GetEmojis(ctx, maxID, selectLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting emojis: %w", err)
		}

		if len(emojis) == 0 {
			// reached end.
			break
		}

		// Use last as the next 'maxID' value.
		maxID = emojis[len(emojis)-1].ID

		for _, emoji := range emojis {
			// Check / fix missing broken emoji.
			fixed, err := e.fixBroken(ctx, emoji)
			if err != nil {
				return total, err
			}

			if fixed {
				// Update
				// count.
				total++
			}
		}
	}

	return total, nil
}

func (e *Emoji) pruneMissing(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	return e.checkFiles(ctx, func() error {
		// Emoji missing files, delete it.
		// NOTE: if we ever support uncaching
		// of emojis, change to e.uncache().
		// In that case we should also rename
		// this function to match the media
		// equivalent -> fixCacheState().
		return e.delete(ctx, emoji)
	},
		emoji.ImageStaticPath,
		emoji.ImagePath,
	)
}

func (e *Emoji) fixBroken(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	// Check we have the required category for emoji.
	_, missing, err := e.getRelatedCategory(ctx, emoji)
	if err != nil {
		return false, err
	}

	if missing {
		if !gtscontext.DryRun(ctx) {
			// Dry run, do nothing.
			return true, nil
		}

		// Remove related category.
		emoji.CategoryID = ""

		// Update emoji model in the database to remove category ID.
		if err := e.state.DB.UpdateEmoji(ctx, emoji, "category_id"); err != nil {
			return true, gtserror.Newf("error updating emoji: %w", err)
		}

		return true, nil
	}

	return false, nil
}

func (e *Emoji) getRelatedCategory(ctx context.Context, emoji *gtsmodel.Emoji) (*gtsmodel.EmojiCategory, bool, error) {
	if emoji.CategoryID == "" {
		// no related category.
		return nil, false, nil
	}

	// Load the category related to this emoji.
	category, err := e.state.DB.GetEmojiCategory(
		gtscontext.SetBarebones(ctx),
		emoji.CategoryID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, false, gtserror.Newf("error fetching category by id %s: %w", emoji.CategoryID, err)
	}

	if category == nil {
		// Category is missing.
		return nil, true, nil
	}

	return category, false, nil
}

func (e *Emoji) delete(ctx context.Context, emoji *gtsmodel.Emoji) error {
	if gtscontext.DryRun(ctx) {
		// Dry run, do nothing.
		return nil
	}

	// Remove emoji and static files.
	_, err := e.removeFiles(ctx,
		emoji.ImageStaticPath,
		emoji.ImagePath,
	)
	if err != nil {
		return gtserror.Newf("error removing emoji files: %w", err)
	}

	// Delete emoji entirely from the database by its ID.
	if err := e.state.DB.DeleteEmojiByID(ctx, emoji.ID); err != nil {
		return gtserror.Newf("error deleting emoji: %w", err)
	}

	return nil
}
