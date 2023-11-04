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

//go:build !nometrics

package metrics

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunotel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	serviceName = "GoToSocial"
)

func Initialize() error {
	if !config.GetMetricsEnabled() {
		return nil
	}

	fmt.Println("Initializing metrics")
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	if config.GetMetricsExporter() == "prometheus" {
		prometheusExporter, err := prometheus.New()
		if err != nil {
			return err
		}

		meterProvider := sdk.NewMeterProvider(
			sdk.WithResource(r),
			sdk.WithReader(prometheusExporter),
		)
		otel.SetMeterProvider(meterProvider)
	}

	return nil
}

func InstrumentGin() gin.HandlerFunc {
	return otelginmetrics.Middleware(serviceName)
}

func InstrumentBun() bun.QueryHook {
	return bunotel.NewQueryHook(
		bunotel.WithMeterProvider(otel.GetMeterProvider()),
	)
}
