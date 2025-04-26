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

package fileserver

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
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
	// FileServePath is the fileserve path minus the 'fileserver/:account_id/:media_type' prefix.
	FileServePath = "/:" + MediaSizeKey + "/:" + FileNameKey
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, FileServePath, m.ServeFile)
	attachHandler(http.MethodHead, FileServePath, m.ServeFile)
}
