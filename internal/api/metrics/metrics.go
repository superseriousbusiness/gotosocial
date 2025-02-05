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

//go:build !nootel

package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Module struct {
	handler http.Handler
}

func New() *Module {
	// Let prometheus use "identity", ie., no compression,
	// or "gzip", to match our own gzip compression middleware.
	opts := promhttp.HandlerOpts{
		OfferedCompressions: []promhttp.Compression{
			promhttp.Identity,
			promhttp.Gzip,
		},
	}

	// Instrument handler itself.
	handler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer, opts),
	)

	return &Module{
		handler: handler,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, "", func(c *gin.Context) {
		// Defer all "/metrics" handling to prom.
		m.handler.ServeHTTP(c.Writer, c.Request)
	})
}
