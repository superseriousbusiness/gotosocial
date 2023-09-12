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
	switch {
	case p.Min.Order != 0:
		return p.Min.Order
	case p.Max.Order != 0:
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

	if p.order().Ascending() {
		// Sort type is ascending, input
		// data is assumed to be ascending.

		if minIdx := p.Min.Find(in); minIdx != -1 {
			// Reslice skipping up to min.
			in = in[minIdx+1:]
		}

		if maxIdx := p.Max.Find(in); maxIdx != -1 {
			// Reslice stripping past max.
			in = in[:maxIdx]
		}

		if p.Limit > 0 && p.Limit < len(in) {
			// Reslice input to limit.
			in = in[:p.Limit]
		}

		if len(in) > 1 {
			// Clone input before
			// any modifications.
			in = slices.Clone(in)

			// Output slice must
			// ALWAYS be descending.
			in = Reverse(in)
		}
	} else {
		// Default sort is descending,
		// catching all cases when NOT
		// ascending (even zero value).

		if maxIdx := p.Max.Find(in); maxIdx != -1 {
			// Reslice skipping up to max.
			in = in[maxIdx+1:]
		}

		if minIdx := p.Min.Find(in); minIdx != -1 {
			// Reslice stripping past min.
			in = in[:minIdx]
		}

		if p.Limit > 0 && p.Limit < len(in) {
			// Reslice input to limit.
			in = in[:p.Limit]
		}
	}

	return in
}

// Next creates a new instance for the next returnable page, using
// given max value. This preserves original limit and max key name.
func (p *Page) Next(lo, hi string) *Page {
	if p == nil || lo == "" || hi == "" {
		// no paging.
		return nil
	}

	// Create new page.
	p2 := new(Page)

	// Set original limit.
	p2.Limit = p.Limit

	if p.order().Ascending() {
		// When ascending, next page
		// needs to start with min at
		// the next highest value.
		p2.Min = p.Min.new(hi)
		p2.Max = p.Max.new("")
	} else {
		// When descending, next page
		// needs to start with max at
		// the next lowest value.
		p2.Min = p.Min.new("")
		p2.Max = p.Max.new(lo)
	}

	return p2
}

// Prev creates a new instance for the prev returnable page, using
// given min value. This preserves original limit and min key name.
func (p *Page) Prev(lo, hi string) *Page {
	if p == nil || lo == "" || hi == "" {
		// no paging.
		return nil
	}

	// Create new page.
	p2 := new(Page)

	// Set original limit.
	p2.Limit = p.Limit

	if p.order().Ascending() {
		// When ascending, prev page
		// needs to start with max at
		// the next lowest value.
		p2.Min = p.Min.new("")
		p2.Max = p.Max.new(lo)
	} else {
		// When descending, next page
		// needs to start with max at
		// the next lowest value.
		p2.Min = p.Min.new(hi)
		p2.Max = p.Max.new("")
	}

	return p2
}

// ToLink performs ToLinkURL() and calls .String() on the resulting URL.
func (p *Page) ToLink(proto, host, path string, queryParams url.Values) string {
	u := p.ToLinkURL(proto, host, path, queryParams)
	if u == nil {
		return ""
	}
	return u.String()
}

// ToLink builds a URL link for given endpoint information and extra query parameters,
// appending this Page's minimum / maximum boundaries and available limit (if any).
func (p *Page) ToLinkURL(proto, host, path string, queryParams url.Values) *url.URL {
	if p == nil {
		// no paging.
		return nil
	}

	if queryParams == nil {
		// Allocate new query parameters.
		queryParams = make(url.Values)
	}

	if p.Min.Value != "" {
		// A page-minimum query parameter is available.
		queryParams.Add(p.Min.Name, p.Min.Value)
	}

	if p.Max.Value != "" {
		// A page-maximum query parameter is available.
		queryParams.Add(p.Max.Name, p.Max.Value)
	}

	if p.Limit > 0 {
		// A page limit query parameter is available.
		queryParams.Add("limit", strconv.Itoa(p.Limit))
	}

	// Build URL string.
	return &url.URL{
		Scheme:   proto,
		Host:     host,
		Path:     path,
		RawQuery: queryParams.Encode(),
	}
}
