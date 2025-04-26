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

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
)

const rateLimitPeriod = 5 * time.Minute

// RateLimit returns a gin middleware that will automatically rate
// limit caller (by IP address), and enrich the response header with
// the following headers:
//
//   - `X-Ratelimit-Limit`     - max requests allowed per time period (fixed).
//   - `X-Ratelimit-Remaining` - requests remaining for this IP before reset.
//   - `X-Ratelimit-Reset`     - ISO8601 timestamp when the rate limit will reset.
//
// If `X-Ratelimit-Limit` is exceeded, the request is aborted and an
// HTTP 429 TooManyRequests status is returned.
//
// If the config AdvancedRateLimitRequests value is <= 0, then a noop
// handler will be returned, which performs no rate limiting.
func RateLimit(limit int, except []netip.Prefix) gin.HandlerFunc {
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

	// It's prettymuch impossible to effectively
	// rate limit the immense IPv6 address space
	// unless we mask some of the bytes.
	//
	// This mask is pretty coarse, and puts IPv6
	// blocking on more or less the same footing
	// as IPv4 blocking in terms of how likely it
	// is to prevent abuse while still allowing
	// legit users access to the service.
	ipv6Mask := net.CIDRMask(64, 128)

	return func(c *gin.Context) {
		// Use Gin's heuristic for determining
		// clientIP, which accounts for reverse
		// proxies and trusted proxies setting.
		clientIP := netip.MustParseAddr(c.ClientIP())

		// Check if this IP is exempt from rate
		// limits and skip further checks if so.
		for _, prefix := range except {
			if prefix.Contains(clientIP) {
				c.Next()
				return
			}
		}

		if clientIP.Is6() {
			// Convert to "net" package IP for mask.
			asIP := net.IP(clientIP.AsSlice())

			// Apply coarse IPv6 mask.
			asIP = asIP.Mask(ipv6Mask)

			// Convert back to netip.Addr from net.IP.
			clientIP, _ = netip.AddrFromSlice(asIP)
		}

		// Fetch rate limit info for this (masked) clientIP.
		context, err := limiter.Get(c, clientIP.String())
		if err != nil {
			// Since we use an in-memory cache now,
			// it's actually impossible for this to
			// error, but handle it nicely anyway in
			// case we switch implementation in future.
			errWithCode := gtserror.NewErrorInternalError(err)

			// Set error on gin context so it'll
			// be picked up by logging middleware.
			c.Error(errWithCode) //nolint:errcheck

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
			apiutil.Data(c,
				http.StatusTooManyRequests,
				apiutil.AppJSON,
				apiutil.ErrorRateLimited,
			)
			c.Abort()
			return
		}

		// Allow the request
		// to continue.
		c.Next()
	}
}
