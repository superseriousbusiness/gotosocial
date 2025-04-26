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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/announcements"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/apps"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/blocks"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/bookmarks"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/conversations"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/customemojis"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/exports"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/favourites"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/featuredtags"
	filtersV1 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v1"
	filtersV2 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v2"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/followedtags"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/followrequests"
	importdata "code.superseriousbusiness.org/gotosocial/internal/api/client/import"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/instance"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/interactionpolicies"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/interactionrequests"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/lists"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/markers"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/media"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/notifications"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/polls"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/preferences"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/push"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/reports"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/search"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/streaming"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/tags"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/timelines"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/tokens"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/user"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/gin-gonic/gin"
)

type Client struct {
	processor *processing.Processor
	db        db.DB

	accounts            *accounts.Module            // api/v1/accounts, api/v1/profile
	admin               *admin.Module               // api/v1/admin
	announcements       *announcements.Module       // api/v1/announcements
	apps                *apps.Module                // api/v1/apps
	blocks              *blocks.Module              // api/v1/blocks
	bookmarks           *bookmarks.Module           // api/v1/bookmarks
	conversations       *conversations.Module       // api/v1/conversations
	customEmojis        *customemojis.Module        // api/v1/custom_emojis
	exports             *exports.Module             // api/v1/exports
	favourites          *favourites.Module          // api/v1/favourites
	featuredTags        *featuredtags.Module        // api/v1/featured_tags
	filtersV1           *filtersV1.Module           // api/v1/filters
	filtersV2           *filtersV2.Module           // api/v2/filters
	followRequests      *followrequests.Module      // api/v1/follow_requests
	followedTags        *followedtags.Module        // api/v1/followed_tags
	importData          *importdata.Module          // api/v1/import
	instance            *instance.Module            // api/v1/instance
	interactionPolicies *interactionpolicies.Module // api/v1/interaction_policies
	interactionRequests *interactionrequests.Module // api/v1/interaction_requests
	lists               *lists.Module               // api/v1/lists
	markers             *markers.Module             // api/v1/markers
	media               *media.Module               // api/v1/media, api/v2/media
	mutes               *mutes.Module               // api/v1/mutes
	notifications       *notifications.Module       // api/v1/notifications
	polls               *polls.Module               // api/v1/polls
	preferences         *preferences.Module         // api/v1/preferences
	push                *push.Module                // api/v1/push
	reports             *reports.Module             // api/v1/reports
	search              *search.Module              // api/v1/search, api/v2/search
	statuses            *statuses.Module            // api/v1/statuses
	streaming           *streaming.Module           // api/v1/streaming
	tags                *tags.Module                // api/v1/tags
	timelines           *timelines.Module           // api/v1/timelines
	tokens              *tokens.Module              // api/v1/tokens
	user                *user.Module                // api/v1/user
}

func (c *Client) Route(r *router.Router, m ...gin.HandlerFunc) {
	// create a new group on the top level client 'api' prefix
	apiGroup := r.AttachGroup("api")

	// attach non-global middlewares appropriate to the client api
	apiGroup.Use(m...)
	apiGroup.Use(
		middleware.TokenCheck(c.db, c.processor.OAuthValidateBearerToken),
		middleware.CacheControl(middleware.CacheControlConfig{
			// Never cache client api responses.
			Directives: []string{"no-store"},
		}),
	)

	// for each client api module, pass it the Handle function
	// so that the module can attach its routes to this group
	h := apiGroup.Handle
	c.accounts.Route(h)
	c.admin.Route(h)
	c.announcements.Route(h)
	c.apps.Route(h)
	c.blocks.Route(h)
	c.bookmarks.Route(h)
	c.conversations.Route(h)
	c.customEmojis.Route(h)
	c.exports.Route(h)
	c.favourites.Route(h)
	c.featuredTags.Route(h)
	c.filtersV1.Route(h)
	c.filtersV2.Route(h)
	c.followRequests.Route(h)
	c.followedTags.Route(h)
	c.importData.Route(h)
	c.instance.Route(h)
	c.interactionPolicies.Route(h)
	c.interactionRequests.Route(h)
	c.lists.Route(h)
	c.markers.Route(h)
	c.media.Route(h)
	c.mutes.Route(h)
	c.notifications.Route(h)
	c.polls.Route(h)
	c.preferences.Route(h)
	c.push.Route(h)
	c.reports.Route(h)
	c.search.Route(h)
	c.statuses.Route(h)
	c.streaming.Route(h)
	c.tags.Route(h)
	c.timelines.Route(h)
	c.tokens.Route(h)
	c.user.Route(h)
}

func NewClient(state *state.State, p *processing.Processor) *Client {
	return &Client{
		processor: p,
		db:        state.DB,

		accounts:            accounts.New(p),
		admin:               admin.New(state, p),
		announcements:       announcements.New(p),
		apps:                apps.New(p),
		blocks:              blocks.New(p),
		bookmarks:           bookmarks.New(p),
		conversations:       conversations.New(p),
		customEmojis:        customemojis.New(p),
		exports:             exports.New(p),
		favourites:          favourites.New(p),
		featuredTags:        featuredtags.New(p),
		filtersV1:           filtersV1.New(p),
		filtersV2:           filtersV2.New(p),
		followRequests:      followrequests.New(p),
		followedTags:        followedtags.New(p),
		importData:          importdata.New(p),
		instance:            instance.New(p),
		interactionPolicies: interactionpolicies.New(p),
		interactionRequests: interactionrequests.New(p),
		lists:               lists.New(p),
		markers:             markers.New(p),
		media:               media.New(p),
		mutes:               mutes.New(p),
		notifications:       notifications.New(p),
		polls:               polls.New(p),
		preferences:         preferences.New(p),
		push:                push.New(p),
		reports:             reports.New(p),
		search:              search.New(p),
		statuses:            statuses.New(p),
		streaming:           streaming.New(p, time.Second*30, 4096),
		tags:                tags.New(p),
		timelines:           timelines.New(p),
		tokens:              tokens.New(p),
		user:                user.New(p),
	}
}
