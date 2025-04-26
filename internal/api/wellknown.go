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
	"code.superseriousbusiness.org/gotosocial/internal/api/wellknown/hostmeta"
	"code.superseriousbusiness.org/gotosocial/internal/api/wellknown/nodeinfo"
	"code.superseriousbusiness.org/gotosocial/internal/api/wellknown/webfinger"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"github.com/gin-gonic/gin"
)

type WellKnown struct {
	nodeInfo  *nodeinfo.Module
	webfinger *webfinger.Module
	hostMeta  *hostmeta.Module
}

func (w *WellKnown) Route(r *router.Router, m ...gin.HandlerFunc) {
	// group .well-known endpoints together
	wellKnownGroup := r.AttachGroup(".well-known")

	// attach middlewares appropriate for this group
	wellKnownGroup.Use(m...)
	wellKnownGroup.Use(
		// Allow public cache for 2 minutes.
		middleware.CacheControl(middleware.CacheControlConfig{
			Directives: []string{"public", "max-age=120"},
			Vary:       []string{"Accept-Encoding"},
		}),
	)

	w.nodeInfo.Route(wellKnownGroup.Handle)
	w.webfinger.Route(wellKnownGroup.Handle)
	w.hostMeta.Route(wellKnownGroup.Handle)
}

func NewWellKnown(p *processing.Processor) *WellKnown {
	return &WellKnown{
		nodeInfo:  nodeinfo.New(p),
		webfinger: webfinger.New(p),
		hostMeta:  hostmeta.New(p),
	}
}
