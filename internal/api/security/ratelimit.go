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

package security

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

type RateLimitOptions struct {
	Period time.Duration
	Limit  int64
}

func (m *Module) LimitReachedHandler(c *gin.Context) {
	code := http.StatusTooManyRequests
	c.AbortWithStatusJSON(code, gin.H{"error": "rate limit reached"})
}

// returns a gin middleware that will automatically rate limit caller (by IP address)
// and enrich the response header with the following headers:
// - `x-ratelimit-limit` maximum number of requests allowed per time period (fixed)
// - `x-ratelimit-remaining` number of remaining requests that can still be performed
// - `x-ratelimit-reset` unix timestamp when the rate limit will reset
// if `x-ratelimit-limit` is exceeded an HTTP 429 error is returned
func (m *Module) RateLimit(rateOptions RateLimitOptions) func(c *gin.Context) {
	rate := limiter.Rate{
		Period: rateOptions.Period,
		Limit:  rateOptions.Limit,
	}

	store := memory.NewStore()

	limiterInstance := limiter.New(store, rate)

	middleware := mgin.NewMiddleware(
		limiterInstance,
		// use custom rate limit reached error
		mgin.WithLimitReachedHandler(m.LimitReachedHandler),
	)

	return middleware
}
