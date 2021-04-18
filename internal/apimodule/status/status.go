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
	IDKey          = "id"
	BasePath       = "/api/v1/statuses"
	BasePathWithID = BasePath + "/:" + IDKey
	ContextPath    = BasePath + "/context"
	RebloggedPath  = BasePath + "/reblogged_by"
	FavouritedPath = BasePath + "/favourited_by"
	FavouritePath  = BasePath + "/favourite"
	ReblogPath     = BasePath + "/reblog"
	UnreblogPath   = BasePath + "/unreblog"
	BookmarkPath   = BasePath + "/bookmark"
	UnbookmarkPath = BasePath + "/unbookmark"
	MutePath       = BasePath + "/mute"
	UnmutePath     = BasePath + "/unmute"
	PinPath        = BasePath + "/pin"
	UnpinPath      = BasePath + "/unpin"
)

type StatusModule struct {
	config         *config.Config
	db             db.DB
	mediaHandler   media.MediaHandler
	mastoConverter mastotypes.Converter
	distributor    distributor.Distributor
	log            *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, mediaHandler media.MediaHandler, mastoConverter mastotypes.Converter, distributor distributor.Distributor, log *logrus.Logger) apimodule.ClientAPIModule {
	return &StatusModule{
		config:         config,
		db:             db,
		mediaHandler:   mediaHandler,
		mastoConverter: mastoConverter,
		distributor:    distributor,
		log:            log,
	}
}

// Route attaches all routes from this module to the given router
func (m *StatusModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, BasePath, m.StatusCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, BasePathWithID, m.muxHandler)
	r.AttachHandler(http.MethodDelete, BasePathWithID, m.muxHandler)
	return nil
}

func (m *StatusModule) CreateTables(db db.DB) error {
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

func (m *StatusModule) muxHandler(c *gin.Context) {
	m.log.Debug("entering mux handler")
	ru := c.Request.RequestURI
	if c.Request.Method == http.MethodGet {
		if strings.HasPrefix(ru, ContextPath) {
			// TODO
		} else if strings.HasPrefix(ru, RebloggedPath) {
			// TODO
		} else {
			m.StatusGETHandler(c)
		}
	}
	if c.Request.Method == http.MethodDelete {
		m.StatusDELETEHandler(c)
	}
}
