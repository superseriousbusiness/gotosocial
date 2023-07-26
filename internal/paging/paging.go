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

package paging

import "golang.org/x/exp/slices"

// Pager provides a means of paging serialized IDs,
// using the terminology of our API endpoint queries.
type Pager struct {
	// SinceID will limit the returned
	// page of IDs to contain newer than
	// since ID (excluding it). Result
	// will be returned DESCENDING.
	SinceID string

	// MinID will limit the returned
	// page of IDs to contain newer than
	// min ID (excluding it). Result
	// will be returned ASCENDING.
	MinID string

	// MaxID will limit the returned
	// page of IDs to contain older
	// than (excluding) this max ID.
	MaxID string

	// Limit will limit the returned
	// page of IDs to at most 'limit'.
	Limit int
}

// Page will page the given slice of GoToSocial IDs according
// to the receiving Pager's SinceID, MinID, MaxID and Limits.
// NOTE THE INPUT SLICE MUST BE SORTED IN ASCENDING ORDER
// (I.E. OLDEST ITEMS AT LOWEST INDICES, NEWER AT HIGHER).
func (p *Pager) PageAsc(ids []string) []string {
	if p == nil {
		// no paging.
		return ids
	}

	var asc bool

	if p.SinceID != "" {
		// If a sinceID is given, we
		// page down i.e. descending.
		asc = false

		for i := 0; i < len(ids); i++ {
			if ids[i] == p.SinceID {
				// Hit the boundary.
				// Reslice to be:
				// "from here"
				ids = ids[i+1:]
				break
			}
		}
	} else if p.MinID != "" {
		// We only support minID if
		// no sinceID is provided.
		//
		// If a minID is given, we
		// page up, i.e. ascending.
		asc = true

		for i := 0; i < len(ids); i++ {
			if ids[i] == p.MinID {
				// Hit the boundary.
				// Reslice to be:
				// "from here"
				ids = ids[i+1:]
				break
			}
		}
	}

	if p.MaxID != "" {
		for i := 0; i < len(ids); i++ {
			if ids[i] == p.MaxID {
				// Hit the boundary.
				// Reslice to be:
				// "up to here"
				ids = ids[:i]
				break
			}
		}
	}

	if !asc && len(ids) > 1 {
		var (
			// Start at front.
			i = 0

			// Start at back.
			j = len(ids) - 1
		)

		// Clone input IDs before
		// we perform modifications.
		ids = slices.Clone(ids)

		for i < j {
			// Swap i,j index values in slice.
			ids[i], ids[j] = ids[j], ids[i]

			// incr + decr,
			// looping until
			// they meet in
			// the middle.
			i++
			j--
		}
	}

	if p.Limit > 0 && p.Limit < len(ids) {
		// Reslice IDs to given limit.
		ids = ids[:p.Limit]
	}

	return ids
}

// Page will page the given slice of GoToSocial IDs according
// to the receiving Pager's SinceID, MinID, MaxID and Limits.
// NOTE THE INPUT SLICE MUST BE SORTED IN ASCENDING ORDER.
// (I.E. NEWEST ITEMS AT LOWEST INDICES, OLDER AT HIGHER).
func (p *Pager) PageDesc(ids []string) []string {
	if p == nil {
		// no paging.
		return ids
	}

	var asc bool

	if p.MaxID != "" {
		for i := 0; i < len(ids); i++ {
			if ids[i] == p.MaxID {
				// Hit the boundary.
				// Reslice to be:
				// "from here"
				ids = ids[i+1:]
				break
			}
		}
	}

	if p.SinceID != "" {
		// If a sinceID is given, we
		// page down i.e. descending.
		asc = false

		for i := 0; i < len(ids); i++ {
			if ids[i] == p.SinceID {
				// Hit the boundary.
				// Reslice to be:
				// "up to here"
				ids = ids[:i]
				break
			}
		}
	} else if p.MinID != "" {
		// We only support minID if
		// no sinceID is provided.
		//
		// If a minID is given, we
		// page up, i.e. ascending.
		asc = true

		for i := 0; i < len(ids); i++ {
			if ids[i] == p.MinID {
				// Hit the boundary.
				// Reslice to be:
				// "up to here"
				ids = ids[:i]
				break
			}
		}
	}

	if asc && len(ids) > 1 {
		var (
			// Start at front.
			i = 0

			// Start at back.
			j = len(ids) - 1
		)

		// Clone input IDs before
		// we perform modifications.
		ids = slices.Clone(ids)

		for i < j {
			// Swap i,j index values in slice.
			ids[i], ids[j] = ids[j], ids[i]

			// incr + decr,
			// looping until
			// they meet in
			// the middle.
			i++
			j--
		}
	}

	if p.Limit > 0 && p.Limit < len(ids) {
		// Reslice IDs to given limit.
		ids = ids[:p.Limit]
	}

	return ids
}
