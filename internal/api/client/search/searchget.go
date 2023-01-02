/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package search

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// SearchGETHandler swagger:operation GET /api/v1/search searchGet
//
// Search for statuses, accounts, or hashtags, on this instance or elsewhere.
//
// If statuses are in the result, they will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
//	---
//	tags:
//	- search
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
//				type: array
//				items:
//					"$ref": "#/definitions/searchResult"
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
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	excludeUnreviewed := false
	excludeUnreviewedString := c.Query(ExcludeUnreviewedKey)
	if excludeUnreviewedString != "" {
		var err error
		excludeUnreviewed, err = strconv.ParseBool(excludeUnreviewedString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", ExcludeUnreviewedKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
	}

	query := c.Query(QueryKey)
	if query == "" {
		err := errors.New("query parameter q was empty")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	resolve := false
	resolveString := c.Query(ResolveKey)
	if resolveString != "" {
		var err error
		resolve, err = strconv.ParseBool(resolveString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", ResolveKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
	}

	limit := 2
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 32)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		limit = int(i)
	}
	if limit > 40 {
		limit = 40
	}
	if limit < 1 {
		limit = 1
	}

	offset := 0
	offsetString := c.Query(OffsetKey)
	if offsetString != "" {
		i, err := strconv.ParseInt(offsetString, 10, 32)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", OffsetKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		offset = int(i)
	}

	following := false
	followingString := c.Query(FollowingKey)
	if followingString != "" {
		var err error
		following, err = strconv.ParseBool(followingString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", FollowingKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
	}

	searchQuery := &apimodel.SearchQuery{
		AccountID:         c.Query(AccountIDKey),
		MaxID:             c.Query(MaxIDKey),
		MinID:             c.Query(MinIDKey),
		Type:              c.Query(TypeKey),
		ExcludeUnreviewed: excludeUnreviewed,
		Query:             query,
		Resolve:           resolve,
		Limit:             limit,
		Offset:            offset,
		Following:         following,
	}

	results, errWithCode := m.processor.SearchGet(c.Request.Context(), authed, searchQuery)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, results)
}
