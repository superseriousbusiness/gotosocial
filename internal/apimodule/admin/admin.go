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
	// BasePath is the base API path for this module
	BasePath = "/api/v1/admin"
	// EmojiPath is used for posting/deleting custom emojis
	EmojiPath = BasePath + "/custom_emojis"
)

// Module implements the ClientAPIModule interface for admin-related actions (reports, emojis, etc)
type Module struct {
	config         *config.Config
	db             db.DB
	mediaHandler   media.Handler
	mastoConverter mastotypes.Converter
	log            *logrus.Logger
}

// New returns a new admin module
func New(config *config.Config, db db.DB, mediaHandler media.Handler, mastoConverter mastotypes.Converter, log *logrus.Logger) apimodule.ClientAPIModule {
	return &Module{
		config:         config,
		db:             db,
		mediaHandler:   mediaHandler,
		mastoConverter: mastoConverter,
		log:            log,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, EmojiPath, m.emojiCreatePOSTHandler)
	return nil
}

// CreateTables creates the necessary tables for this module in the given database
func (m *Module) CreateTables(db db.DB) error {
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
