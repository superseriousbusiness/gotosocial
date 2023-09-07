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

import (
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Page struct {
	// Min is the Page's lower limit value.
	Min Boundary

	// Max is this Page's upper limit value.
	Max Boundary

	// Limit will limit the returned
	// page of items to at most 'limit'.
	Limit int
}

// GetMin is a small helper function to return minimum boundary value (checking for nil page).
func (p *Page) GetMin() string {
	if p == nil {
		return ""
	}
	return p.Min.Value
}

// GetMax is a small helper function to return maximum boundary value (checking for nil page).
func (p *Page) GetMax() string {
	if p == nil {
		return ""
	}
	return p.Max.Value
}

// GetLimit is a small helper function to return limit (checking for nil page and unusable limit).
func (p *Page) GetLimit() int {
	if p == nil || p.Limit < 0 {
		return 0
	}
	return p.Limit
}

// GetOrder is a small helper function to return page sort ordering (checking for nil page).
func (p *Page) GetOrder() Order {
	if p == nil {
		return 0
	}
	return p.order()
}

func (p *Page) order() Order {
	var (
		// Check if min/max values set.
		minValue = zero(p.Min.Value)
		maxValue = zero(p.Max.Value)

		// Check if min/max orders set.
		minOrder = (p.Min.Order != 0)
		maxOrder = (p.Max.Order != 0)
	)

	switch {
	// Boundaries with a value AND order set
	// take priority. Min always comes first.
	case minValue && minOrder:
		return p.Min.Order
	case maxValue && maxOrder:
		return p.Max.Order
	case minOrder:
		return p.Min.Order
	case maxOrder:
		return p.Max.Order
	default:
		return 0
	}
}

// Page will page the given slice of input according
// to the receiving Page's minimum, maximum and limit.
// NOTE: input slice MUST be sorted according to the order is
// expected to be paged in, i.e. it is currently sorted
// according to Page.Order(). Sorted data isn't always according
// to string inequalities so this CANNOT be checked here.
func (p *Page) Page(in []string) []string {
	if p == nil {
		// no paging.
		return in
	}

	if o := p.order(); !o.Ascending() {
		// Default sort is descending,
		// catching all cases when NOT
		// ascending (even zero value).
		//
		// NOTE: sorted data does not always
		// occur according to string ineqs
		// so we unfortunately cannot check.

		if maxIdx := p.Max.Find(in); maxIdx != -1 {
			// Reslice skipping up to max.
			in = in[maxIdx+1:]
		}

		if minIdx := p.Min.Find(in); minIdx != -1 {
			// Reslice stripping past min.
			in = in[:minIdx]
		}
	} else {
		// Sort type is ascending, input
		// data is assumed to be ascending.
		//
		// NOTE: sorted data does not always
		// occur according to string ineqs
		// so we unfortunately cannot check.

		if minIdx := p.Min.Find(in); minIdx != -1 {
			// Reslice skipping up to min.
			in = in[minIdx+1:]
		}

		if maxIdx := p.Max.Find(in); maxIdx != -1 {
			// Reslice stripping past max.
			in = in[:maxIdx]
		}

		if len(in) > 1 {
			var (
				// Output slice must
				// ALWAYS be descending.

				// Start at front.
				i = 0

				// Start at back.
				j = len(in) - 1
			)

			// Clone input before
			// any modifications.
			in = slices.Clone(in)

			for i < j {
				// Swap i,j index values in slice.
				in[i], in[j] = in[j], in[i]

				// incr + decr,
				// looping until
				// they meet in
				// the middle.
				i++
				j--
			}
		}
	}

	if p.Limit > 0 && p.Limit < len(in) {
		// Reslice input to limit.
		in = in[:p.Limit]
	}

	return in
}

// Next creates a new instance for the next returnable page, using
// given max value. This preserves original limit and max key name.
func (p *Page) Next(max string) *Page {
	if p == nil || max == "" {
		// no paging.
		return nil
	}

	// Create new page.
	p2 := new(Page)

	// Set original limit.
	p2.Limit = p.Limit

	// Create new from old.
	p2.Max = p.Max.new(max)

	return p2
}

// Prev creates a new instance for the prev returnable page, using
// given min value. This preserves original limit and min key name.
func (p *Page) Prev(min string) *Page {
	if p == nil || min == "" {
		// no paging.
		return nil
	}

	// Create new page.
	p2 := new(Page)

	// Set original limit.
	p2.Limit = p.Limit

	// Create new from old.
	p2.Min = p.Min.new(min)

	return p2
}

// ToLink builds a URL link for given endpoint information and extra query parameters,
// appending this Page's minimum / maximum boundaries and available limit (if any).
func (p *Page) ToLink(proto, host, path string, queryParams []string) string {
	if p == nil {
		// no paging.
		return ""
	}

	// Check length before
	// adding boundary params.
	old := len(queryParams)

	if minParam := p.Min.Query(); minParam != "" {
		// A page-minimum query parameter is available.
		queryParams = append(queryParams, minParam)
	}

	if maxParam := p.Max.Query(); maxParam != "" {
		// A page-maximum query parameter is available.
		queryParams = append(queryParams, maxParam)
	}

	if len(queryParams) == old {
		// No page boundaries.
		return ""
	}

	if p.Limit > 0 {
		// Build limit key-value query parameter.
		param := "limit=" + strconv.Itoa(p.Limit)

		// Append `limit=$value` query parameter.
		queryParams = append(queryParams, param)
	}

	// Join collected params into query str.
	query := strings.Join(queryParams, "&")

	// Build URL string.
	return (&url.URL{
		Scheme:   proto,
		Host:     host,
		Path:     path,
		RawQuery: query,
	}).String()
}
