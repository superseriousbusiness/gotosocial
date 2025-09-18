# Observability and Metrics

These settings let you tune and configure certain observability related behaviours.

GoToSocial uses OpenTelemetry. The metrics and trace exporters can be configured using the standard OpenTelemetry SDK environment variables. For a full reference, see [the OpenTelemetry docs](https://opentelemetry.io/docs/languages/sdk-configuration/).

## Metrics

Before enabling metrics, [read the guide](../advanced/metrics.md) and ensure you've taken the appropriate security measures for your setup.

If you want to expose metrics with (basic) authentication, you'll need to do this with a reverse proxy.

For more information and examples, see the [GtS metrics documentation](https://docs.gotosocial.org/en/latest/advanced/metrics/). 

## Settings

```yaml
##############################################
##### OBSERVABILITY AND METRICS SETTINGS #####
##############################################

# String. Header name to use to extract a request or
# trace ID from. Typically set by a loadbalancer or proxy.
#
# Default: "X-Request-Id"
request-id-header: "X-Request-Id"

# Bool. Enable OpenTelemetry based tracing support.
#
# When enabling tracing, you must also configure a traces
# exporter using the OTEL environment variable documented here:
#
# https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter
#
# Default: false
tracing-enabled: false

# Bool. Enable OpenTelemetry based metrics support.
#
# To expose Prometheus metrics, you must configure a metrics producer and
# a metrics exporter, using the OTEL environment variables documented here:
#
# https://pkg.go.dev/go.opentelemetry.io/contrib/exporters/autoexport#NewMetricReader
#
# Typically, you will want to set the following environment variables
# (take note of the plural "producers" and singular "exporter"):
#
# - OTEL_METRICS_PRODUCERS=prometheus
# - OTEL_METRICS_EXPORTER=prometheus
#
# With these variables set, a Prometheus metrics endpoint will be exposed at
# localhost:9464/metrics. This can be further configured using the variables:
#
# - OTEL_EXPORTER_PROMETHEUS_HOST
# - OTEL_EXPORTER_PROMETHEUS_PORT
#
# For more information, see the GtS metrics documentation here:
#
# https://docs.gotosocial.org/en/latest/advanced/metrics/
#
# Default: false
metrics-enabled: false
```
