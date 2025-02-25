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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// AccountRelationshipsGETHandler swagger:operation GET /api/v1/accounts/relationships accountRelationships
//
// See your account's relationships with the given account IDs.
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
//		name: id[]
//		type: array
//		items:
//			type: string
//		description: Account IDs.
//		in: query
//		collectionFormat: multi
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			name: account relationships
//			description: Array of account relationships.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/accountRelationship"
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
func (m *Module) AccountRelationshipsGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	targetAccountIDs := c.QueryArray("id[]")
	if len(targetAccountIDs) == 0 {
		// check fallback -- let's be generous and see if maybe it's just set as 'id'?
		id := c.Query("id")
		if id == "" {
			err := errors.New("no account id(s) specified in query")
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
			return
		}
		targetAccountIDs = append(targetAccountIDs, id)
	}

	relationships := []apimodel.Relationship{}

	for _, targetAccountID := range targetAccountIDs {
		r, errWithCode := m.processor.Account().RelationshipGet(c.Request.Context(), authed.Account, targetAccountID)
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}
		relationships = append(relationships, *r)
	}

	apiutil.JSON(c, http.StatusOK, relationships)
}
