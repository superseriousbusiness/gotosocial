package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = (*EndpointMetrics)(nil)

func FederationBuckets() []float64 {
	return prometheus.ExponentialBucketsRange(
		(50 * time.Millisecond).Seconds(),
		(15 * time.Second).Seconds(),
		15,
	)
}

func DefaultBuckets() []float64 {
	return prometheus.ExponentialBucketsRange(
		(10 * time.Millisecond).Seconds(),
		(5 * time.Second).Seconds(),
		10,
	)
}

type EndpointMetrics struct {
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewEndpointMetrics(
	subsystem string,
	buckets []float64,
) *EndpointMetrics {
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "http",
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "The total amonut of requests received",
		},
		[]string{"method", "path", "code"},
	)
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "http",
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "A histogram of request durations",
			Buckets:   buckets,
		},
		[]string{"method", "path", "code"},
	)

	return &EndpointMetrics{
		requestCount:    requestCount,
		requestDuration: requestDuration,
	}
}

func (a *EndpointMetrics) Inc(method, path string, code int) {
	a.requestCount.With(
		prometheus.Labels{
			"method": method,
			"path":   path,
			"code":   GroupForStatusCode(code),
		},
	).Inc()
}

func (a *EndpointMetrics) Observe(method, path string, code int, duration time.Duration) {
	a.requestDuration.With(
		prometheus.Labels{
			"method": method,
			"path":   path,
			"code":   GroupForStatusCode(code),
		},
	).Observe(float64(duration) / float64(time.Second))
}

func (a *EndpointMetrics) Describe(in chan<- *prometheus.Desc) {
	a.requestCount.Describe(in)
	a.requestDuration.Describe(in)
}

func (a *EndpointMetrics) Collect(in chan<- prometheus.Metric) {
	a.requestCount.Collect(in)
	a.requestDuration.Collect(in)
}

func GroupForStatusCode(code int) string {
	switch {
	case 100 <= code && code <= 199:
		return "1xx"
	case 200 <= code && code <= 299:
		return "2xx"
	case 300 <= code && code <= 399:
		return "3xx"
	case 400 <= code && code <= 499:
		return "4xx"
	case 500 <= code && code <= 599:
		return "5xx"
	default:
		return "0xx"
	}
}
