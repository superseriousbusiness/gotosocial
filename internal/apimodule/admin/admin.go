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

package admin

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	basePath  = "/api/v1/admin"
	emojiPath = basePath + "/custom_emojis"
)

type adminModule struct {
	config         *config.Config
	db             db.DB
	mediaHandler   media.MediaHandler
	mastoConverter mastotypes.Converter
	log            *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, mediaHandler media.MediaHandler, mastoConverter mastotypes.Converter, log *logrus.Logger) apimodule.ClientAPIModule {
	return &adminModule{
		config:         config,
		db:             db,
		mediaHandler:   mediaHandler,
		mastoConverter: mastoConverter,
		log:            log,
	}
}

// Route attaches all routes from this module to the given router
func (m *adminModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, emojiPath, m.emojiCreatePOSTHandler)
	return nil
}

func (m *adminModule) CreateTables(db db.DB) error {
	models := []interface{}{
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Follow{},
		&gtsmodel.FollowRequest{},
		&gtsmodel.Status{},
		&gtsmodel.Application{},
		&gtsmodel.EmailDomainBlock{},
		&gtsmodel.MediaAttachment{},
		&gtsmodel.Emoji{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}
