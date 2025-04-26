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

// DomainBlocksGETHandler swagger:operation GET /api/v1/admin/domain_blocks domainBlocksGet
//
// View all domain blocks currently in place.
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
//		name: export
//		type: boolean
//		description: >-
//			If set to `true`, then each entry in the returned list of domain blocks will only consist of
//			the fields `domain` and `public_comment`. This is perfect for when you want to save and share
//			a list of all the domains you have blocked on your instance, so that someone else can easily import them,
//			but you don't want them to see the database IDs of your blocks, or private comments etc.
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:domain_blocks
//
//	responses:
//		'200':
//			description: All domain blocks currently in place.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/domainPermission"
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
func (m *Module) DomainBlocksGETHandler(c *gin.Context) {
	m.getDomainPermissions(c, gtsmodel.DomainPermissionBlock)
}
