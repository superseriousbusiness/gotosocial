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

package robots

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

type Module struct{}

func New() *Module {
	return &Module{}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	// Serve different robots.txt file depending on instance
	// stats mode: Don't disallow scraping nodeinfo if admin
	// has opted in to serving accurate stats there. In all
	// other cases, disallow scraping nodeinfo.
	var handler gin.HandlerFunc
	if config.GetInstanceStatsMode() == config.InstanceStatsModeServe {
		handler = m.robotsGETHandler
	} else {
		handler = m.robotsGETHandlerDisallowNodeInfo
	}

	// Attach handler at empty path as this
	// is already grouped under /robots.txt.
	attachHandler(http.MethodGet, "", handler)
}

func (m *Module) robotsGETHandler(c *gin.Context) {
	c.String(http.StatusOK, apiutil.RobotsTxt)
}

func (m *Module) robotsGETHandlerDisallowNodeInfo(c *gin.Context) {
	c.String(http.StatusOK, apiutil.RobotsTxtDisallowNodeInfo)
}
