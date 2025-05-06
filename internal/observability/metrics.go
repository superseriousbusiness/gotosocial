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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
)

func InitializeMetrics(ctx context.Context, db db.DB) error {
	if !config.GetMetricsEnabled() {
		return nil
	}

	r, err := Resource()
	if err != nil {
		// this can happen if semconv versioning is out-of-sync
		return fmt.Errorf("building tracing resource: %w", err)
	}

	mt, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return err
	}

	meterProvider := sdk.NewMeterProvider(
		sdk.WithExemplarFilter(exemplar.AlwaysOffFilter),
		sdk.WithResource(r),
		sdk.WithReader(mt),
	)

	otel.SetMeterProvider(meterProvider)

	if err := runtime.Start(
		runtime.WithMeterProvider(meterProvider),
	); err != nil {
		return err
	}

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
