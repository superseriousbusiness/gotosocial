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

package lists

import (
	"errors"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// ListAccountsDELETEHandler swagger:operation DELETE /api/v1/lists/{id}/accounts removeListAccounts
//
// Remove one or more accounts from the given list.
//
//	---
//	tags:
//	- lists
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: ID of the list
//		in: path
//		required: true
//	-
//		name: account_ids[]
//		type: array
//		items:
//			type: string
//		description: >-
//			Array of accountIDs to modify.
//			Each accountID must correspond to an account
//			that the requesting account follows.
//		in: formData
//		collectionFormat: multi
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:lists
//
//	responses:
//		'200':
//			description: list accounts updated
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
func (m *Module) ListAccountsDELETEHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteLists,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	targetListID := c.Param(IDKey)
	if targetListID == "" {
		err := errors.New("no list id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.ListAccountsChangeRequest{}

	// XXX: Sending a body with a DELETE request is undefined. Ruby on Rails parses
	// it fine. Go's (*http.Request).ParseForm only parses POST-style forms for POST,
	// PUT, and PATCH request methods. Change the method until we're done with
	// parsing in order to be compatible with Mastodon's client API conventions.
	oldMethod := c.Request.Method
	c.Request.Method = "POST"
	err := c.ShouldBind(form)
	c.Request.Method = oldMethod

	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if len(form.AccountIDs) == 0 {
		err := errors.New("no account IDs given")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if errWithCode := m.processor.List().RemoveFromList(c.Request.Context(), authed.Account, targetListID, form.AccountIDs); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.Data(c, http.StatusOK, apiutil.AppJSON, apiutil.EmptyJSONObject)
}
