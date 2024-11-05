# 指标

GoToSocial 提供了基于 [OpenTelemetry][otel] 的指标。这些指标使用 [Prometheus 暴露格式][prom]，通过 `/metrics` 路径展示。配置设置在 [可观察性配置参考][obs] 中有详细说明。

当前收集的指标包括：

* Go 性能和运行时指标
* Gin (HTTP) 指标
* Bun (数据库) 指标

可以通过以下配置启用指标：

```yaml
metrics-enabled: true
```

虽然指标不包含任何隐私敏感信息，但你可能不希望随便让任何人查看和抓取你的实例的运营指标。

## 启用基本身份验证

你可以为指标端点启用基本身份验证。在 GoToSocial 上，你需要以下配置：

```yaml
metrics-auth-enabled: true
metrics-auth-username: some_username
metrics-auth-password: some_password
```

你可以使用 Prometheus 实例通过以下 `scrape_configs` 配置抓取该端点：

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

## 屏蔽外部抓取

当使用反向代理运行时，可以利用它来屏蔽对指标的外部访问。如果你的 Prometheus 抓取器在与 GoToSocial 实例相同的机器上运行，并可以内部访问它，可以使用这种方法。

例如使用 nginx，通过返回 404 来屏蔽 `/metrics` 端点：

```nginx
location /metrics {
    return 404;
}
```

[otel]: https://opentelemetry.io/
[prom]: https://prometheus.io/docs/instrumenting/exposition_formats/
[obs]: ../configuration/observability.md