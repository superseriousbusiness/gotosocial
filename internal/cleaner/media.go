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
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// Media encompasses a set of
// media cleanup / admin utils.
type Media struct {
	*Cleaner
}

// All will execute all cleaner.Media utilities synchronously, including output logging.
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (m *Media) All(ctx context.Context, maxRemoteDays int) {
	t := time.Now().Add(-24 * time.Hour * time.Duration(maxRemoteDays))
	m.LogUncacheRemote(ctx, t)
	m.LogPruneOrphaned(ctx)
	m.LogPruneUnused(ctx)
	m.LogFixCacheStates(ctx)
	_ = m.state.Storage.Storage.Clean(ctx)
}

// LogUncacheRemote performs Media.UncacheRemote(...), logging the start and outcome.
func (m *Media) LogUncacheRemote(ctx context.Context, olderThan time.Time) {
	log.Infof(ctx, "start older than: %s", olderThan.Format(time.Stamp))
	if n, err := m.UncacheRemote(ctx, olderThan); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "uncached: %d", n)
	}
}

// LogPruneOrphaned performs Media.PruneOrphaned(...), logging the start and outcome.
func (m *Media) LogPruneOrphaned(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := m.PruneOrphaned(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// LogPruneUnused performs Media.PruneUnused(...), logging the start and outcome.
func (m *Media) LogPruneUnused(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := m.PruneUnused(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// LogFixCacheStates performs Media.FixCacheStates(...), logging the start and outcome.
func (m *Media) LogFixCacheStates(ctx context.Context) {
	log.Info(ctx, "start")
	if n, err := m.FixCacheStates(ctx); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "fixed: %d", n)
	}
}

// PruneOrphaned will delete orphaned files from storage (i.e. media missing a database entry).
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (m *Media) PruneOrphaned(ctx context.Context) (int, error) {
	var files []string

	// All media files in storage will have path fitting: {$account}/{$type}/{$size}/{$id}.{$ext}
	if err := m.state.Storage.WalkKeys(ctx, func(ctx context.Context, path string) error {
		if !regexes.FilePath.MatchString(path) {
			// This is not our expected media
			// path format, skip this one.
			return nil
		}

		// Check whether this entry is orphaned.
		orphaned, err := m.isOrphaned(ctx, path)
		if err != nil {
			return gtserror.Newf("error checking orphaned status: %w", err)
		}

		if orphaned {
			// Add this orphaned entry.
			files = append(files, path)
		}

		return nil
	}); err != nil {
		return 0, gtserror.Newf("error walking storage: %w", err)
	}

	// Delete all orphaned files from storage.
	return m.removeFiles(ctx, files...)
}

// PruneUnused will delete all unused media attachments from the database and storage driver.
// Media is marked as unused if not attached to any status, account or account is suspended.
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (m *Media) PruneUnused(ctx context.Context) (int, error) {
	var (
		total int
		maxID string
	)

	for {
		// Fetch the next batch of media attachments up to next max ID.
		attachments, err := m.state.DB.GetAttachments(ctx, maxID, selectLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting attachments: %w", err)
		}

		if len(attachments) == 0 {
			// reached end.
			break
		}

		// Use last ID as the next 'maxID' value.
		maxID = attachments[len(attachments)-1].ID

		for _, media := range attachments {
			// Check / prune unused media attachment.
			fixed, err := m.pruneUnused(ctx, media)
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

// UncacheRemote will uncache all remote media attachments older than given input time.
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (m *Media) UncacheRemote(ctx context.Context, olderThan time.Time) (int, error) {
	var total int

	// Drop time by an hour to improve search.
	olderThan = olderThan.Add(-time.Hour)

	// Store recent time.
	mostRecent := olderThan

	for {
		// Fetch the next batch of attachments older than last-set time.
		attachments, err := m.state.DB.GetRemoteOlderThan(ctx, olderThan, selectLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting remote media: %w", err)
		}

		if len(attachments) == 0 {
			// reached end.
			break
		}

		// Use last created-at as the next 'olderThan' value.
		olderThan = attachments[len(attachments)-1].CreatedAt

		for _, media := range attachments {
			// Check / uncache each remote media attachment.
			uncached, err := m.uncacheRemote(ctx, mostRecent, media)
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

// FixCacheStatus will check all media for up-to-date cache status (i.e. in storage driver).
// Media marked as cached, with any required files missing, will be automatically uncached.
// Context will be checked for `gtscontext.DryRun()` in order to actually perform the action.
func (m *Media) FixCacheStates(ctx context.Context) (int, error) {
	var (
		total int
		maxID string
	)

	for {
		// Fetch the next batch of media attachments up to next max ID.
		attachments, err := m.state.DB.GetAttachments(ctx, maxID, selectLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error getting avatars / headers: %w", err)
		}

		if len(attachments) == 0 {
			// reached end.
			break
		}

		// Use last ID as the next 'maxID' value.
		maxID = attachments[len(attachments)-1].ID

		for _, media := range attachments {
			// Check / fix required media cache states.
			fixed, err := m.fixCacheState(ctx, media)
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

func (m *Media) isOrphaned(ctx context.Context, path string) (bool, error) {
	pathParts := regexes.FilePath.FindStringSubmatch(path)
	if len(pathParts) != 6 {
		// This doesn't match our expectations so
		// it wasn't created by gts; ignore it.
		return false, nil
	}

	var (
		// 0th -> whole match
		// 1st -> account ID
		mediaType = pathParts[2]
		// 3rd -> media sub-type (e.g. small, static)
		mediaID = pathParts[4]
		// 5th -> file extension
	)

	switch media.Type(mediaType) {
	case media.TypeAttachment:
		// Look for media in database stored by ID.
		media, err := m.state.DB.GetAttachmentByID(
			gtscontext.SetBarebones(ctx),
			mediaID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return false, gtserror.Newf("error fetching media by id %s: %w", mediaID, err)
		}

		if media == nil {
			// Not found in DB.
			return true, nil
		}

	case media.TypeEmoji:
		// Generate static URL for this emoji to lookup.
		staticURL := uris.GenerateURIForAttachment(
			pathParts[1], // instance account ID
			string(media.TypeEmoji),
			string(media.SizeStatic),
			mediaID,
			"png",
		)

		// Look for emoji in database stored by static URL.
		// The media ID part of the storage key for emojis can
		// change for refreshed items, so search by generated URL.
		emoji, err := m.state.DB.GetEmojiByID(
			gtscontext.SetBarebones(ctx),
			staticURL,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return false, gtserror.Newf("error fetching emoji by url %s: %w", staticURL, err)
		}

		if emoji == nil {
			// Not found in DB.
			return true, nil
		}
	}

	return false, nil
}

func (m *Media) pruneUnused(ctx context.Context, media *gtsmodel.MediaAttachment) (bool, error) {
	// Check whether we have the required account for media.
	account, missing, err := m.getRelatedAccount(ctx, media)
	if err != nil {
		return false, err
	} else if missing {
		// Media account missing, delete it.
		return true, m.delete(ctx, media)
	}

	if account != nil {
		// Related account exists for this media, check whether it is being used.
		headerInUse := (*media.Header && media.ID == account.HeaderMediaAttachmentID)
		avatarInUse := (*media.Avatar && media.ID == account.AvatarMediaAttachmentID)
		if (headerInUse || avatarInUse) && account.SuspendedAt.IsZero() {
			// note that suspended accounts do not count as in-use.
			return false, nil
		}
	}

	// Check whether we have the required status for media.
	status, missing, err := m.getRelatedStatus(ctx, media)
	if err != nil {
		return false, err
	} else if missing {
		// Media status missing, delete it.
		return true, m.delete(ctx, media)
	}

	if status != nil {
		// Check whether still attached to status.
		for _, id := range status.AttachmentIDs {
			if id == media.ID {
				return false, nil
			}
		}
	}

	// Media totally unused, delete it.
	return true, m.delete(ctx, media)
}

func (m *Media) fixCacheState(ctx context.Context, media *gtsmodel.MediaAttachment) (bool, error) {
	if !*media.Cached {
		// We ignore uncached media, a
		// false negative is a much better
		// situation than a false positive,
		// re-cache will just overwrite it.
		return false, nil
	}

	// Check whether we have the required account for media.
	_, missingAccount, err := m.getRelatedAccount(ctx, media)
	if err != nil || missingAccount {
		return false, err
	}

	// Check whether we have the required status for media.
	_, missingStatus, err := m.getRelatedStatus(ctx, media)
	if err != nil || missingStatus {
		return false, err
	}

	// So we know this a valid cached media entry.
	// Check that we have the files on disk required....
	return m.checkFiles(ctx, func() error {
		// Media missing files, uncache it.
		return m.uncache(ctx, media)
	},
		media.Thumbnail.Path,
		media.File.Path,
	)
}

func (m *Media) uncacheRemote(ctx context.Context, after time.Time, media *gtsmodel.MediaAttachment) (bool, error) {
	// Check whether we have the required account for media.
	account, missing, err := m.getRelatedAccount(ctx, media)
	if err != nil {
		return false, err
	} else if missing {
		// Account missing, skip it.
		return false, nil
	}

	if account != nil && account.FetchedAt.After(after) {
		// This is a recently loaded account,
		// no need to uncache this media.
		return false, nil
	}

	// Check whether we have the required status for media.
	status, missing, err := m.getRelatedStatus(ctx, media)
	if err != nil {
		return false, err
	} else if missing {
		// Status missing, skip it.
		return false, nil
	}

	if status != nil && status.FetchedAt.After(after) {
		// This is a recently loaded status,
		// no need to uncache this media.
		return false, nil
	}

	// This media is too old, uncache it.
	return true, m.uncache(ctx, media)
}

func (m *Media) getRelatedAccount(ctx context.Context, media *gtsmodel.MediaAttachment) (*gtsmodel.Account, bool, error) {
	if media.AccountID == "" {
		// no related account.
		return nil, false, nil
	}

	// Load the account related to this media.
	account, err := m.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		media.AccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, false, gtserror.Newf("error fetching account by id %s: %w", media.AccountID, err)
	}

	if account == nil {
		log.Warnf(ctx, "account missing for %s", media.File.Path)
		return nil, true, nil
	}

	return account, false, nil
}

func (m *Media) getRelatedStatus(ctx context.Context, media *gtsmodel.MediaAttachment) (*gtsmodel.Status, bool, error) {
	if media.StatusID == "" {
		// no related status.
		return nil, false, nil
	}

	// Load the status related to this media.
	status, err := m.state.DB.GetStatusByID(
		gtscontext.SetBarebones(ctx),
		media.StatusID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, false, gtserror.Newf("error fetching status by id %s: %w", media.StatusID, err)
	}

	if status == nil {
		log.Warnf(ctx, "status missing for %s", media.File.Path)
		return nil, true, nil
	}

	return status, false, nil
}

func (m *Media) uncache(ctx context.Context, media *gtsmodel.MediaAttachment) error {
	if gtscontext.DryRun(ctx) {
		// Dry run, do nothing.
		return nil
	}

	// Remove media and thumbnail.
	_, err := m.removeFiles(ctx,
		media.File.Path,
		media.Thumbnail.Path,
	)
	if err != nil {
		return gtserror.Newf("error removing media files: %w", err)
	}

	// Update attachment to reflect that we no longer have it cached.
	media.Cached = func() *bool { i := false; return &i }()
	if err := m.state.DB.UpdateAttachment(ctx, media, "cached"); err != nil {
		return gtserror.Newf("error updating media: %w", err)
	}

	return nil
}

func (m *Media) delete(ctx context.Context, media *gtsmodel.MediaAttachment) error {
	if gtscontext.DryRun(ctx) {
		// Dry run, do nothing.
		return nil
	}

	// Remove media and thumbnail.
	_, err := m.removeFiles(ctx,
		media.File.Path,
		media.Thumbnail.Path,
	)
	if err != nil {
		return gtserror.Newf("error removing media files: %w", err)
	}

	// Delete media attachment entirely from the database.
	if err := m.state.DB.DeleteAttachment(ctx, media.ID); err != nil {
		return gtserror.Newf("error deleting media: %w", err)
	}

	return nil
}
