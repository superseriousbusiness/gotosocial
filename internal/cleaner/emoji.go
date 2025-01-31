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

package cleaner

import (
	"context"
	"errors"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// Emoji encompasses a set of
// emoji cleanup / admin utils.
type Emoji struct{ *Cleaner }

// All will execute all cleaner.Emoji utilities synchronously, including output logging.
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (e *Emoji) All(ctx context.Context, maxRemoteDays int) {
	t := time.Now().Add(-24 * time.Hour * time.Duration(maxRemoteDays))
	e.LogUncacheRemote(ctx, t)
	e.LogFixBroken(ctx)
	e.LogPruneUnused(ctx)
	e.LogFixCacheStates(ctx)
	_ = e.state.Storage.Storage.Clean(ctx)
}

// LogUncacheRemote performs Emoji.UncacheRemote(...), logging the start and outcome.
func (e *Emoji) LogUncacheRemote(ctx context.Context, olderThan time.Time) {
	log.Infof(ctx, "start older than: %s", olderThan.Format(time.Stamp))
	if n, err := e.UncacheRemote(ctx, olderThan); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "uncached: %d", n)
	}
}

// LogFixBroken performs Emoji.FixBroken(...), logging the start and outcome.
func (e *Emoji) LogFixBroken(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := e.FixBroken(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "fixed: %d", n)
	}
}

// LogPruneUnused performs Emoji.PruneUnused(...), logging the start and outcome.
func (e *Emoji) LogPruneUnused(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := e.PruneUnused(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// LogFixCacheStates performs Emoji.FixCacheStates(...), logging the start and outcome.
func (e *Emoji) LogFixCacheStates(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := e.FixCacheStates(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "fixed: %d", n)
	}
}

// UncacheRemote will uncache all remote emoji older than given input time. Context
// will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (e *Emoji) UncacheRemote(ctx context.Context, olderThan time.Time) (int, error) {
	var total int

	// Drop time by a minute to improve search,
	// (i.e. make it olderThan inclusive search).
	olderThan = olderThan.Add(-time.Minute)

	// Store recent time.
	mostRecent := olderThan

	for {
		// Fetch the next batch of cached emojis older than last-set time.
		emojis, err := e.state.DB.GetCachedEmojisOlderThan(ctx, olderThan, selectLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting remote emoji: %w", err)
		}

		// If no emojis / same group is
		// returned, we reached the end.
		if len(emojis) == 0 ||
			olderThan.Equal(emojis[len(emojis)-1].CreatedAt) {
			break
		}

		// Use last createdAt as next 'olderThan' value.
		olderThan = emojis[len(emojis)-1].CreatedAt

		for _, emoji := range emojis {
			// Check / uncache each remote emoji.
			uncached, err := e.uncacheRemote(ctx,
				mostRecent,
				emoji,
			)
			if err != nil {
				return total, err
			}

			if uncached {
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
		page  paging.Page
	)

	// Set page select limit.
	page.Limit = selectLimit

	for {
		// Fetch the next batch of emoji to next max ID.
		emojis, err := e.state.DB.GetEmojis(ctx, &page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting emojis: %w", err)
		}

		// Get current max ID.
		maxID := page.Max.Value

		// If no emoji or the same group is returned, we reached end.
		if len(emojis) == 0 || maxID == emojis[len(emojis)-1].ID {
			break
		}

		// Use last ID as the next 'maxID'.
		maxID = emojis[len(emojis)-1].ID
		page.Max = paging.MaxID(maxID)

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

// PruneUnused will delete all unused emoji media from the database and storage driver.
// Context will be checked for `gtscontext.DryRun()` to perform the action. NOTE: this function
// should be updated to match media.FixCacheStat() if we ever support emoji uncaching.
func (e *Emoji) PruneUnused(ctx context.Context) (int, error) {
	var (
		total int
		page  paging.Page
	)

	// Set page select limit.
	page.Limit = selectLimit

	for {
		// Fetch the next batch of emoji to next max ID.
		emojis, err := e.state.DB.GetRemoteEmojis(ctx, &page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting remote emojis: %w", err)
		}

		// Get current max ID.
		maxID := page.Max.Value

		// If no emoji or the same group is returned, we reached end.
		if len(emojis) == 0 || maxID == emojis[len(emojis)-1].ID {
			break
		}

		// Use last ID as the next 'maxID'.
		maxID = emojis[len(emojis)-1].ID
		page.Max = paging.MaxID(maxID)

		for _, emoji := range emojis {
			// Check / prune unused emoji media.
			fixed, err := e.pruneUnused(ctx, emoji)
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

// FixCacheStatus will check all emoji for up-to-date cache status (i.e. in storage driver).
// Context will be checked for `gtscontext.DryRun()` to perform the action. NOTE: this function
// should be updated to match media.FixCacheStat() if we ever support emoji uncaching.
func (e *Emoji) FixCacheStates(ctx context.Context) (int, error) {
	var (
		total int
		page  paging.Page
	)

	// Set page select limit.
	page.Limit = selectLimit

	for {
		// Fetch the next batch of emoji to next max ID.
		emojis, err := e.state.DB.GetRemoteEmojis(ctx, &page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting remote emojis: %w", err)
		}

		// Get current max ID.
		maxID := page.Max.Value

		// If no emoji or the same group is returned, we reached end.
		if len(emojis) == 0 || maxID == emojis[len(emojis)-1].ID {
			break
		}

		// Use last ID as the next 'maxID'.
		maxID = emojis[len(emojis)-1].ID
		page.Max = paging.MaxID(maxID)

		for _, emoji := range emojis {
			// Check / fix required emoji cache states.
			fixed, err := e.fixCacheState(ctx, emoji)
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

func (e *Emoji) pruneUnused(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	// Start a log entry for emoji.
	l := log.WithContext(ctx).
		WithField("emoji", emoji.ID)

	// Load any related accounts using this emoji.
	accounts, err := e.getRelatedAccounts(ctx, emoji)
	if err != nil {
		return false, err
	} else if len(accounts) > 0 {
		l.Debug("skipping as account emoji in use")
		return false, nil
	}

	// Load any related statuses using this emoji.
	statuses, err := e.getRelatedStatuses(ctx, emoji)
	if err != nil {
		return false, err
	} else if len(statuses) > 0 {
		l.Debug("skipping as status emoji in use")
		return false, nil
	}

	// Check not recently created, give it some time to be "used" again.
	if time.Now().Add(-24 * time.Hour * 7).Before(emoji.CreatedAt) {
		l.Debug("skipping due to recently created")
		return false, nil
	}

	// Emoji totally unused, delete it.
	l.Debug("deleting unused emoji")
	return true, e.delete(ctx, emoji)
}

func (e *Emoji) fixCacheState(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	// Start a log entry for emoji.
	l := log.WithContext(ctx).
		WithField("emoji", emoji.ID)

	// Check whether files exist.
	exist, err := e.haveFiles(ctx,
		emoji.ImageStaticPath,
		emoji.ImagePath,
	)
	if err != nil {
		return false, err
	}

	switch {
	case *emoji.Cached && !exist:
		// Mark as uncached if expected files don't exist.
		l.Debug("cached=true exists=false => marking uncached")
		return true, e.uncache(ctx, emoji)

	case !*emoji.Cached && exist:
		// Remove files if we don't expect them to exist.
		l.Debug("cached=false exists=true => removing files")
		_, err := e.removeFiles(ctx,
			emoji.ImageStaticPath,
			emoji.ImagePath,
		)
		return true, err

	default:
		return false, nil
	}
}

func (e *Emoji) uncacheRemote(ctx context.Context, after time.Time, emoji *gtsmodel.Emoji) (bool, error) {
	if !*emoji.Cached {
		// Already uncached.
		return false, nil
	}

	// Start a log entry for emoji.
	l := log.WithContext(ctx).
		WithField("emoji", emoji.ID)

	// Load any related accounts using this emoji.
	accounts, err := e.getRelatedAccounts(ctx, emoji)
	if err != nil {
		return false, err
	}

	for _, account := range accounts {
		if account.FetchedAt.After(after) {
			l.Debug("skipping due to recently fetched account")
			return false, nil
		}
	}

	// Load any related statuses using this emoji.
	statuses, err := e.getRelatedStatuses(ctx, emoji)
	if err != nil {
		return false, err
	}

	for _, status := range statuses {
		// Check if recently used status.
		if status.FetchedAt.After(after) {
			l.Debug("skipping due to recently fetched status")
			return false, nil
		}

		// Check whether status is bookmarked by active accounts.
		bookmarked, err := e.state.DB.IsStatusBookmarked(ctx, status.ID)
		if err != nil {
			return false, err
		} else if bookmarked {
			l.Debug("skipping due to bookmarked status")
			return false, nil
		}
	}

	// This emoji is too old, uncache it.
	l.Debug("uncaching old remote emoji")
	return true, e.uncache(ctx, emoji)
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
		log.Debugf(ctx, "fixing missing emoji category: %s", emoji.ID)
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

func (e *Emoji) getRelatedAccounts(ctx context.Context, emoji *gtsmodel.Emoji) ([]*gtsmodel.Account, error) {
	accounts, err := e.state.DB.GetAccountsUsingEmoji(ctx, emoji.ID)
	if err != nil {
		return nil, gtserror.Newf("error fetching accounts using emoji %s: %w", emoji.ID, err)
	}
	return accounts, nil
}

func (e *Emoji) getRelatedStatuses(ctx context.Context, emoji *gtsmodel.Emoji) ([]*gtsmodel.Status, error) {
	statuses, err := e.state.DB.GetStatusesUsingEmoji(ctx, emoji.ID)
	if err != nil {
		return nil, gtserror.Newf("error fetching statuses using emoji %s: %w", emoji.ID, err)
	}
	return statuses, nil
}

func (e *Emoji) uncache(ctx context.Context, emoji *gtsmodel.Emoji) error {
	if gtscontext.DryRun(ctx) {
		// Dry run, do nothing.
		return nil
	}

	// Remove emoji and static.
	_, err := e.removeFiles(ctx,
		emoji.ImagePath,
		emoji.ImageStaticPath,
	)
	if err != nil {
		return gtserror.Newf("error removing emoji files: %w", err)
	}

	// Update emoji to reflect that we no longer have it cached.
	log.Debugf(ctx, "marking emoji as uncached: %s", emoji.ID)
	emoji.Cached = func() *bool { i := false; return &i }()
	if err := e.state.DB.UpdateEmoji(ctx, emoji, "cached"); err != nil {
		return gtserror.Newf("error updating emoji: %w", err)
	}

	return nil
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
