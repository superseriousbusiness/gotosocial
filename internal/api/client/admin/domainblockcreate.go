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
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// DomainBlocksPOSTHandler swagger:operation POST /api/v1/admin/domain_blocks domainBlockCreate
//
// Create one or more domain blocks, from a string or a file.
//
// You have two options when using this endpoint: either you can set `import` to `true` and
// upload a file containing multiple domain blocks, JSON-formatted, or you can leave import as
// `false`, and just add one domain block.
//
// The format of the json file should be something like: `[{"domain":"example.org"},{"domain":"whatever.com","public_comment":"they smell"}]`
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
//		name: import
//		in: query
//		description: >-
//			Signal that a list of domain blocks is being imported as a file.
//			If set to `true`, then 'domains' must be present as a JSON-formatted file.
//			If set to `false`, then `domains` will be ignored, and `domain` must be present.
//		type: boolean
//		default: false
//	-
//		name: domains
//		in: formData
//		description: >-
//			JSON-formatted list of domain blocks to import.
//			This is only used if `import` is set to `true`.
//		type: file
//	-
//		name: domain
//		in: formData
//		description: >-
//			Single domain to block.
//			Used only if `import` is not `true`.
//		type: string
//	-
//		name: obfuscate
//		in: formData
//		description: >-
//			Obfuscate the name of the domain when serving it publicly.
//			Eg., `example.org` becomes something like `ex***e.org`.
//			Used only if `import` is not `true`.
//		type: boolean
//	-
//		name: public_comment
//		in: formData
//		description: >-
//			Public comment about this domain block.
//			This will be displayed alongside the domain block if you choose to share blocks.
//			Used only if `import` is not `true`.
//		type: string
//	-
//		name: private_comment
//		in: formData
//		description: >-
//			Private comment about this domain block. Will only be shown to other admins, so this
//			is a useful way of internally keeping track of why a certain domain ended up blocked.
//			Used only if `import` is not `true`.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:domain_blocks
//
//	responses:
//		'200':
//			description: >-
//				The newly created domain block, if `import` != `true`.
//				If a list has been imported, then an `array` of newly created domain blocks will be returned instead.
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
func (m *Module) DomainBlocksPOSTHandler(c *gin.Context) {
	m.createDomainPermissions(c,
		gtsmodel.DomainPermissionBlock,
		m.processor.Admin().DomainPermissionCreate,
		m.processor.Admin().DomainPermissionsImport,
	)
}
