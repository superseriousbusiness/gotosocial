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

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"github.com/gin-gonic/gin"
)

func TestThrottlingMiddleware(t *testing.T) {
	testThrottlingMiddleware(t, 2, time.Second*10)
	testThrottlingMiddleware(t, 4, time.Second*15)
	testThrottlingMiddleware(t, 8, time.Second*30)
}

func testThrottlingMiddleware(t *testing.T, cpuMulti int, retryAfter time.Duration) {
	// Calculate expected request limit + queue.
	limit := runtime.GOMAXPROCS(0) * cpuMulti
	queueLimit := limit * cpuMulti

	// Calculate expected retry-after header string.
	retryAfterStr := strconv.FormatUint(uint64(retryAfter/time.Second), 10)

	// Gin test http engine
	// (used for ctx init).
	e := gin.New()

	// Add middleware to the gin engine handler stack.
	middleware := middleware.Throttle(cpuMulti, retryAfter)
	e.Use(middleware)

	// Set the blocking gin handler.
	handler := blockingHandler()
	e.Handle("GET", "/", handler)

	var cncls []func()

	for i := 0; i < queueLimit+limit; i++ {
		// Prepare a gin test context.
		r := httptest.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()

		// Wrap request with new cancel context.
		ctx, cncl := context.WithCancel(r.Context())
		r = r.WithContext(ctx)

		// Pass req through
		// engine handler.
		go e.ServeHTTP(rw, r)
		time.Sleep(time.Millisecond)

		// Get http result.
		res := rw.Result()

		if i < queueLimit {

			// Check status == 200 (default, i.e not set).
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status code was set (%d) with queueLimit=%d and request=%d", res.StatusCode, queueLimit, i)
			}

			// Add cancel to func slice.
			cncls = append(cncls, cncl)

		} else {

			// Check the returned status code is expected.
			if res.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("did not return status 429 (%d) with queueLimit=%d and request=%d", res.StatusCode, queueLimit, i)
			}

			// Check the returned retry-after header is set.
			if res.Header.Get("Retry-After") != retryAfterStr {
				t.Fatalf("did not return retry-after %s with queueLimit=%d and request=%d", retryAfterStr, queueLimit, i)
			}

			// Cancel on return.
			defer cncl()

		}
	}

	// Cancel all blocked reqs.
	for _, cncl := range cncls {
		cncl()
	}
	time.Sleep(time.Second)

	// Check a bunchh more requests
	// can now make it through after
	// previous requests were released!
	for i := 0; i < limit; i++ {

		// Prepare a gin test context.
		r := httptest.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()

		// Pass req through
		// engine handler.
		go e.ServeHTTP(rw, r)
		time.Sleep(5 * time.Millisecond)

		// Get http result.
		res := rw.Result()

		// Check status == 200 (default, i.e not set).
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status code was set (%d) with queueLimit=%d and request=%d", res.StatusCode, queueLimit, i)
		}
	}
}

func blockingHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		<-ctx.Done()
		ctx.Status(201) // specifically not 200
	}
}
