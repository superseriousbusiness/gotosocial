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

// HeaderFilterAllowPOST swagger:operation POST /api/v1/admin/header_allows headerFilterAllowCreate
//
// Create new "allow" HTTP request header filter.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'200':
//			description: The newly created "allow" header filter.
//			schema:
//				"$ref": "#/definitions/headerFilter"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'500':
//			description: internal server error
func (m *Module) HeaderFilterAllowPOST(c *gin.Context) {
	m.createHeaderFilter(c, m.processor.Admin().CreateAllowHeaderFilter)
}

// HeaderFilterBlockPOST swagger:operation POST /api/v1/admin/header_blocks headerFilterBlockCreate
//
// Create new "block" HTTP request header filter.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'200':
//			description: The newly created "block" header filter.
//			schema:
//				"$ref": "#/definitions/headerFilter"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'500':
//			description: internal server error
func (m *Module) HeaderFilterBlockPOST(c *gin.Context) {
	m.createHeaderFilter(c, m.processor.Admin().CreateBlockHeaderFilter)
}
