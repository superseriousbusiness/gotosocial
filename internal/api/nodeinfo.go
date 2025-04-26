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

package api

import (
	"code.superseriousbusiness.org/gotosocial/internal/api/nodeinfo"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"github.com/gin-gonic/gin"
)

type NodeInfo struct {
	nodeInfo *nodeinfo.Module
}

func (w *NodeInfo) Route(r *router.Router, m ...gin.HandlerFunc) {
	// group nodeinfo endpoints together
	nodeInfoGroup := r.AttachGroup("nodeinfo")

	// attach middlewares appropriate for this group
	nodeInfoGroup.Use(m...)
	nodeInfoGroup.Use(
		// Allow public cache for 24 hours.
		middleware.CacheControl(middleware.CacheControlConfig{
			Directives: []string{"public", "max-age=86400"},
			Vary:       []string{"Accept-Encoding"},
		}),
	)

	// If instance is configured to serve instance stats
	// faithfully at nodeinfo, we should allow robots to
	// crawl nodeinfo endpoints in a limited capacity.
	// In all other cases, disallow everything.
	if config.GetInstanceStatsMode() == config.InstanceStatsModeServe {
		nodeInfoGroup.Use(middleware.RobotsHeaders("allowSome"))
	} else {
		nodeInfoGroup.Use(middleware.RobotsHeaders(""))
	}

	w.nodeInfo.Route(nodeInfoGroup.Handle)
}

func NewNodeInfo(p *processing.Processor) *NodeInfo {
	return &NodeInfo{
		nodeInfo: nodeinfo.New(p),
	}
}
