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

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

func (m *Module) ready(c *gin.Context) {
	if err := m.readyF(c.Request.Context()); err != nil {
		// Set error on the gin context so
		// it's logged by the logging middleware.
		errWithCode := gtserror.NewErrorInternalError(err)
		c.Error(errWithCode) //nolint:errcheck
		c.Status(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

// ReadyGETRequest swagger:operation GET /readyz readyGet
//
// Returns code 200 with no body if GoToSocial is "ready", ie., able to connect to the database backend and do a simple SELECT.
//
// If GtS is not ready, 500 Internal Error will be returned, and an error will be logged (but not returned to the caller, to avoid leaking internals).
//
//	---
//	tags:
//	- health
//
//	responses:
//		'200':
//			description: OK
//		'500':
//			description: Not ready. Check logs for error message.
func (m *Module) ReadyGETRequest(c *gin.Context) {
	m.ready(c)
}

// ReadyHEADRequest swagger:operation HEAD /readyz readyHead
//
// Returns code 200 with no body if GoToSocial is "ready", ie., able to connect to the database backend and do a simple SELECT.
//
// If GtS is not ready, 500 Internal Error will be returned, and an error will be logged (but not returned to the caller, to avoid leaking internals).
//
//	---
//	tags:
//	- health
//
//	responses:
//		'200':
//			description: OK
func (m *Module) ReadyHEADRequest(c *gin.Context) {
	m.ready(c)
}
