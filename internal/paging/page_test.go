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

package paging_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/oklog/ulid"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"golang.org/x/exp/slices"
)

// random reader according to current-time source seed.
var randRd = rand.New(rand.NewSource(time.Now().Unix()))

type Case struct {
	// Name is the test case name.
	Name string

	// Page to use for test.
	Page *paging.Page

	// Input contains test case input ID slice.
	Input []string

	// Expect contains expected test case output.
	Expect []string
}

// CreateCase creates a new test case with random input for function defining test page parameters and expected output.
func CreateCase(name string, getParams func([]string) (input []string, page *paging.Page, expect []string)) Case {
	i := randRd.Intn(100)
	in := generateSlice(i)
	input, page, expect := getParams(in)
	return Case{
		Name:   name,
		Page:   page,
		Input:  input,
		Expect: expect,
	}
}

func TestPage(t *testing.T) {
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Page the input slice.
			out := c.Page.Page(c.Input)

			// Log the results for case of error returns.
			t.Logf("\ninput=%v\noutput=%v\nexpected=%v", c.Input, out, c.Expect)

			// Check paged output is as expected.
			if !slices.Equal(out, c.Expect) {
				t.Error("unexpected paged output")
			}
		})
	}
}

var cases = []Case{
	CreateCase("minID and maxID set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted ascending for min_id
		slices.SortFunc(ids, func(a, b string) bool {
			return b < a
		})

		// Select random indices in slice.
		minIdx := randRd.Intn(len(ids))
		maxIdx := randRd.Intn(len(ids))

		// Select the boundaries.
		minID := ids[minIdx]
		maxID := ids[maxIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutLower(expect, minID)
		expect = cutUpper(expect, maxID)
		expect = paging.Reverse(expect)

		// Return page and expected IDs.
		return ids, &paging.Page{
			Min: paging.MinID(minID, ""),
			Max: paging.MaxID(maxID),
		}, expect
	}),
	CreateCase("minID, maxID and limit set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted ascending for min_id
		slices.SortFunc(ids, func(a, b string) bool {
			return b < a
		})

		// Select random parameters in slice.
		minIdx := randRd.Intn(len(ids))
		maxIdx := randRd.Intn(len(ids))
		limit := randRd.Intn(len(ids))

		// Select the boundaries.
		minID := ids[minIdx]
		maxID := ids[maxIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutLower(expect, minID)
		expect = cutUpper(expect, maxID)
		expect = paging.Reverse(expect)

		// Now limit the slice.
		if limit < len(expect) {
			expect = expect[:limit]
		}

		// Return page and expected IDs.
		return ids, &paging.Page{
			Min:   paging.MinID(minID, ""),
			Max:   paging.MaxID(maxID),
			Limit: limit,
		}, expect
	}),
	CreateCase("minID, maxID and too-large limit set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted ascending for min_id
		slices.SortFunc(ids, func(a, b string) bool {
			return b < a
		})

		// Select random parameters in slice.
		minIdx := randRd.Intn(len(ids))
		maxIdx := randRd.Intn(len(ids))

		// Select the boundaries.
		minID := ids[minIdx]
		maxID := ids[maxIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutLower(expect, minID)
		expect = cutUpper(expect, maxID)
		expect = paging.Reverse(expect)

		// Return page and expected IDs.
		return ids, &paging.Page{
			Min:   paging.MinID(minID, ""),
			Max:   paging.MaxID(maxID),
			Limit: len(ids) * 2,
		}, expect
	}),
	CreateCase("sinceID and maxID set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted descending for since_id
		slices.SortFunc(ids, func(a, b string) bool {
			return a < b
		})

		// Select random indices in slice.
		sinceIdx := randRd.Intn(len(ids))
		maxIdx := randRd.Intn(len(ids))

		// Select the boundaries.
		sinceID := ids[sinceIdx]
		maxID := ids[maxIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutLower(expect, maxID)
		expect = cutUpper(expect, sinceID)

		// Return page and expected IDs.
		return ids, &paging.Page{
			Min: paging.MinID("", sinceID),
			Max: paging.MaxID(maxID),
		}, expect
	}),
	CreateCase("maxID set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted descending for since_id
		slices.SortFunc(ids, func(a, b string) bool {
			return a < b
		})

		// Select random indices in slice.
		maxIdx := randRd.Intn(len(ids))

		// Select the boundaries.
		maxID := ids[maxIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutLower(expect, maxID)

		// Return page and expected IDs.
		return ids, &paging.Page{
			Max: paging.MaxID(maxID),
		}, expect
	}),
	CreateCase("sinceID set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted descending for since_id
		slices.SortFunc(ids, func(a, b string) bool {
			return a < b
		})

		// Select random indices in slice.
		sinceIdx := randRd.Intn(len(ids))

		// Select the boundaries.
		sinceID := ids[sinceIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutUpper(expect, sinceID)

		// Return page and expected IDs.
		return ids, &paging.Page{
			Min: paging.MinID("", sinceID),
		}, expect
	}),
	CreateCase("minID set", func(ids []string) ([]string, *paging.Page, []string) {
		// Ensure input slice sorted ascending for min_id
		slices.SortFunc(ids, func(a, b string) bool {
			return b < a
		})

		// Select random indices in slice.
		minIdx := randRd.Intn(len(ids))

		// Select the boundaries.
		minID := ids[minIdx]

		// Create expected output.
		expect := slices.Clone(ids)
		expect = cutLower(expect, minID)
		expect = paging.Reverse(expect)

		// Return page and expected IDs.
		return ids, &paging.Page{
			Min: paging.MinID(minID, ""),
		}, expect
	}),
}

// cutLower cuts off the lower part of the slice from `bound` downwards.
func cutLower(in []string, bound string) []string {
	for i := 0; i < len(in); i++ {
		if in[i] == bound {
			return in[i+1:]
		}
	}
	return in
}

// cutUpper cuts off the upper part of the slice from `bound` onwards.
func cutUpper(in []string, bound string) []string {
	for i := 0; i < len(in); i++ {
		if in[i] == bound {
			return in[:i]
		}
	}
	return in
}

// generateSlice generates a new slice of len containing ascending sorted slice.
func generateSlice(len int) []string {
	if len <= 0 {
		// minimum testable
		// pageable amount
		len = 2
	}
	now := time.Now()
	in := make([]string, len)
	for i := 0; i < len; i++ {
		// Convert now to timestamp.
		t := ulid.Timestamp(now)

		// Create anew ulid for now.
		u := ulid.MustNew(t, randRd)

		// Add to slice.
		in[i] = u.String()

		// Bump now by 1 second.
		now = now.Add(time.Second)
	}
	return in
}
