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
	"code.superseriousbusiness.org/gotosocial/internal/state"

	"github.com/gin-gonic/gin"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
)

func InitializeMetrics(ctx context.Context, state *state.State) error {
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
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			userCount, err := state.DB.CountInstanceUsers(ctx, thisInstance)
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
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			statusCount, err := state.DB.CountInstanceStatuses(ctx, thisInstance)
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
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			federatingCount, err := state.DB.CountInstanceDomains(ctx, thisInstance)
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

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.delivery.count",
		metric.WithDescription("Current number of delivery workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Delivery.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.delivery.queue",
		metric.WithDescription("Current number of queued delivery worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Delivery.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.dereference.count",
		metric.WithDescription("Current number of dereference workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Dereference.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.dereference.queue",
		metric.WithDescription("Current number of queued dereference worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Dereference.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.client_api.count",
		metric.WithDescription("Current number of client API workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Client.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.client_api.queue",
		metric.WithDescription("Current number of queued client API worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Client.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.fedi_api.count",
		metric.WithDescription("Current number of federator API workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Federator.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.fedi_api.queue",
		metric.WithDescription("Current number of queued federator API worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Federator.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.processing.count",
		metric.WithDescription("Current number of processing workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Processing.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.processing.queue",
		metric.WithDescription("Current number of queued processing worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.Processing.Queue.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge(
		"gotosocial.workers.webpush.count",
		metric.WithDescription("Current number of webpush workers"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.WebPush.Len()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter(
		"gotosocial.workers.webpush.queue",
		metric.WithDescription("Current number of queued webpush worker tasks"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(state.Workers.WebPush.Queue.Len()))
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
