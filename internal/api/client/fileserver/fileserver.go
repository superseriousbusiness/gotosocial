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

package fileserver

import (
	"fmt"
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	// FileServeBasePath forms the first part of the fileserver path.
	FileServeBasePath = "/" + uris.FileserverPath
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
	processor processing.Processor
}

// New returns a new fileServer module
func New(processor processing.Processor) api.ClientModule {
	return &FileServer{
		processor: processor,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *FileServer) Route(s router.Router) error {
	// something like "/fileserver/:account_id/:media_type/:media_size/:file_name"
	fileServePath := fmt.Sprintf("%s/:%s/:%s/:%s/:%s", FileServeBasePath, AccountIDKey, MediaTypeKey, MediaSizeKey, FileNameKey)
	s.AttachHandler(http.MethodGet, fileServePath, m.ServeFile)
	return nil
}
