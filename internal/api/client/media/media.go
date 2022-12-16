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

package media

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	IDKey            = "id"                            // IDKey is the key for media attachment IDs
	APIVersionKey    = "api_version"                   // APIVersionKey is the key for which version of the API to use (v1 or v2)
	APIv1            = "v1"                            // APIV1 corresponds to version 1 of the api
	APIv2            = "v2"                            // APIV2 corresponds to version 2 of the api
	BasePath         = "/:" + APIVersionKey + "/media" // BasePath is the base API path for making media requests through v1 or v2 of the api (for mastodon API compatibility)
	AttachmentWithID = BasePath + "/:" + IDKey         // BasePathWithID corresponds to a media attachment with the given ID
)

type Module struct {
	processor processing.Processor
}

func New(processor processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodPost, BasePath, m.MediaCreatePOSTHandler)
	attachHandler(http.MethodGet, AttachmentWithID, m.MediaGETHandler)
	attachHandler(http.MethodPut, AttachmentWithID, m.MediaPUTHandler)
}
