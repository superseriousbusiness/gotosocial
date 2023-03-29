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

package timeline

import (
	"context"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) prepareXBetweenIDs(ctx context.Context, amount int, behindID string, beforeID string, frontToBack bool) error {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"amount", amount},
			{"behindID", behindID},
			{"beforeID", beforeID},
			{"frontToBack", frontToBack},
		}...)
	l.Trace("entering prepareXBetweenIDs")

	if beforeID >= behindID {
		// This is an impossible situation, we
		// can't prepare anything between these.
		return nil
	}

	if err := t.indexXBetweenIDs(ctx, amount, behindID, beforeID, frontToBack); err != nil {
		// An error here doesn't necessarily mean we
		// can't prepare anything, so log + keep going.
		l.Debugf("error calling prepareXBetweenIDs: %s", err)
	}

	t.Lock()
	defer t.Unlock()

	return nil
}
