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

package web

import (
	"net/http"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/health"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"codeberg.org/gruf/go-cache/v3"
	"github.com/gin-gonic/gin"
)

type MaintenanceModule struct {
	eTagCache cache.Cache[string, eTagCacheEntry]
}

// NewMaintenance returns a module that routes only
// static assets, and returns a code 503 maintenance
// message template to all other requests.
func NewMaintenance() *MaintenanceModule {
	return &MaintenanceModule{
		eTagCache: newETagCache(),
	}
}

// ETagCache implements withETagCache.
func (m *MaintenanceModule) ETagCache() cache.Cache[string, eTagCacheEntry] {
	return m.eTagCache
}

func (m *MaintenanceModule) Route(r *router.Router, mi ...gin.HandlerFunc) {
	// Route static assets.
	routeAssets(m, r, mi...)

	// Serve OK in response to live
	// requests, but not ready requests.
	liveHandler := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
	r.AttachHandler(http.MethodGet, health.LivePath, liveHandler)
	r.AttachHandler(http.MethodHead, health.LivePath, liveHandler)

	// For everything else, serve maintenance template.
	obj := map[string]string{"host": config.GetHost()}
	r.AttachNoRouteHandler(func(c *gin.Context) {
		retryAfter := time.Now().Add(120 * time.Second).UTC()
		c.Writer.Header().Add("Retry-After", "120")
		c.Writer.Header().Add("Retry-After", retryAfter.Format(http.TimeFormat))
		c.Header("Cache-Control", "no-store")
		c.HTML(http.StatusServiceUnavailable, "maintenance.tmpl", obj)
	})
}
