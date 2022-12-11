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

package status

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// IDKey is for status UUIDs
	IDKey = "id"
	// BasePath is the base path for serving the status API
	BasePath = "/api/v1/statuses"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the status being queried.
	BasePathWithID = BasePath + "/:" + IDKey

	// ContextPath is used for fetching context of posts
	ContextPath = BasePathWithID + "/context"

	// FavouritedPath is for seeing who's faved a given status
	FavouritedPath = BasePathWithID + "/favourited_by"
	// FavouritePath is for posting a fave on a status
	FavouritePath = BasePathWithID + "/favourite"
	// UnfavouritePath is for removing a fave from a status
	UnfavouritePath = BasePathWithID + "/unfavourite"

	// RebloggedPath is for seeing who's boosted a given status
	RebloggedPath = BasePathWithID + "/reblogged_by"
	// ReblogPath is for boosting/reblogging a given status
	ReblogPath = BasePathWithID + "/reblog"
	// UnreblogPath is for undoing a boost/reblog of a given status
	UnreblogPath = BasePathWithID + "/unreblog"

	// BookmarkPath is for creating a bookmark on a given status
	BookmarkPath = BasePathWithID + "/bookmark"
	// UnbookmarkPath is for removing a bookmark from a given status
	UnbookmarkPath = BasePathWithID + "/unbookmark"

	// MutePath is for muting a given status so that notifications will no longer be received about it.
	MutePath = BasePathWithID + "/mute"
	// UnmutePath is for undoing an existing mute
	UnmutePath = BasePathWithID + "/unmute"

	// PinPath is for pinning a status to an account profile so that it's the first thing people see
	PinPath = BasePathWithID + "/pin"
	// UnpinPath is for undoing a pin and returning a status to the ever-swirling drain of time and entropy
	UnpinPath = BasePathWithID + "/unpin"
)

// Module implements the ClientAPIModule interface for every related to posting/deleting/interacting with statuses
type Module struct {
	processor processing.Processor
}

// New returns a new account module
func New(processor processing.Processor) api.ClientModule {
	return &Module{
		processor: processor,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, BasePath, m.StatusCreatePOSTHandler)
	r.AttachHandler(http.MethodDelete, BasePathWithID, m.StatusDELETEHandler)

	r.AttachHandler(http.MethodPost, FavouritePath, m.StatusFavePOSTHandler)
	r.AttachHandler(http.MethodPost, UnfavouritePath, m.StatusUnfavePOSTHandler)
	r.AttachHandler(http.MethodGet, FavouritedPath, m.StatusFavedByGETHandler)

	r.AttachHandler(http.MethodPost, ReblogPath, m.StatusBoostPOSTHandler)
	r.AttachHandler(http.MethodPost, UnreblogPath, m.StatusUnboostPOSTHandler)
	r.AttachHandler(http.MethodGet, RebloggedPath, m.StatusBoostedByGETHandler)

	r.AttachHandler(http.MethodPost, BookmarkPath, m.StatusBookmarkPOSTHandler)
	r.AttachHandler(http.MethodPost, UnbookmarkPath, m.StatusUnbookmarkPOSTHandler)

	r.AttachHandler(http.MethodGet, ContextPath, m.StatusContextGETHandler)

	r.AttachHandler(http.MethodGet, BasePathWithID, m.muxHandler)
	return nil
}

// muxHandler is a little workaround to overcome the limitations of Gin
func (m *Module) muxHandler(c *gin.Context) {
	log.Debug("entering mux handler")
	ru := c.Request.RequestURI

	if c.Request.Method == http.MethodGet {
		switch {
		case strings.HasPrefix(ru, ContextPath):
			// TODO
		case strings.HasPrefix(ru, FavouritedPath):
			m.StatusFavedByGETHandler(c)
		default:
			m.StatusGETHandler(c)
		}
	}
}
