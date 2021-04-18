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

package fileserver

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

const (
	AccountIDKey = "account_id"
	MediaTypeKey = "media_type"
	MediaSizeKey = "media_size"
	FileNameKey  = "file_name"

	FilesPath = "files"
)

// FileServer implements the RESTAPIModule interface.
// The goal here is to serve requested media files if the gotosocial server is configured to use local storage.
type FileServer struct {
	config      *config.Config
	db          db.DB
	storage     storage.Storage
	log         *logrus.Logger
	storageBase string
}

// New returns a new fileServer module
func New(config *config.Config, db db.DB, storage storage.Storage, log *logrus.Logger) apimodule.ClientAPIModule {
	return &FileServer{
		config:      config,
		db:          db,
		storage:     storage,
		log:         log,
		storageBase: config.StorageConfig.ServeBasePath,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *FileServer) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, fmt.Sprintf("%s/:%s/:%s/:%s/:%s", m.storageBase, AccountIDKey, MediaTypeKey, MediaSizeKey, FileNameKey), m.ServeFile)
	return nil
}

func (m *FileServer) CreateTables(db db.DB) error {
	models := []interface{}{
		&gtsmodel.MediaAttachment{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}
