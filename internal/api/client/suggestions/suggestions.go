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

package suggestions

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

const (
	BasePath = "/v2/suggestions"
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

// SuggestionsGETHandler swagger:operation GET /api/v1/suggestions getSuggestions
//
// Accounts that are promoted by staff, or that the user has had past positive interactions with, but is not yet following.
//
// THIS ENDPOINT IS CURRENTLY NOT FULLY IMPLEMENTED: it will always return an empty array.
//
//	---
//	tags:
//	- suggestions
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read
//
//	responses:
//		'200':
//			schema:
//				type: array
//				items:
//					type: object
//				maxItems: 0
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
func (m *Module) SuggestionsGETHandler(c *gin.Context) {
	_, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeRead,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiutil.EmptyJSONArray)
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.SuggestionsGETHandler)
}
