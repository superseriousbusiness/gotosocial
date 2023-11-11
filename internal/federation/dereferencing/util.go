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

package dereferencing

import (
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// doOnce wraps a function to only perform it once.
func doOnce(fn func()) func() {
	var once int32
	return func() {
		if once == 0 {
			fn()
			once = 1
		}
	}
}

// pollChanged returns whether a poll has changed in way that
// indicates that this should be an entirely new poll. i.e. if
// the available options have changed, or the expiry has increased.
func pollChanged(existing, latest *gtsmodel.Poll) bool {
	switch {
	case !slices.Equal(existing.Options, latest.Options):
		// easy case, the options changed!
		return true

	case existing.ExpiresAt.Equal(latest.ExpiresAt):
		// again, easy. expiry remained.
		return false

	case latest.ExpiresAt.IsZero() &&
		existing.ClosedAt.IsZero() &&
		!latest.ClosedAt.IsZero():
		// closedAt newly set, and expiresAt
		// unset, indicating a closing poll.
		return false

	default:
		// all other cases
		// we deal as changes
		return true
	}
}

// pollUpdated returns whether a poll has updated, i.e. if the
// vote counts have changed, or if it has expired / been closed.
func pollUpdated(existing, latest *gtsmodel.Poll) bool {
	switch {
	case *existing.Voters != *latest.Voters:
		// easy case, no. votes changed.
		return true

	case !slices.Equal(existing.Votes, latest.Votes):
		// again, easy. per-vote counts changed.
		return true

	case !existing.ClosedAt.Equal(latest.ClosedAt):
		// closedAt has changed, indicating closing.
		return true

	default:
		// no updates.
		return false
	}
}

// pollJustClosed returns whether a poll has *just* closed.
func pollJustClosed(existing, latest *gtsmodel.Poll) bool {
	return existing.ClosedAt.IsZero() && latest.Closed()
}
