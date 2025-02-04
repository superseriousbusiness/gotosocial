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

package nodeinfo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	NodeInfo20            = "2.0"
	NodeInfo20ContentType = "application/json; profile=\"http://nodeinfo.diaspora.software/ns/schema/" + NodeInfo20 + "#\""
	NodeInfo21            = "2.1"
	NodeInfo21ContentType = "application/json; profile=\"http://nodeinfo.diaspora.software/ns/schema/" + NodeInfo21 + "#\""
	NodeInfoSchema        = "schema"
	NodeInfoPath          = "/:" + NodeInfoSchema
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
	attachHandler(http.MethodGet, NodeInfoPath, m.NodeInfo2GETHandler)
}
