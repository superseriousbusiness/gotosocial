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

// BasePathV1 is the base API path for making media requests through v1 of the api (for mastodon API compatibility)
const BasePathV1 = "/api/v1/media"

// BasePathV2 is the base API path for making media requests through v2 of the api (for mastodon API compatibility)
const BasePathV2 = "/api/v2/media"

// IDKey is the key for media attachment IDs
const IDKey = "id"

// BasePathWithIDV1 corresponds to a media attachment with the given ID
const BasePathWithIDV1 = BasePathV1 + "/:" + IDKey

// BasePathWithIDV2 corresponds to a media attachment with the given ID
const BasePathWithIDV2 = BasePathV2 + "/:" + IDKey

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
	// v1 handlers
	s.AttachHandler(http.MethodPost, BasePathV1, m.MediaCreatePOSTHandler)
	s.AttachHandler(http.MethodGet, BasePathWithIDV1, m.MediaGETHandler)
	s.AttachHandler(http.MethodPut, BasePathWithIDV1, m.MediaPUTHandler)

	// v2 handlers
	s.AttachHandler(http.MethodPost, BasePathV2, m.MediaCreatePOSTHandler)
	s.AttachHandler(http.MethodGet, BasePathWithIDV2, m.MediaGETHandler)
	s.AttachHandler(http.MethodPut, BasePathWithIDV2, m.MediaPUTHandler)

	return nil
}
