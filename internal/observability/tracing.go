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

	"codeberg.org/gruf/go-kv"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

const (
	tracerKey  = "gotosocial-server-tracer"
	tracerName = "github.com/superseriousbusiness/gotosocial/internal/observability"
)

func InitializeTracing() error {
	if !config.GetTracingEnabled() {
		return nil
	}

	insecure := config.GetTracingInsecureTransport()

	var tpo trace.TracerProviderOption
	switch config.GetTracingTransport() {
	case "grpc":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(config.GetTracingEndpoint()),
		}
		if insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exp, err := otlptracegrpc.New(context.Background(), opts...)
		if err != nil {
			return fmt.Errorf("building tracing exporter: %w", err)
		}
		tpo = trace.WithBatcher(exp)
	case "http":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(config.GetTracingEndpoint()),
		}
		if insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			return fmt.Errorf("building tracing exporter: %w", err)
		}
		tpo = trace.WithBatcher(exp)
	default:
		return fmt.Errorf("invalid tracing transport: %s", config.GetTracingTransport())
	}
	r, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName("GoToSocial"),
			semconv.ServiceVersion(config.GetSoftwareVersion()),
		),
	)
	if err != nil {
		// this can happen if semconv versioning is out-of-sync
		return fmt.Errorf("building tracing resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		tpo,
		trace.WithResource(r),
	)
	otel.SetTracerProvider(tp)
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)
	log.Hook(func(ctx context.Context, kvs []kv.Field) []kv.Field {
		span := oteltrace.SpanFromContext(ctx)
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
		oteltrace.WithInstrumentationVersion(config.GetSoftwareVersion()),
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
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(httpconv.ServerRequest(config.GetHost(), c.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		rAttr := semconv.HTTPRoute(spanName)
		opts = append(opts, oteltrace.WithAttributes(rAttr))
		id := gtscontext.RequestID(c.Request.Context())
		if id != "" {
			opts = append(opts, oteltrace.WithAttributes(attribute.String("requestID", id)))
		}
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()
		span.SetStatus(httpconv.ServerStatus(status))
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
			span := oteltrace.SpanFromContext(c.Request.Context())
			span.SetAttributes(attribute.String("requestID", id))
		}
	}
}
