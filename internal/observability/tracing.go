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
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"codeberg.org/gruf/go-kv"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

const (
	tracerKey  = "gotosocial-server-tracer"
	tracerName = "code.superseriousbusiness.org/gotosocial/internal/observability"
)

func InitializeTracing(ctx context.Context) error {
	if !config.GetTracingEnabled() {
		return nil
	}

	r, err := Resource()
	if err != nil {
		// this can happen if semconv versioning is out-of-sync
		return fmt.Errorf("building tracing resource: %w", err)
	}

	se, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(r),
		sdktrace.WithBatcher(se),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Hook(func(ctx context.Context, kvs []kv.Field) []kv.Field {
		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().HasTraceID() {
			return append(kvs, kv.Field{K: "traceID", V: span.SpanContext().TraceID().String()})
		}
		return kvs
	})
	return nil
}

// InstrumentGin is a middleware injecting tracing information based on the
// otelgin implementation found at
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gin-gonic/gin/otelgin/gintrace.go
func TracingMiddleware() gin.HandlerFunc {
	provider := otel.GetTracerProvider()
	tracer := provider.Tracer(
		tracerName,
		trace.WithInstrumentationVersion(config.GetSoftwareVersion()),
	)
	propagator := otel.GetTextMapPropagator()
	return func(c *gin.Context) {
		spanName := c.FullPath()
		// Do not trace a request if it didn't match a route. This doesn't omit
		// all 404s as a request matching /:user for a user that doesn't exist
		// still matches the route
		if spanName == "" {
			c.Next()
			return
		}

		c.Set(tracerKey, tracer)
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()
		ctx := propagator.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
		opts := []trace.SpanStartOption{
			trace.WithAttributes(ServerRequestAttributes(c.Request)...),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		rAttr := semconv.HTTPRoute(spanName)
		opts = append(opts, trace.WithAttributes(rAttr))
		id := gtscontext.RequestID(c.Request.Context())
		if id != "" {
			opts = append(opts, trace.WithAttributes(attribute.String("requestID", id)))
		}
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()
		if status > 0 {
			span.SetAttributes(semconv.HTTPResponseStatusCode(status))
		}
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}
	}
}

func InjectRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := gtscontext.RequestID(c.Request.Context())
		if id != "" {
			span := trace.SpanFromContext(c.Request.Context())
			span.SetAttributes(attribute.String("requestID", id))
		}
	}
}

func ServerRequestAttributes(req *http.Request) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 8)
	attrs = append(attrs, method(req.Method))
	attrs = append(attrs, semconv.URLFull(req.URL.RequestURI()))
	attrs = append(attrs, semconv.URLScheme(req.URL.Scheme))
	attrs = append(attrs, semconv.UserAgentOriginal(req.UserAgent()))
	attrs = append(attrs, semconv.NetworkProtocolName("http"))
	attrs = append(attrs, semconv.NetworkProtocolVersion(fmt.Sprintf("%d:%d", req.ProtoMajor, req.ProtoMinor)))

	if ip, port, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		iport, _ := strconv.Atoi(port)
		attrs = append(attrs,
			semconv.NetworkPeerAddress(ip),
			semconv.NetworkPeerPort(iport),
		)
	} else if req.RemoteAddr != "" {
		attrs = append(attrs,
			semconv.NetworkPeerAddress(req.RemoteAddr),
		)
	}

	return attrs
}

func method(m string) attribute.KeyValue {
	var methodLookup = map[string]attribute.KeyValue{
		http.MethodConnect: semconv.HTTPRequestMethodConnect,
		http.MethodDelete:  semconv.HTTPRequestMethodDelete,
		http.MethodGet:     semconv.HTTPRequestMethodGet,
		http.MethodHead:    semconv.HTTPRequestMethodHead,
		http.MethodOptions: semconv.HTTPRequestMethodOptions,
		http.MethodPatch:   semconv.HTTPRequestMethodPatch,
		http.MethodPost:    semconv.HTTPRequestMethodPost,
		http.MethodPut:     semconv.HTTPRequestMethodPut,
		http.MethodTrace:   semconv.HTTPRequestMethodTrace,
	}

	if kv, ok := methodLookup[m]; ok {
		return kv
	}

	return semconv.HTTPRequestMethodGet
}
