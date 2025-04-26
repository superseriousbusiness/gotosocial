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

// DomainBlockUpdatePUTHandler swagger:operation PUT /api/v1/admin/domain_blocks/{id} domainBlockUpdate
//
// Update a single domain block.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
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
//	-
//		name: obfuscate
//		in: formData
//		description: >-
//			Obfuscate the name of the domain when serving it publicly.
//			Eg., `example.org` becomes something like `ex***e.org`.
//		type: boolean
//	-
//		name: public_comment
//		in: formData
//		description: >-
//			Public comment about this domain block.
//			This will be displayed alongside the domain block if you choose to share blocks.
//		type: string
//	-
//		name: private_comment
//		in: formData
//		description: >-
//			Private comment about this domain block. Will only be shown to other admins, so this
//			is a useful way of internally keeping track of why a certain domain ended up blocked.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:domain_blocks
//
//	responses:
//		'200':
//			description: The updated domain block.
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
//		'500':
//			description: internal server error
func (m *Module) DomainBlockUpdatePUTHandler(c *gin.Context) {
	m.updateDomainPermission(c, gtsmodel.DomainPermissionBlock)
}
