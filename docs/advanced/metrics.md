# Metrics

GoToSocial comes with [OpenTelemetry][otel] based metrics built-in with pull-style Prometheus exporter. Currently the following metrics are collected:
* Go performance and runtime metrics
* Gin (HTTP) metrics
* Bun (database) metrics

How to configure metrics is explained in the [Observability configuration reference][obs]. Quickstart: add the following to your GoToSocial configuration and restart your instance:

```yaml
metrics-enabled: true
metrics-exporter: "prometheus"
```

This will expose the metrics under the **public** endpoint `/metrics`.

[otel]: https://opentelemetry.io/
[obs]: ../configuration/observability.md