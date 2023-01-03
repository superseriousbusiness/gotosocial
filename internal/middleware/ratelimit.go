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

package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	limitergin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

const rateLimitPeriod = 5 * time.Minute

// RateLimit returns a gin middleware that will automatically rate limit caller (by IP address),
// and enrich the response header with the following headers:
//
//   - `x-ratelimit-limit`     - maximum number of requests allowed per time period (fixed).
//   - `x-ratelimit-remaining` - number of remaining requests that can still be performed.
//   - `x-ratelimit-reset`     - unix timestamp when the rate limit will reset.
//
// If `x-ratelimit-limit` is exceeded, the request is aborted and an HTTP 429 TooManyRequests
// status is returned.
//
// If the config AdvancedRateLimitRequests value is <= 0, then a noop handler will be returned,
// which performs no rate limiting.
func RateLimit(limit int) gin.HandlerFunc {
	if limit <= 0 {
		// use noop middleware if ratelimiting is disabled
		return func(ctx *gin.Context) {}
	}

	limiter := limiter.New(
		memory.NewStore(),
		limiter.Rate{Period: rateLimitPeriod, Limit: int64(limit)},
		limiter.WithIPv6Mask(net.CIDRMask(64, 128)), // apply /64 mask to IPv6 addresses
	)

	// use custom rate limit reached error
	handler := func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit reached"})
	}

	return limitergin.NewMiddleware(
		limiter,
		limitergin.WithLimitReachedHandler(handler),
	)
}
