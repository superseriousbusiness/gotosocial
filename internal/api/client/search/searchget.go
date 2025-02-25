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

package search

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// SearchGETHandler swagger:operation GET /api/{api_version}/search searchGet
//
// Search for statuses, accounts, or hashtags, on this instance or elsewhere.
//
// If statuses are in the result, they will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
//	---
//	tags:
//	- search
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: api_version
//		type: string
//		in: path
//		description: >-
//			Version of the API to use. Must be either `v1` or `v2`.
//			If v1 is used, Hashtag results will be a slice of strings.
//			If v2 is used, Hashtag results will be a slice of apimodel tags.
//		required: true
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items *OLDER* than the given max ID.
//			The item with the specified ID will not be included in the response.
//			Currently only used if 'type' is set to a specific type.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items *immediately newer* than the given min ID.
//			The item with the specified ID will not be included in the response.
//			Currently only used if 'type' is set to a specific type.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of each type of item to return.
//		default: 20
//		maximum: 40
//		minimum: 1
//		in: query
//		required: false
//	-
//		name: offset
//		type: integer
//		description: >-
//			Page number of results to return (starts at 0).
//			This parameter is currently not used, page by selecting
//			a specific query type and using maxID and minID instead.
//		default: 0
//		maximum: 10
//		minimum: 0
//		in: query
//		required: false
//	-
//		name: q
//		type: string
//		description: |-
//			Query string to search for. This can be in the following forms:
//			- `@[username]` -- search for an account with the given username on any domain. Can return multiple results.
//			- @[username]@[domain]` -- search for a remote account with exact username and domain. Will only ever return 1 result at most.
//			- `https://example.org/some/arbitrary/url` -- search for an account OR a status with the given URL. Will only ever return 1 result at most.
//			- `#[hashtag_name]` -- search for a hashtag with the given hashtag name, or starting with the given hashtag name. Case insensitive. Can return multiple results.
//			- any arbitrary string -- search for accounts or statuses containing the given string. Can return multiple results.
//
//			Arbitrary string queries may include the following operators:
//			- `from:localuser`, `from:remoteuser@instance.tld`: restrict results to statuses created by the specified account.
//		in: query
//		required: true
//	-
//		name: type
//		type: string
//		description: |-
//			Type of item to return. One of:
//			- `` -- empty string; return any/all results.
//			- `accounts` -- return only account(s).
//			- `statuses` -- return only status(es).
//			- `hashtags` -- return only hashtag(s).
//			If `type` is specified, paging can be performed using max_id and min_id parameters.
//			If `type` is not specified, see the `offset` parameter for paging.
//		in: query
//	-
//		name: resolve
//		type: boolean
//		description: >-
//			If searching query is for `@[username]@[domain]`, or a URL, allow the GoToSocial
//			instance to resolve the search by making calls to remote instances (webfinger, ActivityPub, etc).
//		default: false
//		in: query
//	-
//		name: following
//		type: boolean
//		description: >-
//			If search type includes accounts, and search query is an arbitrary string, show only accounts
//			that the requesting account follows. If this is set to `true`, then the GoToSocial instance will
//			enhance the search by also searching within account notes, not just in usernames and display names.
//		default: false
//		in: query
//	-
//		name: exclude_unreviewed
//		type: boolean
//		description: >-
//			If searching for hashtags, exclude those not yet approved by instance admin.
//			Currently this parameter is unused.
//		default: false
//		in: query
//	-
//		name: account_id
//		type: string
//		description: >-
//			Restrict results to statuses created by the specified account.
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- read:search
//
//	responses:
//		'200':
//			name: search results
//			description: Results of the search.
//			schema:
//				"$ref": "#/definitions/searchResult"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) SearchGETHandler(c *gin.Context) {
	apiVersion, errWithCode := apiutil.ParseAPIVersion(
		c.Param(apiutil.APIVersionKey),
		[]string{apiutil.APIv1, apiutil.APIv2}...,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadSearch,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		// For moving/moved accounts, just return
		// empty to avoid breaking client apps.
		results := &apimodel.SearchResult{
			Accounts: make([]*apimodel.Account, 0),
			Statuses: make([]*apimodel.Status, 0),
			Hashtags: make([]any, 0),
		}
		apiutil.JSON(c, http.StatusOK, results)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	limit, errWithCode := apiutil.ParseLimit(c.Query(apiutil.LimitKey), 20, 40, 1)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	offset, errWithCode := apiutil.ParseSearchOffset(c.Query(apiutil.SearchOffsetKey), 0, 10, 0)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	query, errWithCode := apiutil.ParseSearchQuery(c.Query(apiutil.SearchQueryKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resolve, errWithCode := apiutil.ParseSearchResolve(c.Query(apiutil.SearchResolveKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	following, errWithCode := apiutil.ParseSearchFollowing(c.Query(apiutil.SearchFollowingKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	excludeUnreviewed, errWithCode := apiutil.ParseSearchExcludeUnreviewed(c.Query(apiutil.SearchExcludeUnreviewedKey), false)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	searchRequest := &apimodel.SearchRequest{
		MaxID:             c.Query(apiutil.MaxIDKey),
		MinID:             c.Query(apiutil.MinIDKey),
		Limit:             limit,
		Offset:            offset,
		Query:             query,
		QueryType:         c.Query(apiutil.SearchTypeKey),
		Resolve:           resolve,
		Following:         following,
		ExcludeUnreviewed: excludeUnreviewed,
		AccountID:         c.Query(apiutil.AccountIDKey),
		APIv1:             apiVersion == apiutil.APIv1,
	}

	results, errWithCode := m.processor.Search().Get(c.Request.Context(), authed.Account, searchRequest)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, results)
}
