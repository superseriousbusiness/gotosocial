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

//go:build !notracing

package tracing

import (
	"context"
	"fmt"

	"codeberg.org/gruf/go-kv"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunotel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func Initialize() error {
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
	case "jaeger":
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.GetTracingEndpoint())))
		if err != nil {
			return fmt.Errorf("building tracing exporter: %w", err)
		}
		tpo = trace.WithBatcher(exp)
	default:
		return fmt.Errorf("invalid tracing transport: %s", config.GetTracingTransport())
	}
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("GoToSocial"),
		),
	)

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

func InstrumentGin() gin.HandlerFunc {
	return otelgin.Middleware(config.GetHost())
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

func InstrumentBun() bun.QueryHook {
	return bunotel.NewQueryHook(
		bunotel.WithFormattedQueries(true),
	)
}
