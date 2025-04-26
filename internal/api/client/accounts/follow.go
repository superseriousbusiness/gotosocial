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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// AccountFollowPOSTHandler swagger:operation POST /api/v1/accounts/{id}/follow accountFollow
//
// Follow account with id.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
// If you already follow (request) the given account, then the follow (request) will be updated instead using the
// `reblogs` and `notify` parameters.
//
//	---
//	tags:
//	- accounts
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	parameters:
//	-
//		name: id
//		required: true
//		in: path
//		description: ID of the account to follow.
//		type: string
//	-
//		name: reblogs
//		type: boolean
//		default: true
//		description: Show reblogs from this account.
//		in: formData
//	-
//		name: notify
//		type: boolean
//		default: false
//		description: Notify when this account posts.
//		in: formData
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- write:follows
//
//	responses:
//		'200':
//			name: account relationship
//			description: Your relationship to this account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
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
func (m *Module) AccountFollowPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteFollows,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		err := errors.New("no account id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.AccountFollowRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}
	form.ID = targetAcctID

	relationship, errWithCode := m.processor.Account().FollowCreate(c.Request.Context(), authed.Account, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, relationship)
}
