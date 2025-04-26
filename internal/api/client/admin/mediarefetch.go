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

package admin

import (
	"fmt"
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// MediaRefetchPOSTHandler swagger:operation POST /api/v1/admin/media_refetch mediaRefetch
//
// Refetch media specified in the database but missing from storage.
// Currently, this only includes remote emojis.
// This endpoint is useful when data loss has occurred, and you want to try to recover to a working state.
//
//	---
//	tags:
//	- admin
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	parameters:
//	-
//		name: domain
//		in: query
//		description: >-
//			Domain to refetch media from.
//			If empty, all domains will be refetched.
//		type: string
//
//	responses:
//		'202':
//			description: >-
//				Request accepted and will be processed.
//				Check the logs for progress / errors.
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) MediaRefetchPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminWrite,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if errWithCode := m.processor.Admin().MediaRefetch(c.Request.Context(), authed.Account, c.Query(DomainQueryKey)); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.Data(c, http.StatusOK, apiutil.AppJSON, apiutil.StatusAcceptedJSON)
}
