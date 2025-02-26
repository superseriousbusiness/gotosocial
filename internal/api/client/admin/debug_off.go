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

//go:build !debug && !debugenv
// +build !debug,!debugenv

package admin

import (
	"github.com/gin-gonic/gin"
)

// #######################################################
// # goswagger is generated using empty / off debug by   #
// # default, so put all the swagger documentation here! #
// #######################################################

// DebugAPUrlHandler swagger:operation GET /api/v1/admin/debug/apurl debugAPUrl
//
// Perform a GET to the specified ActivityPub URL and return detailed debugging information.
//
// Only enabled / exposed if GoToSocial was built and is running with flag DEBUG=1.
//
//	---
//	tags:
//	- debug
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: url
//		type: string
//		description: >-
//			The URL / ActivityPub ID to dereference.
//			This should be a full URL, including protocol.
//			Eg., `https://example.org/users/someone`
//		in: query
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'200':
//			name: Debug response.
//			schema:
//				"$ref": "#/definitions/debugAPUrlResponse"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) DebugAPUrlHandler(c *gin.Context) {}

// DebugClearCachesHandler swagger:operation POST /api/v1/admin/debug/caches/clear debugClearCaches
//
// Sweep/clear all in-memory caches.
//
// Only enabled / exposed if GoToSocial was built and is running with flag DEBUG=1.
//
//	---
//	tags:
//	- debug
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
//			description: All good baby!
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) DebugClearCachesHandler(c *gin.Context) {}
