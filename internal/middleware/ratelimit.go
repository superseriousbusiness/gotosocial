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
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

const rateLimitPeriod = 5 * time.Minute

// RateLimit returns a gin middleware that will automatically rate limit caller (by IP address),
// and enrich the response header with the following headers:
//
//   - `X-Ratelimit-Limit`     - maximum number of requests allowed per time period (fixed).
//   - `X-Ratelimit-Remaining` - number of remaining requests that can still be performed.
//   - `X-Ratelimit-Reset`     - ISO8601 timestamp when the rate limit will reset.
//
// If `X-Ratelimit-Limit` is exceeded, the request is aborted and an HTTP 429 TooManyRequests
// status is returned.
//
// If the config AdvancedRateLimitRequests value is <= 0, then a noop handler will be returned,
// which performs no rate limiting.
func RateLimit(limit int) gin.HandlerFunc {
	if limit <= 0 {
		// Rate limiting is disabled.
		// Return noop middleware.
		return func(ctx *gin.Context) {}
	}

	limiter := limiter.New(
		memory.NewStore(),
		limiter.Rate{
			Period: rateLimitPeriod,
			Limit:  int64(limit),
		},
	)

	var (
		exceptionsStrs  = config.GetAdvancedRateLimitExceptions()
		exceptionsLen   = len(exceptionsStrs)
		checkExceptions = (exceptionsLen != 0)
		exceptions      []netip.Prefix
	)

	if checkExceptions {
		// Parse rate limit exception CIDRs, if set.
		exceptions = make([]netip.Prefix, len(exceptionsStrs))
		for i, e := range exceptionsStrs {
			p, err := netip.ParsePrefix(e)
			if err != nil {
				log.Panicf(nil, "could not parse rate limit exception %s as netip.Prefix: %q", e, err)
			}

			exceptions[i] = p
		}
	}

	// It's prettymuch impossible to effectively
	// rate limit the immense IPv6 address space
	// unless we mask some of the bytes.
	ipv6Mask := net.CIDRMask(64, 128)

	return func(c *gin.Context) {
		// Use Gin's heuristic for determining
		// clientIP, which accounts for reverse
		// proxies and trusted proxies setting.
		clientIP := net.ParseIP(c.ClientIP())

		if checkExceptions {
			// Check if this IP is exempted
			// from being rate limited.
			for _, exception := range exceptions {
				if exception.Contains(clientIP) {
					// Allow the request
					// to continue.
					c.Next()
					return
				}
			}
		}

		// If clientIP is IPv6, mask it.
		if len(clientIP) == net.IPv6len {
			clientIP = clientIP.Mask(ipv6Mask)
		}

		// Fetch rate limit info for this clientIP.
		context, err := limiter.Get(c, clientIP.String())
		if err != nil {
			// Since we use an in-memory cache now,
			// it's actually impossible for this to
			// error, but handle it nicely anyway in
			// case we switch implementation in future.
			errWithCode := gtserror.NewErrorInternalError(err)

			// Set error on gin context so it'll
			// be picked up by logging middleware.
			c.Error(errWithCode)

			// Bail with 500.
			c.AbortWithStatusJSON(
				errWithCode.Code(),
				gin.H{"error": errWithCode.Safe()},
			)
			return
		}

		// Provide reset in same format used by
		// Mastodon. There's no real standard as
		// to what format X-RateLimit-Reset SHOULD
		// use, but since most clients interacting
		// with us will expect the Mastodon version,
		// it makes sense to take this.
		resetT := time.Unix(context.Reset, 0)
		reset := util.FormatISO8601(resetT)

		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", reset)

		if context.Reached {
			// Return JSON error message for
			// consistency with other endpoints.
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{"error": "rate limit reached"},
			)
			return
		}

		// Allow the request
		// to continue.
		c.Next()
	}
}
