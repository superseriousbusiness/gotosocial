/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (m *manager) PruneOrphaned(ctx context.Context, dry bool) (int, error) {
	var totalPruned int

	// keys in storage will look like the following:
	// `[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[MEDIA_ID].[EXTENSION]`
	// we can filter out keys we're not interested in by
	// matching through a regex
	var matchCount int
	match := func(storageKey string) bool {
		if regexes.FilePath.MatchString(storageKey) {
			matchCount++
			return true
		}
		return false
	}

	log.Info("checking storage keys for orphaned pruning candidates...")
	iterator, err := m.storage.Iterator(ctx, match)
	if err != nil {
		return 0, fmt.Errorf("PruneOrphaned: error getting storage iterator: %w", err)
	}

	// make sure we have some keys, and also advance
	// the iterator to the first non-empty key
	if !iterator.Next() {
		return 0, nil
	}

	instanceAccount, err := m.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return 0, fmt.Errorf("PruneOrphaned: error getting instance account: %w", err)
	}
	instanceAccountID := instanceAccount.ID

	// for each key in the iterator, check if entry is orphaned
	log.Info("got %d orphaned pruning candidates, checking for orphaned status, please wait...")
	var checkedKeys int
	orphanedKeys := make([]string, 0, matchCount)
	for key := iterator.Key(); iterator.Next(); key = iterator.Key() {
		if m.orphaned(ctx, key, instanceAccountID) {
			orphanedKeys = append(orphanedKeys, key)
		}
		checkedKeys++
		if checkedKeys%50 == 0 {
			log.Infof("checked %d of %d orphaned pruning candidates...", checkedKeys, matchCount)
		}
	}
	iterator.Release()

	if !dry {
		// the real deal, we have to delete stuff
		for _, key := range orphanedKeys {
			log.Infof("key %s corresponds to orphaned media, will remove it now", key)
			if err := m.storage.Delete(ctx, key); err != nil {
				log.Errorf("error deleting item with key %s from storage: %s", key, err)
				continue
			}
			totalPruned++
		}
	} else {
		// just a dry run, don't delete anything
		for _, key := range orphanedKeys {
			log.Infof("DRY RUN: key %s corresponds to orphaned media which would be deleted", key)
			totalPruned++
		}
	}

	return totalPruned, nil
}

func (m *manager) orphaned(ctx context.Context, key string, instanceAccountID string) bool {
	pathParts := regexes.FilePath.FindStringSubmatch(key)
	if len(pathParts) != 6 {
		return false
	}

	mediaType := pathParts[2]
	mediaID := pathParts[4]

	var orphaned bool
	switch Type(mediaType) {
	case TypeAttachment, TypeHeader, TypeAvatar:
		if _, err := m.db.GetAttachmentByID(ctx, mediaID); err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				orphaned = true
			} else {
				log.Errorf("orphaned: error calling GetAttachmentByID: %s", err)
			}
		}
	case TypeEmoji:
		// look using the static URL for the emoji, since the MEDIA_ID part of
		// the key for emojis will not necessarily correspond to the file that's
		// currently being used as the emoji image
		staticURI := uris.GenerateURIForAttachment(instanceAccountID, string(TypeEmoji), string(SizeStatic), mediaID, mimePng)
		if _, err := m.db.GetEmojiByStaticURL(ctx, staticURI); err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				orphaned = true
			} else {
				log.Errorf("orphaned: error calling GetEmojiByID: %s", err)
			}
		}
	default:
		orphaned = true
	}

	return orphaned
}
