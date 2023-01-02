/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"strconv"
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

// ThrottleOpts passes options to the Throttle middleware.
type ThrottleOpts struct {
	Limit                 int // how many requests can be in-process at once
	BacklogLimit          int // how many requests can be queued + waiting
	BacklogTimeoutSeconds int // how long to queue requests before timing them out
	RetryAfterSeconds     int //
}

// Throttle returns a gin middleware that performs throttling of incoming requests,
// ensuring that only t.limit
//
// Callers will first attempt to get a backlog token. Once they have that, they will
// wait in the backlog queue until they can get a token to allow their request to be
// processed.
//
// If the backlog queue is full, the request context is closed, or the caller has been
// waiting in the backlog for too long, this function will abort the request chain,
// write a JSON error into the response, set an appropriate Retry-After value, and set
// the HTTP response code to 429: Too Many Requests.
//
// If BacklogLimit is < 0, or any of the other opt values are < 1, then this function
// will instead return a noop handler.
//
// Useful links:
//
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429
func Throttle(t ThrottleOpts) gin.HandlerFunc {
	if t.BacklogLimit < 0 || t.Limit < 1 || t.BacklogTimeoutSeconds < 1 || t.RetryAfterSeconds < 1 {
		// throttling is disabled, return a noop middleware
		return func(c *gin.Context) {}
	}

	var (
		tokens          = make(chan token, t.Limit)
		backlogTokens   = make(chan token, t.Limit+t.BacklogLimit)
		retryAfter      = strconv.Itoa(t.RetryAfterSeconds)
		backlogDuration = time.Duration(t.BacklogTimeoutSeconds) * time.Second
	)

	// prefill token buckets
	for i := 0; i < t.Limit+t.BacklogLimit; i++ {
		if i < t.Limit {
			tokens <- token{}
		}
		backlogTokens <- token{}
	}

	// bail instructs the requester to return after retryAfter seconds, returns a 429,
	// and writes the given message into the "error" field of a returned json object
	bail := func(c *gin.Context, msg string) {
		c.Header("Retry-After", retryAfter)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": msg})
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
