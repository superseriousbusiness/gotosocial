# 健康检查

GoToSocial 提供了两个健康检查 HTTP 端点：`/readyz` 和 `/livez`。

这些端点可以用来检查 GoToSocial 是否可访问，并能够进行简单的数据库查询。

`/livez` 会始终返回 200 OK 响应且无内容，支持 GET 和 HEAD 请求。这用于检查 GoToSocial 服务是否存活。

如果 GoToSocial 能够对配置的数据库后台执行一个非常简单的 SELECT 查询，`/readyz` 会在 GET 和 HEAD 请求下返回 200 OK 响应且无内容。如果执行 SELECT 时发生错误，错误会被记录，并返回 500 Internal Server Error，但无内容。

你可以使用上述端点在容器运行时/编排系统中实现健康检查。

例如，在 Docker 设置中，你可以在 docker-compose.yaml 中添加以下内容：

```yaml
healthcheck:
  test: wget --no-verbose --tries=1 --spider http://localhost:8080/readyz || exit 1
  interval: 120s
  retries: 5
  start_period: 30s
  timeout: 10s
```

上述健康检查将在 30 秒后开始，每两分钟检查一次服务是否可用，通过对 `/readyz` 进行 HEAD 请求。如果检查连续失败五次，服务将被标记为不健康。你可以在使用的编排系统中利用此功能强制重启容器。

!!! warning
    在慢速硬件上进行数据库迁移时，迁移可能会超过上述健康检查所允许的 10 分钟。

    在这样的系统上，你可能需要增加健康检查的间隔或重试次数，以确保不会在迁移中途停止 GoToSocial（这会很糟糕！）。

!!! tip
    尽管健康检查端点不透露任何敏感信息，并且只运行非常简单的查询，你可能希望避免将它们暴露给外部世界。你可以在 nginx 中通过在 `server` 段中添加以下代码片段来实现：

    ```nginx
    location /livez {
      return 404;
    }
    location /readyz {
      return 404;
    }
    ```

    这样会导致 nginx 在请求传递给 GoToSocial 之前拦截这些请求，并直接返回 404 Not Found。

参考资料：

- [Dockerfile 参考](https://docs.docker.com/reference/dockerfile/#healthcheck)
- [Compose 文件参考](https://docs.docker.com/compose/compose-file/compose-file-v3/#healthcheck)
