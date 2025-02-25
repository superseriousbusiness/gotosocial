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
)

// HeaderFilterAllowDELETE swagger:operation DELETE /api/v1/admin/header_allows/{id} headerFilterAllowDelete
//
// Delete the "allow" header filter with the given ID.
//
//	---
//	tags:
//	- admin
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: Target header filter ID.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'202':
//			description: Accepted
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'500':
//			description: internal server error
func (m *Module) HeaderFilterAllowDELETE(c *gin.Context) {
	m.deleteHeaderFilter(c, m.processor.Admin().DeleteAllowHeaderFilter)
}

// HeaderFilterBlockDELETE swagger:operation DELETE /api/v1/admin/header_blocks/{id} headerFilterBlockDelete
//
// Delete the "block" header filter with the given ID.
//
//	---
//	tags:
//	- admin
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: Target header filter ID.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'202':
//			description: Accepted
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'500':
//			description: internal server error
func (m *Module) HeaderFilterBlockDELETE(c *gin.Context) {
	m.deleteHeaderFilter(c, m.processor.Admin().DeleteBlockHeaderFilter)
}
