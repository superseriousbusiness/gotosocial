# Observability

These settings let you tune and configure certain observability related behaviours.

GoToSocial uses OpenTelemetry. The metrics and trace exporters can be configured using the standard OpenTelemetry SDK environment variables. For a full reference, see [the OpenTelemetry docs](https://opentelemetry.io/docs/languages/sdk-configuration/).

## Metrics

Before enabling metrics, [read the guide](../advanced/metrics.md) and ensure you've taken the appropriate security measures for your setup.

### Tracing

To enable tracing, set `tracing-enabled` to `true`.

Valid values for `OTEL_TRACES_EXPORTER` are [documented here](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_exporter).

### Metrics

To enable metrics, set `metrics-enabled` to `true`. By default this'll use OTLP to push metrics to an OpenTelemetry receiver.

Valid values for `OTEL_METRICS_EXPORTER` are [documented here](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_metrics_exporter)

For Prometheus, set `OTEL_METRICS_EXPORTER` to `prometheus`. This will start a **second** HTTP server, bound for `localhost:9464` with the OpenTelemetry metrics. You can change this using:
* `OTEL_EXPORTER_PROMETHEUS_HOST`
* `OTEL_EXPORTER_PROMETHEUS_PORT`

### Authentication

If you want to expose the metrics with authentication, you'll need a reverse proxy. You can configure it to proxy to `http://localhost:9464/metrics` and handle authentication in your reverse proxy however you like.

## Settings

```yaml
##################################
##### OBSERVABILITY SETTINGS #####
##################################

# String. Header name to use to extract a request or trace ID from. Typically set by a
# loadbalancer or proxy.
# Default: "X-Request-Id"
request-id-header: "X-Request-Id"

# Bool. Enable OpenTelemetry based tracing support.
# Default: false
tracing-enabled: false

# Bool. Enable OpenTelemetry based metrics support.
# Default: false
metrics-enabled: false
```