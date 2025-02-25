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

// AccountAliasPOSTHandler swagger:operation POST /api/v1/accounts/alias accountAlias
//
// Alias your account to another account by setting alsoKnownAs to the given URI.
//
// This is useful when you want to move from another account this this account.
//
// In such cases, you should set the alsoKnownAs of this account to the URI of
// the account you want to move from.
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
//		name: also_known_as_uris
//		in: formData
//		description: >-
//			ActivityPub URI/IDs of target accounts to which this account
//			is being aliased. Eg., `["https://example.org/users/some_account"]`.
//
//			Use an empty array to unset alsoKnownAs, clearing the aliases.
//		type: string
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- write:accounts
//
//	responses:
//		'200':
//			description: "The newly updated account."
//			schema:
//				"$ref": "#/definitions/account"
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
func (m *Module) AccountAliasPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.AccountAliasRequest{}
	if err := c.ShouldBind(&form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Account().Alias(c.Request.Context(), authed.Account, form.AlsoKnownAsURIs)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}
