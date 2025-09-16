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

package middleware

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/gin-gonic/gin"
)

// UserAgentOrTeapot returns a gin middleware
// which aborts requests with empty user agent
// strings, returning code 418 - I'm a teapot.
//
// If `instance-allow-empty-user-agents` is
// true, it logs a debug msg instead of aborting.
func UserAgentOrTeapot() gin.HandlerFunc {

	// Build variables outside the handler
	// so they're not instantiated every
	// time a request is processed.
	var (
		rsp         = []byte(`{"error": "I'm a teapot: no user-agent sent with request"}`)
		allowEmpty = config.GetInstanceAllowEmptyUserAgents()
	)

	return func(c *gin.Context) {
		ua := c.Request.UserAgent()
		if ua != "" {
			// All good.
			return
		}

		if allowEmpty {
			// No user-agent was
			// set but that's OK.
			log.Debugf(
				c.Request.Context(),
				"allowing request with empty User-Agent from client %s",
				c.ClientIP(),
			)
			return
		}

		// No user-agent set and that's not ok!
		//
		// Give them a taste of the ol' teapot.
		apiutil.Data(c, http.StatusTeapot, apiutil.AppJSON, rsp)
		c.Abort()
	}
}
