package otelsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumName = "github.com/uptrace/opentelemetry-go-extra/otelsql"

var dbRowsAffected = attribute.Key("db.rows_affected")

type config struct {
	tracerProvider trace.TracerProvider
	tracer         trace.Tracer //nolint:structcheck

	meterProvider metric.MeterProvider
	meter         metric.Meter

	attrs []attribute.KeyValue

	queryFormatter func(query string) string
}

func newConfig(opts []Option) *config {
	c := &config{
		tracerProvider: otel.GetTracerProvider(),
		meterProvider:  global.MeterProvider(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *config) formatQuery(query string) string {
	if c.queryFormatter != nil {
		return c.queryFormatter(query)
	}
	return query
}

type dbInstrum struct {
	*config

	queryHistogram instrument.Int64Histogram
}

func newDBInstrum(opts []Option) *dbInstrum {
	t := &dbInstrum{
		config: newConfig(opts),
	}

	if t.tracer == nil {
		t.tracer = t.tracerProvider.Tracer(instrumName)
	}
	if t.meter == nil {
		t.meter = t.meterProvider.Meter(instrumName)
	}

	var err error
	t.queryHistogram, err = t.meter.Int64Histogram(
		"go.sql.query_timing",
		instrument.WithDescription("Timing of processed queries"),
		instrument.WithUnit("milliseconds"),
	)
	if err != nil {
		panic(err)
	}

	return t
}

func (t *dbInstrum) withSpan(
	ctx context.Context,
	spanName string,
	query string,
	fn func(ctx context.Context, span trace.Span) error,
) error {
	var startTime time.Time
	if query != "" {
		startTime = time.Now()
	}

	attrs := make([]attribute.KeyValue, 0, len(t.attrs)+1)
	attrs = append(attrs, t.attrs...)
	if query != "" {
		attrs = append(attrs, semconv.DBStatementKey.String(t.formatQuery(query)))
	}

	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...))
	err := fn(ctx, span)
	span.End()

	if query != "" {
		t.queryHistogram.Record(ctx, time.Since(startTime).Milliseconds(), t.attrs...)
	}

	if !span.IsRecording() {
		return err
	}

	switch err {
	case nil,
		driver.ErrSkip,
		io.EOF, // end of rows iterator
		sql.ErrNoRows:
		// ignore
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

type Option func(c *config)

// WithTracerProvider configures a tracer provider that is used to create a tracer.
func WithTracerProvider(tracerProvider trace.TracerProvider) Option {
	return func(c *config) {
		c.tracerProvider = tracerProvider
	}
}

// WithAttributes configures attributes that are used to create a span.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(c *config) {
		c.attrs = append(c.attrs, attrs...)
	}
}

// WithDBSystem configures a db.system attribute. You should prefer using
// WithAttributes and semconv, for example, `otelsql.WithAttributes(semconv.DBSystemSqlite)`.
func WithDBSystem(system string) Option {
	return func(c *config) {
		c.attrs = append(c.attrs, semconv.DBSystemKey.String(system))
	}
}

// WithDBName configures a db.name attribute.
func WithDBName(name string) Option {
	return func(c *config) {
		c.attrs = append(c.attrs, semconv.DBNameKey.String(name))
	}
}

// WithMeterProvider configures a metric.Meter used to create instruments.
func WithMeterProvider(meterProvider metric.MeterProvider) Option {
	return func(c *config) {
		c.meterProvider = meterProvider
	}
}

// WithQueryFormatter configures a query formatter
func WithQueryFormatter(queryFormatter func(query string) string) Option {
	return func(c *config) {
		c.queryFormatter = queryFormatter
	}
}

// ReportDBStatsMetrics reports DBStats metrics using OpenTelemetry Metrics API.
func ReportDBStatsMetrics(db *sql.DB, opts ...Option) {
	cfg := newConfig(opts)

	if cfg.meter == nil {
		cfg.meter = cfg.meterProvider.Meter(instrumName)
	}

	meter := cfg.meter
	labels := cfg.attrs

	maxOpenConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_max_open",
		instrument.WithDescription("Maximum number of open connections to the database"),
	)
	openConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_open",
		instrument.WithDescription("The number of established connections both in use and idle"),
	)
	inUseConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_in_use",
		instrument.WithDescription("The number of connections currently in use"),
	)
	idleConns, _ := meter.Int64ObservableGauge(
		"go.sql.connections_idle",
		instrument.WithDescription("The number of idle connections"),
	)
	connsWaitCount, _ := meter.Int64ObservableCounter(
		"go.sql.connections_wait_count",
		instrument.WithDescription("The total number of connections waited for"),
	)
	connsWaitDuration, _ := meter.Int64ObservableCounter(
		"go.sql.connections_wait_duration",
		instrument.WithDescription("The total time blocked waiting for a new connection"),
		instrument.WithUnit("nanoseconds"),
	)
	connsClosedMaxIdle, _ := meter.Int64ObservableCounter(
		"go.sql.connections_closed_max_idle",
		instrument.WithDescription("The total number of connections closed due to SetMaxIdleConns"),
	)
	connsClosedMaxIdleTime, _ := meter.Int64ObservableCounter(
		"go.sql.connections_closed_max_idle_time",
		instrument.WithDescription("The total number of connections closed due to SetConnMaxIdleTime"),
	)
	connsClosedMaxLifetime, _ := meter.Int64ObservableCounter(
		"go.sql.connections_closed_max_lifetime",
		instrument.WithDescription("The total number of connections closed due to SetConnMaxLifetime"),
	)

	if _, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats := db.Stats()

			o.ObserveInt64(maxOpenConns, int64(stats.MaxOpenConnections), labels...)

			o.ObserveInt64(openConns, int64(stats.OpenConnections), labels...)
			o.ObserveInt64(inUseConns, int64(stats.InUse), labels...)
			o.ObserveInt64(idleConns, int64(stats.Idle), labels...)

			o.ObserveInt64(connsWaitCount, stats.WaitCount, labels...)
			o.ObserveInt64(connsWaitDuration, int64(stats.WaitDuration), labels...)
			o.ObserveInt64(connsClosedMaxIdle, stats.MaxIdleClosed, labels...)
			o.ObserveInt64(connsClosedMaxIdleTime, stats.MaxIdleTimeClosed, labels...)
			o.ObserveInt64(connsClosedMaxLifetime, stats.MaxLifetimeClosed, labels...)

			return nil
		},
		maxOpenConns,
		openConns,
		inUseConns,
		idleConns,
		connsWaitCount,
		connsWaitDuration,
		connsClosedMaxIdle,
		connsClosedMaxIdleTime,
		connsClosedMaxLifetime,
	); err != nil {
		panic(err)
	}
}
