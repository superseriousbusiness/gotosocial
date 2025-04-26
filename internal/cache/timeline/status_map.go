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
	"maps"
	"sync/atomic"
)

// StatusTimelines is a concurrency safe map of StatusTimeline{}
// objects, optimizing *very heavily* for reads over writes.
type StatusTimelines struct {
	ptr atomic.Pointer[map[string]*StatusTimeline] // ronly except by CAS
	cap int
}

// Init stores the given argument(s) such that any created StatusTimeline{}
// objects by MustGet() will initialize them with the given arguments.
func (t *StatusTimelines) Init(cap int) { t.cap = cap }

// MustGet will attempt to fetch StatusTimeline{} stored under key, else creating one.
func (t *StatusTimelines) MustGet(key string) *StatusTimeline {
	var tt *StatusTimeline

	for {
		// Load current ptr.
		cur := t.ptr.Load()

		// Get timeline map to work on.
		var m map[string]*StatusTimeline

		if cur != nil {
			// Look for existing
			// timeline in cache.
			tt = (*cur)[key]
			if tt != nil {
				return tt
			}

			// Get clone of current
			// before modifications.
			m = maps.Clone(*cur)
		} else {
			// Allocate new timeline map for below.
			m = make(map[string]*StatusTimeline)
		}

		if tt == nil {
			// Allocate new timeline.
			tt = new(StatusTimeline)
			tt.Init(t.cap)
		}

		// Store timeline
		// in new map.
		m[key] = tt

		// Attempt to update the map ptr.
		if !t.ptr.CompareAndSwap(cur, &m) {

			// We failed the
			// CAS, reloop.
			continue
		}

		// Successfully inserted
		// new timeline model.
		return tt
	}
}

// Delete will delete the stored StatusTimeline{} under key, if any.
func (t *StatusTimelines) Delete(key string) {
	for {
		// Load current ptr.
		cur := t.ptr.Load()

		// Check for empty map / not in map.
		if cur == nil || (*cur)[key] == nil {
			return
		}

		// Get clone of current
		// before modifications.
		m := maps.Clone(*cur)

		// Delete ID.
		delete(m, key)

		// Attempt to update the map ptr.
		if !t.ptr.CompareAndSwap(cur, &m) {

			// We failed the
			// CAS, reloop.
			continue
		}

		// Successfully
		// deleted ID.
		return
	}
}

// RemoveByStatusIDs calls RemoveByStatusIDs() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) RemoveByStatusIDs(statusIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.RemoveByStatusIDs(statusIDs...)
		}
	}
}

// RemoveByAccountIDs calls RemoveByAccountIDs() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) RemoveByAccountIDs(accountIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.RemoveByAccountIDs(accountIDs...)
		}
	}
}

// UnprepareByStatusIDs calls UnprepareByStatusIDs() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) UnprepareByStatusIDs(statusIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.UnprepareByStatusIDs(statusIDs...)
		}
	}
}

// UnprepareByAccountIDs calls UnprepareByAccountIDs() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) UnprepareByAccountIDs(accountIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.UnprepareByAccountIDs(accountIDs...)
		}
	}
}

// Unprepare attempts to call UnprepareAll() for StatusTimeline{} under key.
func (t *StatusTimelines) Unprepare(key string) {
	if p := t.ptr.Load(); p != nil {
		if tt := (*p)[key]; tt != nil {
			tt.UnprepareAll()
		}
	}
}

// UnprepareAll calls UnprepareAll() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) UnprepareAll() {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.UnprepareAll()
		}
	}
}

// Trim calls Trim() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) Trim() {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.Trim()
		}
	}
}

// Clear attempts to call Clear() for StatusTimeline{} under key.
func (t *StatusTimelines) Clear(key string) {
	if p := t.ptr.Load(); p != nil {
		if tt := (*p)[key]; tt != nil {
			tt.Clear()
		}
	}
}

// ClearAll calls Clear() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) ClearAll() {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.Clear()
		}
	}
}
