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

/*
	The code in this file is adapted from MIT-licensed code in github.com/go-chi/chi. Thanks chi (thi)!

	See: https://github.com/go-chi/chi/blob/e6baba61759b26ddf7b14d1e02d1da81a4d76c08/middleware/throttle.go

	And: https://github.com/sponsors/pkieltyka
*/

package middleware

import (
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
)

// token represents a request that is being processed.
type token struct{}

// Throttle returns a gin middleware that performs throttling of incoming requests,
// ensuring that only a certain number of requests are handled concurrently, to reduce
// congestion of the server.
//
// Limits are configured using available CPUs and the given cpuMultiplier value.
// Open request limit is available CPUs * multiplier; backlog limit is limit * multiplier.
//
// Example values for multiplier 8:
//
//	1 cpu = 08 open, 064 backlog
//	2 cpu = 16 open, 128 backlog
//	4 cpu = 32 open, 256 backlog
//
// Example values for multiplier 4:
//
//	1 cpu = 04 open, 016 backlog
//	2 cpu = 08 open, 032 backlog
//	4 cpu = 16 open, 064 backlog
//
// Callers will first attempt to get a backlog token. Once they have that, they will
// wait in the backlog queue until they can get a token to allow their request to be
// processed.
//
// If the backlog queue is full, the request context is closed, or the caller has been
// waiting in the backlog for too long, this function will abort the request chain,
// write a JSON error into the response, set an appropriate Retry-After value, and set
// the HTTP response code to 503: Service Unavailable.
//
// If the multiplier is <= 0, a noop middleware will be returned instead.
//
// RetryAfter determines the Retry-After header value to be sent to throttled requests.
//
// Useful links:
//
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503
func Throttle(cpuMultiplier int, retryAfter time.Duration) gin.HandlerFunc {
	if cpuMultiplier <= 0 {
		// throttling is disabled, return a noop middleware
		return func(c *gin.Context) {}
	}

	if retryAfter < 0 {
		retryAfter = 0
	}

	var (
		limit         = runtime.GOMAXPROCS(0) * cpuMultiplier
		queueLimit    = limit * cpuMultiplier
		tokens        = make(chan token, limit)
		requestCount  = atomic.Int64{}
		retryAfterStr = strconv.FormatUint(uint64(retryAfter/time.Second), 10) // #nosec G115 -- Checked right above
	)

	// prefill token channel
	for i := 0; i < limit; i++ {
		tokens <- token{}
	}

	return func(c *gin.Context) {
		// Always decrement request counter.
		defer func() { requestCount.Add(-1) }()

		// Increment request count.
		n := requestCount.Add(1)

		// Check whether the request
		// count is over queue limit.
		if n > int64(queueLimit) {
			c.Header("Retry-After", retryAfterStr)
			apiutil.Data(c,
				http.StatusTooManyRequests,
				apiutil.AppJSON,
				apiutil.ErrorCapacityExceeded,
			)
			c.Abort()
			return
		}

		// Sit and wait in the
		// queue for free token.
		select {

		case <-c.Request.Context().Done():
			// request context has
			// been canceled already.
			return

		case tok := <-tokens:
			// caller has successfully
			// received a token, allowing
			// request to be processed.

			defer func() {
				// when we're finished, return
				// this token to the bucket.
				tokens <- tok
			}()

			// Process
			// request!
			c.Next()
		}
	}
}
