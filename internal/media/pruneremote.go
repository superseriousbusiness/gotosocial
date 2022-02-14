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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// amount of media attachments to select at a time from the db when pruning
const selectPruneLimit = 20

func (m *manager) PruneRemote(ctx context.Context, olderThanDays int) error {
	// convert days into a duration string
	olderThanHoursString := fmt.Sprintf("%dh", olderThanDays*24)
	// parse the duration string into a duration
	olderThanHours, err := time.ParseDuration(olderThanHoursString)
	if err != nil {
		return fmt.Errorf("PruneRemote: %d", err)
	}
	// 'subtract' that from the time now to give our threshold
	olderThan := time.Now().Add(-olderThanHours)

selectLoop:
	for attachments, err := m.db.GetRemoteOlderThan(ctx, olderThan, selectPruneLimit); ; {
		if len(attachments) == 0 || err == db.ErrNoEntries {
			// we're done
			break selectLoop
		}

		if err != nil {
			// there's been a real error
			return err
		}

		for _, attachment := range attachments {
			if err := m.PruneOne(ctx, attachment); err != nil {
				return err
			}
		}
	}

	// make sure we don't have a real error when we leave the loop
	if err != db.ErrNoEntries {
		return err
	}
	return nil
}

func (m *manager) PruneOne(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {

	// delete the full size attachment from storage
	m.storage.Delete(attachment.File.Path)

	return m.db.UpdateByPrimaryKey(ctx, attachment)
}
