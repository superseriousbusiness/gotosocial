# Metrics

GoToSocial comes with [OpenTelemetry][otel] based metrics built-in with pull-style Prometheus exporter. Currently the following metrics are collected:
* Go performance and runtime metrics
* Gin (HTTP) metrics
* Bun (database) metrics

How to configure metrics is explained in the [Observability configuration reference][obs]. Quickstart: add the following to your GoToSocial configuration and restart your instance:

```yaml
metrics-enabled: true
metrics-auth-enabled: true
metrics-auth-username: some_username
metrics-auth-password: some_password
```

This will expose the metrics under the endpoint `/metrics`, protected with HTTP Basic Authentication.

A following is an example how to configure a job in Prometheus `scrape_configs`:

```yaml
  - job_name: gotosocial
    metrics_path: /metrics
    scheme: https
    basic_auth:
        username: some_username
        password: some_password
    static_configs:
    - targets:
      - example.org
```

[otel]: https://opentelemetry.io/
[obs]: ../configuration/observability.md