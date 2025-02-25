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
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// AccountMovePOSTHandler swagger:operation POST /api/v1/accounts/move accountMove
//
// Move your account to another account.
//
//	---
//	tags:
//	- accounts
//
//	consumes:
//	- multipart/form-data
//
//	parameters:
//	-
//		name: password
//		in: formData
//		description: Password of the account user, for confirmation.
//		type: string
//		required: true
//	-
//		name: moved_to_uri
//		in: formData
//		description: >-
//			ActivityPub URI/ID of the target account. Eg., `https://example.org/users/some_account`.
//			The target account must be alsoKnownAs the requesting account in order for the move to be successful.
//		type: string
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- write:accounts
//
//	responses:
//		'202':
//			description: The account move has been accepted and the account will be moved.
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'422':
//			description: Unprocessable. Check the response body for more details.
//		'500':
//			description: internal server error
func (m *Module) AccountMovePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.AccountMoveRequest{}
	if err := c.ShouldBind(&form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if errWithCode := m.processor.Account().MoveSelf(c.Request.Context(), authed, form); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusAccepted, map[string]string{
		"message": "accepted",
	})
}
