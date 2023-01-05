/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
	The code in this file is adapted from MIT-licensed code in github.com/go-chi/chi. Thanks chi (thi)!

	See: https://github.com/go-chi/chi/blob/e6baba61759b26ddf7b14d1e02d1da81a4d76c08/middleware/throttle.go

	And: https://github.com/sponsors/pkieltyka
*/

package middleware

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	errCapacityExceeded = "server capacity exceeded"
	errTimedOut         = "timed out while waiting for a pending request to complete"
	errContextCanceled  = "context canceled"
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
// Useful links:
//
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503
func Throttle(cpuMultiplier int) gin.HandlerFunc {
	if cpuMultiplier <= 0 {
		// throttling is disabled, return a noop middleware
		return func(c *gin.Context) {}
	}

	var (
		limit              = runtime.GOMAXPROCS(0) * cpuMultiplier
		backlogLimit       = limit * cpuMultiplier
		backlogChannelSize = limit + backlogLimit
		tokens             = make(chan token, limit)
		backlogTokens      = make(chan token, backlogChannelSize)
		retryAfter         = "30" // seconds
		backlogDuration    = 30 * time.Second
	)

	// prefill token channels
	for i := 0; i < limit; i++ {
		tokens <- token{}
	}

	for i := 0; i < backlogChannelSize; i++ {
		backlogTokens <- token{}
	}

	// bail instructs the requester to return after retryAfter seconds, returns a 503,
	// and writes the given message into the "error" field of a returned json object
	bail := func(c *gin.Context, msg string) {
		c.Header("Retry-After", retryAfter)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": msg})
		c.Abort()
	}

	return func(c *gin.Context) {
		// inside this select, the caller tries to get a backlog token
		select {
		case <-c.Request.Context().Done():
			// request context has been canceled already
			bail(c, errContextCanceled)
		case btok := <-backlogTokens:
			// take a backlog token and wait
			timer := time.NewTimer(backlogDuration)
			defer func() {
				// when we're finished, return the backlog token to the bucket
				backlogTokens <- btok
			}()

			// inside *this* select, the caller has a backlog token,
			// and they're waiting for their turn to be processed
			select {
			case <-timer.C:
				// waiting too long in the backlog
				bail(c, errTimedOut)
			case <-c.Request.Context().Done():
				// the request context has been canceled already
				timer.Stop()
				bail(c, errContextCanceled)
			case tok := <-tokens:
				// the caller gets a token, so their request can now be processed
				timer.Stop()
				defer func() {
					// whatever happens to the request, put the
					// token back in the bucket when we're finished
					tokens <- tok
				}()
				c.Next() // <- finally process the caller's request
			}

		default:
			// we don't have space in the backlog queue
			bail(c, errCapacityExceeded)
		}
	}
}
