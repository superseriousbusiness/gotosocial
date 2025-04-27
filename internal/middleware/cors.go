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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a new gin middleware which allows CORS requests to be processed.
// This is necessary in order for web/browser-based clients like Semaphore to work.
func CORS() gin.HandlerFunc {
	cfg := cors.Config{
		// todo: use config to customize this
		AllowAllOrigins: true,

		// adds the following:
		// 	"chrome-extension://"
		// 	"safari-extension://"
		// 	"moz-extension://"
		// 	"ms-browser-extension://"
		AllowBrowserExtensions: true,
		AllowMethods: []string{
			"POST",
			"PUT",
			"DELETE",
			"GET",
			"PATCH",
			"OPTIONS",
		},
		AllowHeaders: []string{
			// basic cors stuff
			"Origin",
			"Content-Length",
			"Content-Type",

			// needed to pass oauth bearer tokens
			"Authorization",

			// Some clients require this; see:
			//   - https://docs.joinmastodon.org/methods/statuses/#headers
			//   - https://codeberg.org/superseriousbusiness/gotosocial/issues/1664
			"Idempotency-Key",

			// needed for websocket upgrade requests
			"Upgrade",
			"Sec-WebSocket-Extensions",
			"Sec-WebSocket-Key",
			"Sec-WebSocket-Protocol",
			"Sec-WebSocket-Version",
			"Connection",
		},
		AllowWebSockets: true,
		ExposeHeaders: []string{
			// needed for accessing next/prev links when making GET timeline requests
			"Link",

			// needed so clients can handle rate limits
			"X-RateLimit-Reset",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-Request-Id",

			// websocket stuff
			"Connection",
			"Sec-WebSocket-Accept",
			"Upgrade",
		},
		MaxAge: 2 * time.Minute,
	}

	return cors.New(cfg)
}
