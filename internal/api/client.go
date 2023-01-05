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
	"github.com/superseriousbusiness/gotosocial/internal/api/client/accounts"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/apps"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/blocks"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/bookmarks"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/customemojis"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/favourites"
	filter "github.com/superseriousbusiness/gotosocial/internal/api/client/filters"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/followrequests"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/lists"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/notifications"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/search"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/streaming"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/timelines"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/user"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Client struct {
	processor processing.Processor
	db        db.DB

	accounts       *accounts.Module       // api/v1/accounts
	admin          *admin.Module          // api/v1/admin
	apps           *apps.Module           // api/v1/apps
	blocks         *blocks.Module         // api/v1/blocks
	bookmarks      *bookmarks.Module      // api/v1/bookmarks
	customEmojis   *customemojis.Module   // api/v1/custom_emojis
	favourites     *favourites.Module     // api/v1/favourites
	filters        *filter.Module         // api/v1/filters
	followRequests *followrequests.Module // api/v1/follow_requests
	instance       *instance.Module       // api/v1/instance
	lists          *lists.Module          // api/v1/lists
	media          *media.Module          // api/v1/media, api/v2/media
	notifications  *notifications.Module  // api/v1/notifications
	search         *search.Module         // api/v1/search, api/v2/search
	statuses       *statuses.Module       // api/v1/statuses
	streaming      *streaming.Module      // api/v1/streaming
	timelines      *timelines.Module      // api/v1/timelines
	user           *user.Module           // api/v1/user
}

func (c *Client) Route(r router.Router, m ...gin.HandlerFunc) {
	// create a new group on the top level client 'api' prefix
	apiGroup := r.AttachGroup("api")

	// attach non-global middlewares appropriate to the client api
	apiGroup.Use(m...)
	apiGroup.Use(
		middleware.TokenCheck(c.db, c.processor.OAuthValidateBearerToken),
		middleware.CacheControl("no-store"), // never cache api responses
	)

	// for each client api module, pass it the Handle function
	// so that the module can attach its routes to this group
	h := apiGroup.Handle
	c.accounts.Route(h)
	c.admin.Route(h)
	c.apps.Route(h)
	c.blocks.Route(h)
	c.bookmarks.Route(h)
	c.customEmojis.Route(h)
	c.favourites.Route(h)
	c.filters.Route(h)
	c.followRequests.Route(h)
	c.instance.Route(h)
	c.lists.Route(h)
	c.media.Route(h)
	c.notifications.Route(h)
	c.search.Route(h)
	c.statuses.Route(h)
	c.streaming.Route(h)
	c.timelines.Route(h)
	c.user.Route(h)
}

func NewClient(db db.DB, p processing.Processor) *Client {
	return &Client{
		processor: p,
		db:        db,

		accounts:       accounts.New(p),
		admin:          admin.New(p),
		apps:           apps.New(p),
		blocks:         blocks.New(p),
		bookmarks:      bookmarks.New(p),
		customEmojis:   customemojis.New(p),
		favourites:     favourites.New(p),
		filters:        filter.New(p),
		followRequests: followrequests.New(p),
		instance:       instance.New(p),
		lists:          lists.New(p),
		media:          media.New(p),
		notifications:  notifications.New(p),
		search:         search.New(p),
		statuses:       statuses.New(p),
		streaming:      streaming.New(p),
		timelines:      timelines.New(p),
		user:           user.New(p),
	}
}
