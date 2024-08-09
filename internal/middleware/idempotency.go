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
	"encoding/json"
	"net/http"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
)

// Idempotency returns a piece of gin middleware
// capable of handling the Idempotency-Key header:
// https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/
func Idempotency() gin.HandlerFunc {

	// Prepare response given when request already handled.
	alreadyHandled, err := json.Marshal(map[string]string{
		"status": "request already handled",
	})
	if err != nil {
		panic(err)
	}

	// Prepare expected error response JSON ahead of time.
	errorConflict, err := json.Marshal(map[string]string{
		"error": "request already under way",
	})
	if err != nil {
		panic(err)
	}

	// Prepare an idempotency responses cache for responses.
	responses := ttl.New[string, int](0, 1000, 5*time.Minute)
	if !responses.Start(time.Minute) {
		panic("failed to start idempotency cache")
	}

	return func(c *gin.Context) {
		// Ignore requests that don't provide
		// a body, i.e. generally will not be
		// updating any server resources.
		switch c.Request.Method {
		case "HEAD", "GET":
			c.Next()
			return
		}

		// Look for idempotency key provided in header.
		key := c.Request.Header.Get("Idempotency-Key")
		if key == "" {

			// When no key is
			// provided, just
			// skip the rest of
			// this middleware.
			c.Next()
			return
		}

		// Update the key we use to include general
		// request fingerprint along with idempotency
		// key to ensure uniqueness across logged-in
		// device sessions regardless of IP.
		//
		// NOTE: using the auth header is only an option
		// because we ONLY support bearer oauth tokens.
		key = c.Request.Header.Get("Authorization") +
			c.Request.UserAgent() +
			c.Request.Method +
			c.Request.URL.RequestURI() +
			key

		// Look for stored response.
		code, _ := responses.Get(key)
		switch code {

		// Not yet
		// handled.
		case 0:

		// Request is already
		// under way for key.
		case -1:
			apiutil.Data(c,
				http.StatusConflict,
				apiutil.AppJSON,
				errorConflict,
			)
			return

		// Already handled
		// this request.
		default:
			apiutil.Data(c,
				code,
				apiutil.AppJSON,
				alreadyHandled,
			)
			return
		}

		defer func() {
			if code := c.Writer.Status(); code != 0 {
				// Store response in map,
				// codes of zero indicate
				// a panic during handling.
				responses.Set(key, code)
			}
		}()

		// Pass on to next
		// handler in chain.
		c.Next()
	}
}
