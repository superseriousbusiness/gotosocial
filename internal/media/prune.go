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

// amount of media entries to select at a time from the db when pruning
const selectPruneLimit = 20

func (m *manager) PruneAll(ctx context.Context, mediaCacheRemoteDays int, blocking bool) error {
	f := func(innerCtx context.Context) error {
		errs := gtserror.MultiError{}

		pruned, err := m.PruneUnusedLocal(innerCtx)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error pruning unused local media (%s)", err))
		} else {
			log.Infof("pruned %d unused local media", pruned)
		}

		pruned, err = m.PruneUnusedRemote(innerCtx)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error pruning unused remote media: (%s)", err))
		} else {
			log.Infof("pruned %d unused remote media", pruned)
		}

		pruned, err = m.UncacheRemote(innerCtx, mediaCacheRemoteDays, false)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error uncacheing remote media older than %d day(s): (%s)", mediaCacheRemoteDays, err))
		} else {
			log.Infof("uncached %d remote media older than %d day(s)", pruned, mediaCacheRemoteDays)
		}

		pruned, err = m.PruneOrphaned(innerCtx, false)
		if err != nil {
			errs = append(errs, fmt.Sprintf("error pruning orphaned media: (%s)", err))
		} else {
			log.Infof("pruned %d orphaned media", pruned)
		}

		if err := m.storage.Storage.Clean(innerCtx); err != nil {
			errs = append(errs, fmt.Sprintf("error cleaning storage: (%s)", err))
		} else {
			log.Info("cleaned storage")
		}

		return errs.Combine()
	}

	if blocking {
		return f(ctx)
	}

	go func() {
		if err := f(context.Background()); err != nil {
			log.Error(err)
		}
	}()

	return nil
}

func (m *manager) deleteAttachment(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	if attachment.File.Path != "" {
		// delete the full size file from storage
		if err := m.storage.Delete(ctx, attachment.File.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
	}

	if attachment.Thumbnail.Path != "" {
		// delete the thumbnail file from storage
		if err := m.storage.Delete(ctx, attachment.Thumbnail.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
	}

	// delete the attachment completely
	return m.db.DeleteByID(ctx, attachment.ID, attachment)
}

func (m *manager) uncacheAttachment(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	var changed bool

	if attachment.File.Path != "" {
		// delete the full size file from storage
		if err := m.storage.Delete(ctx, attachment.File.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		cached := false
		attachment.Cached = &cached
		changed = true
	}

	if attachment.Thumbnail.Path != "" {
		// delete the thumbnail file from storage
		if err := m.storage.Delete(ctx, attachment.Thumbnail.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		cached := false
		attachment.Cached = &cached
		changed = true
	}

	if !changed {
		return nil
	}

	// update the attachment to reflect that we no longer have it cached
	return m.db.UpdateByID(ctx, attachment, attachment.ID, "updated_at", "cached")
}

func (m *manager) PruneUnusedRemote(ctx context.Context) (int, error) {
	var (
		totalPruned int
		maxID       string
		attachments []*gtsmodel.MediaAttachment
		err         error
	)

	for attachments, err = m.db.GetAvatarsAndHeaders(ctx, maxID, selectPruneLimit); err == nil && len(attachments) != 0; attachments, err = m.db.GetAvatarsAndHeaders(ctx, maxID, selectPruneLimit) {
		// use the id of the last attachment in the slice as the next 'maxID' value
		maxID = attachments[len(attachments)-1].ID

		// Prune each attachment that meets one of the following criteria:
		// - has no owning account in the database
		// - is a header but isn't the owning account's current header
		// - is an avatar but isn't the owning account's current avatar
		for _, attachment := range attachments {
			if attachment.Account == nil ||
				(*attachment.Header && attachment.ID != attachment.Account.HeaderMediaAttachmentID) ||
				(*attachment.Avatar && attachment.ID != attachment.Account.AvatarMediaAttachmentID) {
				if err := m.deleteAttachment(ctx, attachment); err != nil {
					return totalPruned, err
				}
				totalPruned++
			}
		}
	}

	// make sure we don't have a real error when we leave the loop
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

	iterator, err := m.storage.Iterator(ctx, match) // make sure this iterator is always released
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
	instanceAccount, err := m.db.GetInstanceAccount(ctx, "")
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

	// Assume we'll prune all the orphans we found.
	pruned := len(orphanedKeys)

	if !dry {
		// This is not a drill!
		// We have to delete stuff!
		for _, key := range orphanedKeys {
			if err := m.storage.Delete(ctx, key); err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					// Weird race condition? we didn't need
					// to prune this one, so don't count it.
					pruned--
					continue
				}
				return 0, fmt.Errorf("PruneOrphaned: error deleting item with key %s from storage: %w", key, err)
			}
		}
	}

	return pruned, nil
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

	switch Type(mediaType) {
	case TypeAttachment, TypeHeader, TypeAvatar:
		if _, err := m.db.GetAttachmentByID(ctx, mediaID); err != nil {
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
		if _, err := m.db.GetEmojiByStaticURL(ctx, staticURL); err != nil {
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

	eligible, err := m.db.CountRemoteOlderThan(ctx, olderThan)
	if err != nil {
		return 0, err
	}

	if dry || eligible == 0 {
		return eligible, nil
	}

	var totalPruned int
	for {
		// Select "selectPruneLimit" status attacchments at a time for pruning
		attachments, err := m.db.GetRemoteOlderThan(ctx, olderThan, selectPruneLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return totalPruned, err
		} else if len(attachments) == 0 {
			break
		}

		// use the age of the oldest attachment (last in slice) as the next 'olderThan' value
		log.Tracef("PruneAllRemote: got %d status attachments older than %s", len(attachments), olderThan)
		olderThan = attachments[len(attachments)-1].CreatedAt

		for _, attachment := range attachments {
			if err := m.uncacheAttachment(ctx, attachment); err != nil {
				return totalPruned, err
			}
			totalPruned++
		}
	}

	return totalPruned, nil
}

func (m *manager) PruneUnusedLocal(ctx context.Context) (int, error) {
	var (
		totalPruned int
		maxID       string
		attachments []*gtsmodel.MediaAttachment
		err         error
		olderThan   = time.Now().Add(-time.Hour * 24 * time.Duration(UnusedLocalAttachmentCacheDays))
	)

	for attachments, err = m.db.GetLocalUnattachedOlderThan(ctx, olderThan, maxID, selectPruneLimit); err == nil && len(attachments) != 0; attachments, err = m.db.GetLocalUnattachedOlderThan(ctx, olderThan, maxID, selectPruneLimit) {
		// use the id of the last attachment in the slice as the next 'maxID' value
		maxID = attachments[len(attachments)-1].ID

		for _, attachment := range attachments {
			if err := m.deleteAttachment(ctx, attachment); err != nil {
				return totalPruned, err
			}
			totalPruned++
		}
	}

	// make sure we don't have a real error when we leave the loop
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return totalPruned, err
	}

	return totalPruned, nil
}
