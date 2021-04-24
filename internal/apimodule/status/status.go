/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	"github.com/superseriousbusiness/gotosocial/internal/media"
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
	config         *config.Config
	db             db.DB
	mediaHandler   media.Handler
	mastoConverter mastotypes.Converter
	distributor    distributor.Distributor
	log            *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, mediaHandler media.Handler, mastoConverter mastotypes.Converter, distributor distributor.Distributor, log *logrus.Logger) apimodule.ClientAPIModule {
	return &Module{
		config:         config,
		db:             db,
		mediaHandler:   mediaHandler,
		mastoConverter: mastoConverter,
		distributor:    distributor,
		log:            log,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, BasePath, m.StatusCreatePOSTHandler)
	r.AttachHandler(http.MethodDelete, BasePathWithID, m.StatusDELETEHandler)

	r.AttachHandler(http.MethodPost, FavouritePath, m.StatusFavePOSTHandler)
	r.AttachHandler(http.MethodPost, UnfavouritePath, m.StatusFavePOSTHandler)

	r.AttachHandler(http.MethodGet, BasePathWithID, m.muxHandler)
	return nil
}

// CreateTables populates necessary tables in the given DB
func (m *Module) CreateTables(db db.DB) error {
	models := []interface{}{
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Block{},
		&gtsmodel.Follow{},
		&gtsmodel.FollowRequest{},
		&gtsmodel.Status{},
		&gtsmodel.StatusFave{},
		&gtsmodel.StatusBookmark{},
		&gtsmodel.StatusMute{},
		&gtsmodel.StatusPin{},
		&gtsmodel.Application{},
		&gtsmodel.EmailDomainBlock{},
		&gtsmodel.MediaAttachment{},
		&gtsmodel.Emoji{},
		&gtsmodel.Tag{},
		&gtsmodel.Mention{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}

// muxHandler is a little workaround to overcome the limitations of Gin
func (m *Module) muxHandler(c *gin.Context) {
	m.log.Debug("entering mux handler")
	ru := c.Request.RequestURI

	switch c.Request.Method {
	case http.MethodGet:
		if strings.HasPrefix(ru, ContextPath) {
			// TODO
		} else if strings.HasPrefix(ru, FavouritedPath) {
			m.StatusFavedByGETHandler(c)
		} else {
			m.StatusGETHandler(c)
		}
	}
}
