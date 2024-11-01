# 追踪

GoToSocial 内置了基于 [OpenTelemetry][otel] 的追踪功能。虽然并没有贯穿每个函数，但我们的 HTTP 处理程序和数据库库会创建跨度。在 [可观测性配置参考][obs] 中解释了如何配置追踪。

为了接收这些追踪，你需要一些工具来摄取并可视化它们。有很多选项，包括自托管和商业选项。

我们提供了一个示例，说明如何使用 [Grafana Tempo][tempo] 来抓取数据跨度，并使用 [Grafana][grafana] 来检索它们。请注意，我们提供的配置不适合生产环境。可以安全地用于本地开发，并可为设置你自己的追踪基础设施提供一个良好的起点。

你需要获取 [`example/tracing`][ext] 中的文件。获取这些文件后，你可以运行 `docker-compose up -d` 来启动 Tempo 和 Grafana。在两个服务运行后，可以将以下内容添加到 GoToSocial 配置中，并重新启动你的实例：

```yaml
tracing-enabled: true
tracing-transport: "grpc"
tracing-endpoint: "localhost:4317"
tracing-insecure-transport: true
```

[otel]: https://opentelemetry.io/
[obs]: ../configuration/observability.md
[tempo]: https://grafana.com/oss/tempo/
[grafana]: https://grafana.com/oss/grafana/
[ext]: https://github.com/superseriousbusiness/gotosocial/tree/main/example/tracing

## 查询和可视化追踪

在对你的实例执行几个查询后，你可以在 Grafana 中找到它们。你可以使用 Explore 选项卡并选择 Tempo 作为数据源。由于我们的 Grafana 示例配置启用了 [TraceQL][traceql]，Explore 选项卡将默认选择 TraceQL 查询类型。你可以改为选择“搜索”，并在“GoToSocial”服务名称下找到所有 GoToSocial 发出的追踪。

使用 TraceQL 时，一个简单的查询来查找与 `/api/v1/instance` 请求相关的所有追踪可以这样写：

```
{.http.route = "/api/v1/instance"}
```

如果你想查看所有 GoToSocial 追踪，可以运行：

```
{.service.name = "GoToSocial"}
```

选择一个追踪后，将打开第二个面板，显示对应数据跨度的可视化视图。你可以从那里深入浏览，通过点击每个子跨度查看其执行的操作。

![Grafana 显示 /api/v1/instance 端点的追踪](../public/tracing.png)

[traceql]: https://grafana.com/docs/tempo/latest/traceql/
