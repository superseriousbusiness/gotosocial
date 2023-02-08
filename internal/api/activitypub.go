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
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/emoji"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/publickey"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/users"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type ActivityPub struct {
	emoji                    *emoji.Module
	users                    *users.Module
	publicKey                *publickey.Module
	signatureCheckMiddleware gin.HandlerFunc
}

func (a *ActivityPub) Route(r router.Router, m ...gin.HandlerFunc) {
	// create groupings for the 'emoji' and 'users' prefixes
	emojiGroup := r.AttachGroup("emoji")
	usersGroup := r.AttachGroup("users")

	// attach shared, non-global middlewares to both of these groups
	cacheControlMiddleware := middleware.CacheControl("no-store")
	emojiGroup.Use(m...)
	usersGroup.Use(m...)
	emojiGroup.Use(a.signatureCheckMiddleware, cacheControlMiddleware)
	usersGroup.Use(a.signatureCheckMiddleware, cacheControlMiddleware)

	a.emoji.Route(emojiGroup.Handle)
	a.users.Route(usersGroup.Handle)
}

// Public key endpoint requires different middleware + cache policies from other AP endpoints.
func (a *ActivityPub) RoutePublicKey(r router.Router, m ...gin.HandlerFunc) {
	publicKeyGroup := r.AttachGroup(publickey.PublicKeyPath)
	publicKeyGroup.Use(a.signatureCheckMiddleware, middleware.CacheControl("public,max-age=604800"))
	a.publicKey.Route(publicKeyGroup.Handle)
}

func NewActivityPub(db db.DB, p processing.Processor) *ActivityPub {
	return &ActivityPub{
		emoji:                    emoji.New(p),
		users:                    users.New(p),
		publicKey:                publickey.New(p),
		signatureCheckMiddleware: middleware.SignatureCheck(db.IsURIBlocked),
	}
}
