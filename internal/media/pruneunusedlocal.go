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
	"fmt"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (m *manager) PruneUnusedLocalAttachments(ctx context.Context) (int, error) {
	var totalPruned int
	var maxID string
	var attachments []*gtsmodel.MediaAttachment
	var err error

	olderThan, err := parseOlderThan(UnusedLocalAttachmentCacheDays)
	if err != nil {
		return totalPruned, fmt.Errorf("PruneUnusedLocalAttachments: error parsing olderThanDays %d: %s", UnusedLocalAttachmentCacheDays, err)
	}
	log.Infof("PruneUnusedLocalAttachments: pruning unused local attachments older than %s", olderThan)

	// select 20 attachments at a time and prune them
	for attachments, err = m.db.GetLocalUnattachedOlderThan(ctx, olderThan, maxID, selectPruneLimit); err == nil && len(attachments) != 0; attachments, err = m.db.GetLocalUnattachedOlderThan(ctx, olderThan, maxID, selectPruneLimit) {
		// use the id of the last attachment in the slice as the next 'maxID' value
		l := len(attachments)
		maxID = attachments[l-1].ID
		log.Tracef("PruneUnusedLocalAttachments: got %d unused local attachments older than %s with maxID < %s", l, olderThan, maxID)

		for _, attachment := range attachments {
			if err := m.pruneOneLocal(ctx, attachment); err != nil {
				return totalPruned, err
			}
			totalPruned++
		}
	}

	// make sure we don't have a real error when we leave the loop
	if err != nil && err != db.ErrNoEntries {
		return totalPruned, err
	}

	log.Infof("PruneUnusedLocalAttachments: finished pruning: pruned %d entries", totalPruned)
	return totalPruned, nil
}

func (m *manager) pruneOneLocal(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	if attachment.File.Path != "" {
		// delete the full size attachment from storage
		log.Tracef("pruneOneLocal: deleting %s", attachment.File.Path)
		if err := m.storage.Delete(ctx, attachment.File.Path); err != nil && err != storage.ErrNotFound {
			return err
		}
	}

	if attachment.Thumbnail.Path != "" {
		// delete the thumbnail from storage
		log.Tracef("pruneOneLocal: deleting %s", attachment.Thumbnail.Path)
		if err := m.storage.Delete(ctx, attachment.Thumbnail.Path); err != nil && err != storage.ErrNotFound {
			return err
		}
	}

	// delete the attachment completely
	return m.db.DeleteByID(ctx, attachment.ID, attachment)
}
