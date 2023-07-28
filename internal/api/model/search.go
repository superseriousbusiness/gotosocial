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

package model

// SearchRequest models a search request.
type SearchRequest struct {
	MaxID             string
	MinID             string
	Limit             int
	Offset            int
	Query             string
	QueryType         string
	Resolve           bool
	Following         bool
	ExcludeUnreviewed bool
	APIv1             bool // Set to 'true' if using version 1 of the search API.
}

// SearchResult models a search result.
//
// swagger:model searchResult
type SearchResult struct {
	Accounts []*Account `json:"accounts"`
	Statuses []*Status  `json:"statuses"`
	// Slice of strings if api v1, slice of tags if api v2.
	Hashtags []any `json:"hashtags"`
}
