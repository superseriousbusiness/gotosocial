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

package instance

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

const (
	InstanceInformationPathV1 = "/v1/instance"
	InstanceInformationPathV2 = "/v2/instance"
	InstancePeersPath         = InstanceInformationPathV1 + "/peers"
	InstanceRulesPath         = InstanceInformationPathV1 + "/rules"
	InstanceBlocklistPath     = InstanceInformationPathV1 + "/domain_blocks"
	InstanceAllowlistPath     = InstanceInformationPathV1 + "/domain_allows"
	PeersFilterKey            = "filter" // PeersFilterKey is used to provide filters to /api/v1/instance/peers
	PeersFlatKey              = "flat"   // PeersFlatKey is used to set "flat=true" in /api/v1/instance/peers
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
	attachHandler(http.MethodGet, InstanceInformationPathV1, m.InstanceInformationGETHandlerV1)
	attachHandler(http.MethodGet, InstanceInformationPathV2, m.InstanceInformationGETHandlerV2)
	attachHandler(http.MethodPatch, InstanceInformationPathV1, m.InstanceUpdatePATCHHandler)
	attachHandler(http.MethodGet, InstancePeersPath, m.InstancePeersGETHandler)
	attachHandler(http.MethodGet, InstanceRulesPath, m.InstanceRulesGETHandler)
	attachHandler(http.MethodGet, InstanceBlocklistPath, m.InstanceDomainBlocksGETHandler)
	attachHandler(http.MethodGet, InstanceAllowlistPath, m.InstanceDomainAllowsGETHandler)
}
