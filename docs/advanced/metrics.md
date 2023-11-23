# Metrics

GoToSocial comes with [OpenTelemetry][otel] based metrics. The metrics are exposed using the [Prometheus exposition format][prom] on the `/metrics` path. The configuration settings are documented in the [Observability configuration reference][obs].

Currently the following metrics are collected:

* Go performance and runtime metrics
* Gin (HTTP) metrics
* Bun (database) metrics

Metrics can be enable with the following configuration:

```yaml
metrics-enabled: true
```

Though metrics do not contain anything privacy sensitive, you may not want to allow just anyone to view and scrape operational metrics of your instance.

## Enabling basic authentication

You can enable basic authentication for the metrics endpoint. On the GoToSocial, side you'll need the following configuration:

```yaml
metrics-auth-enabled: true
metrics-auth-username: some_username
metrics-auth-password: some_password
```

You can scrape that endpoint with a Prometheus instance using the following configuration in your `scrape_configs`:

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

## Blocking external scraping

When running with a reverse proxy you can use it to block external access to metrics. You can use this approach if your Prometheus scraper runs on the same machine as your GoToSocial instance and can thus access it internally.

For example with nginx, block the `/metrics` endpoint by returning a 404:

```nginx
location /metrics {
    return 404;
}
```

[otel]: https://opentelemetry.io/
[prom]: https://prometheus.io/docs/instrumenting/exposition_formats/
[obs]: ../configuration/observability.md