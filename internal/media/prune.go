/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package media

import (
	"context"
	"errors"
	"fmt"
	"time"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	selectPruneLimit          = 50 // Amount of media entries to select at a time from the db when pruning.
	unusedLocalAttachmentDays = 3  // Number of days to keep local media in storage if not attached to a status.
)

func (m *manager) PruneAll(ctx context.Context, mediaCacheRemoteDays int, blocking bool) error {
	const dry = false

	f := func(innerCtx context.Context) error {
		errs := gtserror.MultiError{}

		pruned, err := m.PruneUnusedLocal(innerCtx, dry)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error pruning unused local media (%s)", err))
		} else {
			log.Infof(ctx, "pruned %d unused local media", pruned)
		}

		pruned, err = m.PruneUnusedRemote(innerCtx, dry)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error pruning unused remote media: (%s)", err))
		} else {
			log.Infof(ctx, "pruned %d unused remote media", pruned)
		}

		pruned, err = m.UncacheRemote(innerCtx, mediaCacheRemoteDays, dry)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error uncacheing remote media older than %d day(s): (%s)", mediaCacheRemoteDays, err))
		} else {
			log.Infof(ctx, "uncached %d remote media older than %d day(s)", pruned, mediaCacheRemoteDays)
		}

		pruned, err = m.PruneOrphaned(innerCtx, dry)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error pruning orphaned media: (%s)", err))
		} else {
			log.Infof(ctx, "pruned %d orphaned media", pruned)
		}

		if err := m.state.Storage.Storage.Clean(innerCtx); err != nil {
			errs = append(errs, fmt.Sprintf("error cleaning storage: (%s)", err))
		} else {
			log.Info(ctx, "cleaned storage")
		}

		return errs.Combine()
	}

	if blocking {
		return f(ctx)
	}

	go func() {
		if err := f(context.Background()); err != nil {
			log.Error(ctx, err)
		}
	}()

	return nil
}

func (m *manager) PruneUnusedRemote(ctx context.Context, dry bool) (int, error) {
	var (
		totalPruned int
		maxID       string
		attachments []*gtsmodel.MediaAttachment
		err         error
	)

	// We don't know in advance how many remote attachments will meet
	// our criteria for being 'unused'. So a dry run in this case just
	// means we iterate through as normal, but do nothing with each entry
	// instead of removing it. Define this here so we don't do the 'if dry'
	// check inside the loop a million times.
	var f func(ctx context.Context, attachment *gtsmodel.MediaAttachment) error
	if !dry {
		f = m.deleteAttachment
	} else {
		f = func(_ context.Context, _ *gtsmodel.MediaAttachment) error {
			return nil // noop
		}
	}

	for attachments, err = m.state.DB.GetAvatarsAndHeaders(ctx, maxID, selectPruneLimit); err == nil && len(attachments) != 0; attachments, err = m.state.DB.GetAvatarsAndHeaders(ctx, maxID, selectPruneLimit) {
		maxID = attachments[len(attachments)-1].ID // use the id of the last attachment in the slice as the next 'maxID' value

		for _, attachment := range attachments {
			// Retrieve owning account if possible.
			var account *gtsmodel.Account
			if accountID := attachment.AccountID; accountID != "" {
				account, err = m.state.DB.GetAccountByID(ctx, attachment.AccountID)
				if err != nil && !errors.Is(err, db.ErrNoEntries) {
					// Only return on a real error.
					return 0, fmt.Errorf("PruneUnusedRemote: error fetching account with id %s: %w", accountID, err)
				}
			}

			// Prune each attachment that meets one of the following criteria:
			// - Has no owning account in the database.
			// - Is a header but isn't the owning account's current header.
			// - Is an avatar but isn't the owning account's current avatar.
			if account == nil ||
				(*attachment.Header && attachment.ID != account.HeaderMediaAttachmentID) ||
				(*attachment.Avatar && attachment.ID != account.AvatarMediaAttachmentID) {
				if err := f(ctx, attachment); err != nil {
					return totalPruned, err
				}
				totalPruned++
			}
		}
	}

	// Make sure we don't have a real error when we leave the loop.
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return totalPruned, err
	}

	return totalPruned, nil
}

func (m *manager) PruneOrphaned(ctx context.Context, dry bool) (int, error) {
	// keys in storage will look like the following:
	// `[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[MEDIA_ID].[EXTENSION]`
	// We can filter out keys we're not interested in by
	// matching through a regex.
	var matchCount int
	match := func(storageKey string) bool {
		if regexes.FilePath.MatchString(storageKey) {
			matchCount++
			return true
		}
		return false
	}

	iterator, err := m.state.Storage.Iterator(ctx, match) // make sure this iterator is always released
	if err != nil {
		return 0, fmt.Errorf("PruneOrphaned: error getting storage iterator: %w", err)
	}

	// Ensure we have some keys, and also advance
	// the iterator to the first non-empty key.
	if !iterator.Next() {
		iterator.Release()
		return 0, nil // nothing else to do here
	}

	// Emojis are stored under the instance account,
	// so we need the ID of the instance account for
	// the next part.
	instanceAccount, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		iterator.Release()
		return 0, fmt.Errorf("PruneOrphaned: error getting instance account: %w", err)
	}
	instanceAccountID := instanceAccount.ID

	// For each key in the iterator, check if entry is orphaned.
	orphanedKeys := make([]string, 0, matchCount)
	for key := iterator.Key(); iterator.Next(); key = iterator.Key() {
		orphaned, err := m.orphaned(ctx, key, instanceAccountID)
		if err != nil {
			iterator.Release()
			return 0, fmt.Errorf("PruneOrphaned: checking orphaned status: %w", err)
		}

		if orphaned {
			orphanedKeys = append(orphanedKeys, key)
		}
	}
	iterator.Release()

	totalPruned := len(orphanedKeys)

	if dry {
		// Dry run: don't remove anything.
		return totalPruned, nil
	}

	// This is not a drill!
	// We have to delete stuff!
	return totalPruned, m.removeFiles(ctx, orphanedKeys...)
}

func (m *manager) orphaned(ctx context.Context, key string, instanceAccountID string) (bool, error) {
	pathParts := regexes.FilePath.FindStringSubmatch(key)
	if len(pathParts) != 6 {
		// This doesn't match our expectations so
		// it wasn't created by gts; ignore it.
		return false, nil
	}

	var (
		mediaType = pathParts[2]
		mediaID   = pathParts[4]
		orphaned  = false
	)

	// Look for keys in storage that we don't have an attachment for.
	switch Type(mediaType) {
	case TypeAttachment, TypeHeader, TypeAvatar:
		if _, err := m.state.DB.GetAttachmentByID(ctx, mediaID); err != nil {
			if !errors.Is(err, db.ErrNoEntries) {
				return false, fmt.Errorf("error calling GetAttachmentByID: %w", err)
			}
			orphaned = true
		}
	case TypeEmoji:
		// Look using the static URL for the emoji. Emoji images can change, so
		// the MEDIA_ID part of the key for emojis will not necessarily correspond
		// to the file that's currently being used as the emoji image.
		staticURL := uris.GenerateURIForAttachment(instanceAccountID, string(TypeEmoji), string(SizeStatic), mediaID, mimePng)
		if _, err := m.state.DB.GetEmojiByStaticURL(ctx, staticURL); err != nil {
			if !errors.Is(err, db.ErrNoEntries) {
				return false, fmt.Errorf("error calling GetEmojiByStaticURL: %w", err)
			}
			orphaned = true
		}
	}

	return orphaned, nil
}

func (m *manager) UncacheRemote(ctx context.Context, olderThanDays int, dry bool) (int, error) {
	if olderThanDays < 0 {
		return 0, nil
	}

	olderThan := time.Now().Add(-time.Hour * 24 * time.Duration(olderThanDays))

	if dry {
		// Dry run, just count eligible entries without removing them.
		return m.state.DB.CountRemoteOlderThan(ctx, olderThan)
	}

	var (
		totalPruned int
		attachments []*gtsmodel.MediaAttachment
		err         error
	)

	for attachments, err = m.state.DB.GetRemoteOlderThan(ctx, olderThan, selectPruneLimit); err == nil && len(attachments) != 0; attachments, err = m.state.DB.GetRemoteOlderThan(ctx, olderThan, selectPruneLimit) {
		olderThan = attachments[len(attachments)-1].CreatedAt // use the created time of the last attachment in the slice as the next 'olderThan' value

		for _, attachment := range attachments {
			if err := m.uncacheAttachment(ctx, attachment); err != nil {
				return totalPruned, err
			}
			totalPruned++
		}
	}

	// Make sure we don't have a real error when we leave the loop.
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return totalPruned, err
	}

	return totalPruned, nil
}

func (m *manager) PruneUnusedLocal(ctx context.Context, dry bool) (int, error) {
	olderThan := time.Now().Add(-time.Hour * 24 * time.Duration(unusedLocalAttachmentDays))

	if dry {
		// Dry run, just count eligible entries without removing them.
		return m.state.DB.CountLocalUnattachedOlderThan(ctx, olderThan)
	}

	var (
		totalPruned int
		attachments []*gtsmodel.MediaAttachment
		err         error
	)

	for attachments, err = m.state.DB.GetLocalUnattachedOlderThan(ctx, olderThan, selectPruneLimit); err == nil && len(attachments) != 0; attachments, err = m.state.DB.GetLocalUnattachedOlderThan(ctx, olderThan, selectPruneLimit) {
		olderThan = attachments[len(attachments)-1].CreatedAt // use the created time of the last attachment in the slice as the next 'olderThan' value

		for _, attachment := range attachments {
			if err := m.deleteAttachment(ctx, attachment); err != nil {
				return totalPruned, err
			}
			totalPruned++
		}
	}

	// Make sure we don't have a real error when we leave the loop.
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return totalPruned, err
	}

	return totalPruned, nil
}

/*
	Handy little helpers
*/

func (m *manager) deleteAttachment(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	if err := m.removeFiles(ctx, attachment.File.Path, attachment.Thumbnail.Path); err != nil {
		return err
	}

	// Delete attachment completely.
	return m.state.DB.DeleteByID(ctx, attachment.ID, attachment)
}

func (m *manager) uncacheAttachment(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	if err := m.removeFiles(ctx, attachment.File.Path, attachment.Thumbnail.Path); err != nil {
		return err
	}

	// Update attachment to reflect that we no longer have it cached.
	attachment.UpdatedAt = time.Now()
	cached := false
	attachment.Cached = &cached
	return m.state.DB.UpdateByID(ctx, attachment, attachment.ID, "updated_at", "cached")
}

func (m *manager) removeFiles(ctx context.Context, keys ...string) error {
	errs := make(gtserror.MultiError, 0, len(keys))

	for _, key := range keys {
		if err := m.state.Storage.Delete(ctx, key); err != nil && !errors.Is(err, storage.ErrNotFound) {
			errs = append(errs, "storage error removing "+key+": "+err.Error())
		}
	}

	return errs.Combine()
}
