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
	"code.superseriousbusiness.org/gotosocial/internal/api/metrics"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"github.com/gin-gonic/gin"
)

type Metrics struct {
	metrics *metrics.Module
}

func (mt *Metrics) Route(r *router.Router, m ...gin.HandlerFunc) {
	if !config.GetMetricsEnabled() {
		// Noop: metrics
		// not enabled.
		return
	}

	// Create new group on top level "metrics" prefix.
	metricsGroup := r.AttachGroup("metrics")
	metricsGroup.Use(m...)
	metricsGroup.Use(
		middleware.CacheControl(middleware.CacheControlConfig{
			// Never cache metrics responses.
			Directives: []string{"no-store"},
		}),
	)

	// Attach basic auth if enabled.
	if config.GetMetricsAuthEnabled() {
		var (
			username = config.GetMetricsAuthUsername()
			password = config.GetMetricsAuthPassword()
			accounts = gin.Accounts{username: password}
		)
		metricsGroup.Use(gin.BasicAuth(accounts))
	}

	mt.metrics.Route(metricsGroup.Handle)
}

func NewMetrics() *Metrics {
	return &Metrics{
		metrics: metrics.New(),
	}
}
