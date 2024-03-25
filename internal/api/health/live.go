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

package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LiveGETRequest swagger:operation GET /livez liveGet
//
// Returns code 200 with no body if GoToSocial is "live", ie., able to respond to HTTP requests.
//
//	---
//	tags:
//	- health
//
//	responses:
//		'200':
//			description: OK
func (m *Module) LiveGETRequest(c *gin.Context) {
	c.Status(http.StatusOK)
}

// LiveHEADRequest swagger:operation HEAD /livez liveHead
//
// Returns code 200 if GoToSocial is "live", ie., able to respond to HTTP requests.
//
//	---
//	tags:
//	- health
//
//	responses:
//		'200':
//			description: OK
func (m *Module) LiveHEADRequest(c *gin.Context) {
	c.Status(http.StatusOK)
}
