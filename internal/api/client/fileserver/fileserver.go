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

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// AccountIDKey is the url key for account id (an account ulid)
	AccountIDKey = "account_id"
	// MediaTypeKey is the url key for media type (usually something like attachment or header etc)
	MediaTypeKey = "media_type"
	// MediaSizeKey is the url key for the desired media size--original/small/static
	MediaSizeKey = "media_size"
	// FileNameKey is the actual filename being sought. Will usually be a UUID then something like .jpeg
	FileNameKey = "file_name"
)

// FileServer implements the RESTAPIModule interface.
// The goal here is to serve requested media files if the gotosocial server is configured to use local storage.
type FileServer struct {
	config      *config.Config
	processor   processing.Processor
	storageBase string
}

// New returns a new fileServer module
func New(config *config.Config, processor processing.Processor) api.ClientModule {
	return &FileServer{
		config:      config,
		processor:   processor,
		storageBase: config.StorageConfig.ServeBasePath,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *FileServer) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, fmt.Sprintf("%s/:%s/:%s/:%s/:%s", m.storageBase, AccountIDKey, MediaTypeKey, MediaSizeKey, FileNameKey), m.ServeFile)
	return nil
}
