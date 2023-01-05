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

package api

import (
	"context"
	"net/url"

	"github.com/gin-gonic/gin"
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

func (a *ActivityPub) Route(r router.Router, m ...gin.HandlerFunc) {
	// create groupings for the 'emoji' and 'users' prefixes
	emojiGroup := r.AttachGroup("emoji")
	usersGroup := r.AttachGroup("users")

	// instantiate + attach shared, non-global middlewares to both of these groups
	var (
		signatureCheckMiddleware = middleware.SignatureCheck(a.isURIBlocked)
		cacheControlMiddleware   = middleware.CacheControl("no-store")
	)
	emojiGroup.Use(m...)
	usersGroup.Use(m...)
	emojiGroup.Use(signatureCheckMiddleware, cacheControlMiddleware)
	usersGroup.Use(signatureCheckMiddleware, cacheControlMiddleware)

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
