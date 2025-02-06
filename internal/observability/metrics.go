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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	serviceName = "GoToSocial"
)

func InitializeMetrics(db db.DB) error {
	if !config.GetMetricsEnabled() {
		return nil
	}

	if config.GetMetricsAuthEnabled() {
		if config.GetMetricsAuthPassword() == "" || config.GetMetricsAuthUsername() == "" {
			return errors.New("metrics-auth-username and metrics-auth-password must be set when metrics-auth-enabled is true")
		}
	}

	r, _ := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(config.GetSoftwareVersion()),
		),
	)

	prometheusExporter, err := prometheus.New()
	if err != nil {
		return err
	}

	meterProvider := sdk.NewMeterProvider(
		sdk.WithExemplarFilter(exemplar.AlwaysOffFilter),
		sdk.WithResource(r),
		sdk.WithReader(prometheusExporter),
	)

	otel.SetMeterProvider(meterProvider)

	meter := meterProvider.Meter(serviceName)

	thisInstance := config.GetHost()

	_, err = meter.Int64ObservableGauge(
		"gotosocial.instance.total_users",
		metric.WithDescription("Total number of users on this instance"),
		metric.WithInt64Callback(func(c context.Context, o metric.Int64Observer) error {
			userCount, err := db.CountInstanceUsers(c, thisInstance)
			if err != nil {
				return err
			}
			o.Observe(int64(userCount))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.instance.total_statuses",
		metric.WithDescription("Total number of statuses on this instance"),
		metric.WithInt64Callback(func(c context.Context, o metric.Int64Observer) error {
			statusCount, err := db.CountInstanceStatuses(c, thisInstance)
			if err != nil {
				return err
			}
			o.Observe(int64(statusCount))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.instance.total_federating_instances",
		metric.WithDescription("Total number of other instances this instance is federating with"),
		metric.WithInt64Callback(func(c context.Context, o metric.Int64Observer) error {
			federatingCount, err := db.CountInstanceDomains(c, thisInstance)
			if err != nil {
				return err
			}
			o.Observe(int64(federatingCount))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

func MetricsMiddleware() gin.HandlerFunc {
	return otelginmetrics.Middleware(serviceName)
}
