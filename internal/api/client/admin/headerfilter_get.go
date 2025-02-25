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

import "github.com/gin-gonic/gin"

// HeaderFilterAllowGET swagger:operation GET /api/v1/admin/header_allows/{id} headerFilterAllowGet
//
// Get "allow" header filter with the given ID.
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
//		- admin:read
//
//	responses:
//		'200':
//			description: The requested "allow" header filter.
//			schema:
//				"$ref": "#/definitions/headerFilter"
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
func (m *Module) HeaderFilterAllowGET(c *gin.Context) {
	m.getHeaderFilter(c, m.processor.Admin().GetAllowHeaderFilter)
}

// HeaderFilterBlockGET swagger:operation GET /api/v1/admin/header_blocks/{id} headerFilterBlockGet
//
// Get "block" header filter with the given ID.
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
//		- admin:read
//
//	responses:
//		'200':
//			description: The requested "block" header filter.
//			schema:
//				"$ref": "#/definitions/headerFilter"
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
func (m *Module) HeaderFilterBlockGET(c *gin.Context) {
	m.getHeaderFilter(c, m.processor.Admin().GetBlockHeaderFilter)
}

// HeaderFilterAllowsGET swagger:operation GET /api/v1/admin/header_allows headerFilterAllowsGet
//
// Get all "allow" header filters currently in place.
//
//	---
//	tags:
//	- admin
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: All "allow" header filters currently in place.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/headerFilter"
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
func (m *Module) HeaderFilterAllowsGET(c *gin.Context) {
	m.getHeaderFilters(c, m.processor.Admin().GetAllowHeaderFilters)
}

// HeaderFilterBlocksGET swagger:operation GET /api/v1/admin/header_blocks headerFilterBlocksGet
//
// Get all "allow" header filters currently in place.
//
//	---
//	tags:
//	- admin
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: All "block" header filters currently in place.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/headerFilter"
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
func (m *Module) HeaderFilterBlocksGET(c *gin.Context) {
	m.getHeaderFilters(c, m.processor.Admin().GetBlockHeaderFilters)
}
