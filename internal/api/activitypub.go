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

package api

import (
	"code.superseriousbusiness.org/gotosocial/internal/api/activitypub/emoji"
	"code.superseriousbusiness.org/gotosocial/internal/api/activitypub/publickey"
	"code.superseriousbusiness.org/gotosocial/internal/api/activitypub/users"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"github.com/gin-gonic/gin"
)

type ActivityPub struct {
	emoji                    *emoji.Module
	users                    *users.Module
	publicKey                *publickey.Module
	signatureCheckMiddleware gin.HandlerFunc
}

func (a *ActivityPub) Route(r *router.Router, m ...gin.HandlerFunc) {
	// create groupings for the 'emoji' and 'users' prefixes
	emojiGroup := r.AttachGroup("emoji")
	usersGroup := r.AttachGroup("users")

	// attach shared, non-global middlewares to both of these groups
	ccMiddleware := middleware.CacheControl(middleware.CacheControlConfig{
		Directives: []string{"no-store"},
	})
	emojiGroup.Use(m...)
	usersGroup.Use(m...)
	emojiGroup.Use(a.signatureCheckMiddleware, ccMiddleware)
	usersGroup.Use(a.signatureCheckMiddleware, ccMiddleware)

	a.emoji.Route(emojiGroup.Handle)
	a.users.Route(usersGroup.Handle)
}

// Public key endpoint requires different middleware + cache policies from other AP endpoints.
func (a *ActivityPub) RoutePublicKey(r *router.Router, m ...gin.HandlerFunc) {
	// Create grouping for the 'users/[username]/main-key' prefix.
	publicKeyGroup := r.AttachGroup(publickey.PublicKeyPath)

	// Attach middleware allowing public cacheing of main-key.
	ccMiddleware := middleware.CacheControl(middleware.CacheControlConfig{
		Directives: []string{"public", "max-age=604800"},
		Vary:       []string{"Accept", "Accept-Encoding"},
	})
	publicKeyGroup.Use(m...)
	publicKeyGroup.Use(a.signatureCheckMiddleware, ccMiddleware)

	a.publicKey.Route(publicKeyGroup.Handle)
}

func NewActivityPub(db db.DB, p *processing.Processor) *ActivityPub {
	return &ActivityPub{
		emoji:                    emoji.New(p),
		users:                    users.New(p),
		publicKey:                publickey.New(p),
		signatureCheckMiddleware: middleware.SignatureCheck(db.IsURIBlocked),
	}
}
