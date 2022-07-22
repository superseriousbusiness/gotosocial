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

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	IDKey                  = "id"                                // IDKey is the key for media attachment IDs
	APIVersionKey          = "api_version"                       // APIVersionKey is the key for which version of the API to use (v1 or v2)
	BasePathWithAPIVersion = "/api/:" + APIVersionKey + "/media" // BasePathWithAPIVersion is the base API path for making media requests through v1 or v2 of the api (for mastodon API compatibility)
	BasePathWithIDV1       = "/api/v1/media/:" + IDKey           // BasePathWithID corresponds to a media attachment with the given ID
)

// Module implements the ClientAPIModule interface for media
type Module struct {
	processor processing.Processor
}

// New returns a new auth module
func New(processor processing.Processor) api.ClientModule {
	return &Module{
		processor: processor,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodPost, BasePathWithAPIVersion, m.MediaCreatePOSTHandler)
	s.AttachHandler(http.MethodGet, BasePathWithIDV1, m.MediaGETHandler)
	s.AttachHandler(http.MethodPut, BasePathWithIDV1, m.MediaPUTHandler)
	return nil
}
