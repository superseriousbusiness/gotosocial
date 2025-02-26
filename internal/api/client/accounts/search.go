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

package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// AccountSearchGETHandler swagger:operation GET /api/v1/accounts/search accountSearchGet
//
// Search for accounts by username and/or display name.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: limit
//		type: integer
//		description: Number of results to try to return.
//		default: 40
//		maximum: 80
//		minimum: 1
//		in: query
//	-
//		name: offset
//		type: integer
//		description: >-
//			Page number of results to return (starts at 0).
//			This parameter is currently not used, offsets
//			over 0 will always return 0 results.
//		default: 0
//		maximum: 10
//		minimum: 0
//		in: query
//	-
//		name: q
//		type: string
//		description: |-
//			Query string to search for. This can be in the following forms:
//			- `@[username]` -- search for an account with the given username on any domain. Can return multiple results.
//			- `@[username]@[domain]` -- search for a remote account with exact username and domain. Will only ever return 1 result at most.
//			- any arbitrary string -- search for accounts containing the given string in their username or display name. Can return multiple results.
//		in: query
//		required: true
//	-
//		name: resolve
//		type: boolean
//		description: >-
//			If query is for `@[username]@[domain]`, or a URL, allow the GoToSocial instance to resolve
//			the search by making calls to remote instances (webfinger, ActivityPub, etc).
//		default: false
//		in: query
//	-
//		name: following
//		type: boolean
//		description: >-
//			Show only accounts that the requesting account follows. If this is set to `true`, then the GoToSocial instance
//			will enhance the search by also searching within account notes, not just in usernames and display names.
//		default: false
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			name: search results
//			description: Results of the search.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
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
func (m *Module) AccountSearchGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		// For moving/moved accounts, just return
		// empty to avoid breaking client apps.
		apiutil.Data(c, http.StatusOK, apiutil.AppJSON, apiutil.EmptyJSONArray)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	limit, errWithCode := apiutil.ParseLimit(c.Query(apiutil.LimitKey), 40, 80, 1)
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

	results, errWithCode := m.processor.Search().Accounts(
		c.Request.Context(),
		authed.Account,
		query,
		limit,
		offset,
		resolve,
		following,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, results)
}
