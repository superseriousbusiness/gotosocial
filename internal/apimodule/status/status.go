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

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	idKey          = "id"
	basePath       = "/api/v1/statuses"
	basePathWithID = basePath + "/:" + idKey
	contextPath    = basePath + "/context"
	rebloggedPath  = basePath + "/reblogged_by"
	favouritedPath = basePath + "/favourited_by"
	favouritePath  = basePath + "/favourite"
	reblogPath     = basePath + "/reblog"
	unreblogPath   = basePath + "/unreblog"
	bookmarkPath   = basePath + "/bookmark"
	unbookmarkPath = basePath + "/unbookmark"
	mutePath       = basePath + "/mute"
	unmutePath     = basePath + "/unmute"
	pinPath        = basePath + "/pin"
	unpinPath      = basePath + "/unpin"
)

type statusModule struct {
	config       *config.Config
	db           db.DB
	oauthServer  oauth.Server
	mediaHandler media.MediaHandler
	distributor  distributor.Distributor
	log          *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, oauthServer oauth.Server, mediaHandler media.MediaHandler, distributor distributor.Distributor, log *logrus.Logger) apimodule.ClientAPIModule {
	return &statusModule{
		config:       config,
		db:           db,
		mediaHandler: mediaHandler,
		distributor:  distributor,
		log:          log,
	}
}

// Route attaches all routes from this module to the given router
func (m *statusModule) Route(r router.Router) error {
	// r.AttachHandler(http.MethodPost, basePath, m.accountCreatePOSTHandler)
	// r.AttachHandler(http.MethodGet, basePathWithID, m.muxHandler)
	return nil
}

func (m *statusModule) CreateTables(db db.DB) error {
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.FollowRequest{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
		&model.MediaAttachment{},
		&model.Emoji{},
		&model.Tag{},
		&model.Mention{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}
