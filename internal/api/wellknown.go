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
	"github.com/superseriousbusiness/gotosocial/internal/api/wellknown/nodeinfo"
	"github.com/superseriousbusiness/gotosocial/internal/api/wellknown/webfinger"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type WellKnown struct {
	nodeInfo  *nodeinfo.Module
	webfinger *webfinger.Module
}

func (w *WellKnown) Route(r router.Router, m ...gin.HandlerFunc) {
	// group .well-known endpoints together
	wellKnownGroup := r.AttachGroup(".well-known")

	// attach middlewares appropriate for this group
	wellKnownGroup.Use(m...)
	wellKnownGroup.Use(
		// allow .well-known responses to be cached for 2 minutes
		middleware.CacheControl("public", "max-age=120"),
	)

	w.nodeInfo.Route(wellKnownGroup.Handle)
	w.webfinger.Route(wellKnownGroup.Handle)
}

func NewWellKnown(p processing.Processor) *WellKnown {
	return &WellKnown{
		nodeInfo:  nodeinfo.New(p),
		webfinger: webfinger.New(p),
	}
}
