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

package observability

import (
	"net/http"
	"slices"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// ginMiddleware returns a middleware that
// records metrics for incoming requests.
func ginMiddleware() gin.HandlerFunc {
	meter := otel.Meter("gin", metric.WithInstrumentationVersion(config.GetSoftwareVersion()))

	activeReqs, _ := meter.Int64UpDownCounter(
		"http.server.requests_active",
		metric.WithDescription("Number of requests currently active"),
	)

	totalReqs, _ := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total number of requests served"),
	)

	reqSize, _ := meter.Int64Histogram(
		"http.server.request_size",
		metric.WithDescription("Request content length (approximate)"),
		metric.WithUnit("bytes"),
	)

	respSize, _ := meter.Int64Histogram(
		"http.server.response_size",
		metric.WithDescription("Response content length"),
		metric.WithUnit("bytes"),
	)

	duration, _ := meter.Int64Histogram(
		"http.server.duration",
		metric.WithDescription("Duration of request -> response"),
		metric.WithUnit("ms"),
	)

	return func(c *gin.Context) {

		ctx := c.Request.Context()
		route := c.FullPath()
		start := time.Now()

		// Generate request attributes.
		reqAttributes := []attribute.KeyValue{
			semconv.HTTPServerNameKey.String("GoToSocial"),
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPRouteKey.String(route),
		}

		// Increment active request count,
		// decrement again when we're finished.
		activeReqs.Add(ctx, 1, metric.WithAttributes(reqAttributes...))
		defer activeReqs.Add(ctx, -1, metric.WithAttributes(reqAttributes...))

		// Process request so we can
		// record response metrics.
		c.Next()

		// Add HTTP response code to request
		// attributes to create response attributes.
		respAttributes := slices.Clone(reqAttributes)
		respAttributes = append(
			respAttributes,
			semconv.HTTPStatusCodeKey.Int(c.Writer.Status()),
		)

		// Increment total requests.
		totalReqs.Add(ctx, 1, metric.WithAttributes(respAttributes...))

		// Record request size.
		reqSize.Record(
			ctx,
			computeApproximateRequestSize(c.Request),
			metric.WithAttributes(respAttributes...),
		)

		// Record response size.
		respSize.Record(
			ctx,
			int64(c.Writer.Size()),
			metric.WithAttributes(respAttributes...),
		)

		// Record req + resp duration.
		duration.Record(
			ctx,
			time.Since(start).Milliseconds(),
			metric.WithAttributes(respAttributes...),
		)
	}
}

func computeApproximateRequestSize(r *http.Request) (sz int64) {
	len := func(s string) int64 { return int64(len(s)) }

	// First line
	sz += len(r.Method)
	if r.URL != nil {
		sz += len(r.URL.Path)
		sz++ // for '?'
		sz += len(r.URL.RawQuery)
		sz += len(r.URL.RawFragment)
	}
	sz += len(r.Proto)
	sz += 2 // for '\r\n'

	// Next are lines for request headers.
	for name, values := range r.Header {

		// Each value on its own line.
		for _, value := range values {
			sz += len(name)
			sz += len(value)
			sz += 2 // for '\r\n'
		}
	}

	// Finally, any request body (if set),
	// this includes (multipart) form data.
	sz += max(r.ContentLength, 0)

	return
}
