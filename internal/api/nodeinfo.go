/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/nodeinfo"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type NodeInfo struct {
	nodeInfo *nodeinfo.Module
}

func (w *NodeInfo) Route(r router.Router, m ...gin.HandlerFunc) {
	// group nodeinfo endpoints together
	nodeInfoGroup := r.AttachGroup("nodeinfo")

	// attach middlewares appropriate for this group
	nodeInfoGroup.Use(m...)
	nodeInfoGroup.Use(
		// allow cache for 2 minutes
		middleware.CacheControl("public", "max-age=120"),
	)

	w.nodeInfo.Route(nodeInfoGroup.Handle)
}

func NewNodeInfo(p processing.Processor) *NodeInfo {
	return &NodeInfo{
		nodeInfo: nodeinfo.New(p),
	}
}
