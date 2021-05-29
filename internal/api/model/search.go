/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package model

// SearchQuery corresponds to search parameters as submitted through the client API.
// See https://docs.joinmastodon.org/methods/search/
type SearchQuery struct {
	// If provided, statuses returned will be authored only by this account
	AccountID string
	// Return results older than this id
	MaxID string
	// Return results immediately newer than this id
	MinID string
	// Enum(accounts, hashtags, statuses)
	Type string
	// Filter out unreviewed tags? Defaults to false. Use true when trying to find trending tags.
	ExcludeUnreviewed bool
	// The search query
	Query string
	// Attempt WebFinger lookup. Defaults to false.
	Resolve bool
	// Maximum number of results to load, per type. Defaults to 20. Max 40.
	Limit int
	// Offset in search results. Used for pagination. Defaults to 0.
	Offset int
	// Only include accounts that the user is following. Defaults to false.
	Following bool
}

// SearchResult corresponds to a search result, containing accounts, statuses, and hashtags.
// See https://docs.joinmastodon.org/methods/search/
type SearchResult struct {
	Accounts []Account `json:"accounts"`
	Statuses []Status  `json:"statuses"`
	Hashtags []Tag     `json:"hashtags"`
}
