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

package api

import (
	"context"
	"net/url"
	"runtime"

	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/emoji"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/users"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type ActivityPub struct {
	emoji *emoji.Module
	users *users.Module

	isURIBlocked func(context.Context, *url.URL) (bool, db.Error)
}

func (a *ActivityPub) Route(r router.Router) {
	// create groupings for the 'emoji' and 'users' prefixes
	emojiGroup := r.AttachGroup("emoji")
	usersGroup := r.AttachGroup("users")

	// configure throttle limits according to CPU numbers;
	//
	// example values:
	// 1 cpu = 08 open, 064 backlog
	// 2 cpu = 16 open, 128 backlog
	// 4 cpu = 32 open, 256 backlog
	maxProcs := runtime.GOMAXPROCS(0)
	limit := 8 * maxProcs     // allow eight concurrent open connections per cpu
	backlogLimit := 8 * limit // allow eight times that in backlog
	backlogTimeoutSeconds := 30
	retryAfterSeconds := backlogTimeoutSeconds

	throttleOpts := middleware.ThrottleOpts{
		Limit:                 limit,
		BacklogLimit:          backlogLimit,
		BacklogTimeoutSeconds: backlogTimeoutSeconds,
		RetryAfterSeconds:     retryAfterSeconds,
	}

	// instantiate + attach shared, non-global middlewares to both of these groups
	var (
		throttlingMiddleware     = middleware.Throttle(throttleOpts)
		rateLimitMiddleware      = middleware.RateLimit() // nolint:contextcheck
		signatureCheckMiddleware = middleware.SignatureCheck(a.isURIBlocked)
		gzipMiddleware           = middleware.Gzip()
		cacheControlMiddleware   = middleware.CacheControl("no-store")
	)
	emojiGroup.Use(throttlingMiddleware, rateLimitMiddleware, signatureCheckMiddleware, gzipMiddleware, cacheControlMiddleware)
	usersGroup.Use(throttlingMiddleware, rateLimitMiddleware, signatureCheckMiddleware, gzipMiddleware, cacheControlMiddleware)

	a.emoji.Route(emojiGroup.Handle)
	a.users.Route(usersGroup.Handle)
}

func NewActivityPub(db db.DB, p processing.Processor) *ActivityPub {
	return &ActivityPub{
		emoji: emoji.New(p),
		users: users.New(p),

		isURIBlocked: db.IsURIBlocked,
	}
}
