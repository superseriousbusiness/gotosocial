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

package instance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	// InstanceInformationPath is for serving instance info requests, minus the 'api' prefix.
	InstanceInformationPath = "/v1/instance"
	// InstancePeersPath is for serving instance peers requests.
	InstancePeersPath = InstanceInformationPath + "/peers"
	// PeersFilterKey is used to provide filters to /api/v1/instance/peers
	PeersFilterKey = "filter"
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
	attachHandler(http.MethodGet, InstanceInformationPath, m.InstanceInformationGETHandler)
	attachHandler(http.MethodPatch, InstanceInformationPath, m.InstanceUpdatePATCHHandler)
	attachHandler(http.MethodGet, InstancePeersPath, m.InstancePeersGETHandler)
}
