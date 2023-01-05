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
	"time"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (m *manager) PruneAllRemote(ctx context.Context, olderThanDays int) (int, error) {
	var totalPruned int

	olderThan := time.Now().Add(-time.Hour * 24 * time.Duration(olderThanDays))
	log.Infof("PruneAllRemote: pruning media older than %s", olderThan)

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

		// prune each status attachment
		for _, attachment := range attachments {
			if err := m.pruneOneRemote(ctx, attachment); err != nil {
				return totalPruned, err
			}
			totalPruned++
		}
	}

	log.Infof("PruneAllRemote: finished pruning remote media: pruned %d entries", totalPruned)
	return totalPruned, nil
}

func (m *manager) pruneOneRemote(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	var changed bool

	if attachment.File.Path != "" {
		// delete the full size attachment from storage
		log.Tracef("pruneOneRemote: deleting %s", attachment.File.Path)
		if err := m.storage.Delete(ctx, attachment.File.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		cached := false
		attachment.Cached = &cached
		changed = true
	}

	if attachment.Thumbnail.Path != "" {
		// delete the thumbnail from storage
		log.Tracef("pruneOneRemote: deleting %s", attachment.Thumbnail.Path)
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
