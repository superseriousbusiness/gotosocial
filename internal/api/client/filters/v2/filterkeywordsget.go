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

package v2

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// FilterKeywordsGETHandler swagger:operation GET /api/v2/filters/{id}/keywords filterKeywordsGet
//
// Get all filter keywords for a given filter.
//
//	---
//	tags:
//	- filters
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: ID of the filter
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:filters
//
//	responses:
//		'200':
//			name: filterKeywords
//			description: Requested filter keywords.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/filterKeyword"
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
func (m *Module) FilterKeywordsGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadFilters,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	filterID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiFilter, errWithCode := m.processor.FiltersV2().KeywordsGetForFilterID(c.Request.Context(), authed.Account, filterID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, apiFilter)
}
