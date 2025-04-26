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
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/gin-gonic/gin"
)

// DomainBlockDELETEHandler swagger:operation DELETE /api/v1/admin/domain_blocks/{id} domainBlockDelete
//
// Delete domain block with the given ID.
//
//	---
//	tags:
//	- admin
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: The id of the domain block.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:domain_blocks
//
//	responses:
//		'200':
//			description: The domain block that was just deleted.
//			schema:
//				"$ref": "#/definitions/domainPermission"
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
//		'409':
//			description: >-
//				Conflict: There is already an admin action running that conflicts with this action.
//				Check the error message in the response body for more information. This is a temporary
//				error; it should be possible to process this action if you try again in a bit.
//		'500':
//			description: internal server error
func (m *Module) DomainBlockDELETEHandler(c *gin.Context) {
	m.deleteDomainPermission(c, gtsmodel.DomainPermissionBlock)
}
