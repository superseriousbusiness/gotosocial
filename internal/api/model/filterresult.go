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

// FilterResult is returned along with a filtered status to explain why it was filtered.
//
// swagger:model filterResult
//
// ---
// tags:
// - filters
type FilterResult struct {
	// The filter that was matched.
	Filter FilterV2 `json:"filter"`
	// The keywords within the filter that were matched.
	KeywordMatches []string `json:"keyword_matches"`
	// The status IDs within the filter that were matched.
	StatusMatches []string `json:"status_matches"`
}
