/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

// SearchQuery models a search request.
//
// swagger:parameters searchGet
type SearchQuery struct {
	// If type is `statuses`, then statuses returned will be authored only by this account.
	//
	// in: query
	AccountID string `json:"account_id"`
	// Return results *older* than this id.
	//
	// The entry with this ID will not be included in the search results.
	// in: query
	MaxID string `json:"max_id"`
	// Return results *newer* than this id.
	//
	// The entry with this ID will not be included in the search results.
	// in: query
	MinID string `json:"min_id"`
	// Type of the search query to perform.
	//
	// Must be one of: `accounts`, `hashtags`, `statuses`.
	//
	// enum:
	// - accounts
	// - hashtags
	// - statuses
	// required: true
	// in: query
	Type string `json:"type"`
	// Filter out tags that haven't been reviewed and approved by an instance admin.
	//
	// default: false
	// in: query
	ExcludeUnreviewed bool `json:"exclude_unreviewed"`
	// String to use as a search query.
	//
	// For accounts, this should be in the format `@someaccount@some.instance.com`, or the format `https://some.instance.com/@someaccount`
	//
	// For a status, this can be in the format: `https://some.instance.com/@someaccount/SOME_ID_OF_A_STATUS`
	//
	// required: true
	// in: query
	Query string `json:"q"`
	// Attempt to resolve the query by performing a remote webfinger lookup, if the query includes a remote host.
	// default: false
	Resolve bool `json:"resolve"`
	// Maximum number of results to load, per type.
	// default: 20
	// minimum: 1
	// maximum: 40
	// in: query
	Limit int `json:"limit"`
	// Offset for paginating search results.
	//
	// default: 0
	// in: query
	Offset int `json:"offset"`
	// Only include accounts that the searching account is following.
	// default: false
	// in: query
	Following bool `json:"following"`
}

// SearchResult models a search result.
//
// swagger:model searchResult
type SearchResult struct {
	Accounts []Account `json:"accounts"`
	Statuses []Status  `json:"statuses"`
	Hashtags []Tag     `json:"hashtags"`
}
