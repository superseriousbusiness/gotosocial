# 出站 HTTP 代理

GoToSocial 支持配置 HTTP 代理使用的标准环境变量，用于出站请求：

* `HTTP_PROXY`
* `HTTPS_PROXY`
* `NO_PROXY`

这些环境变量的小写版本也同样被识别。在处理 https 请求时，`HTTPS_PROXY` 的优先级高于 `HTTP_PROXY`。

环境变量的值可以是完整的 URL 或 `host[:port]`，在这种情况下默认使用 "http" 协议。支持的协议包括 "http"、"https" 和 "socks5"。

## systemd

使用 systemd 运行时，可以在 `Service` 部分使用 `Environment` 选项添加必要的环境变量。

如何操作可以参考 [`systemd.exec` 手册](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#Environment)。

## 容器运行时

可以在 compose 文件的 `environment` 键下设置环境变量。你也可以在命令行中使用 `-e KEY=VALUE` 或 `--env KEY=VALUE` 传递给 Docker 或 Podman 的 `run` 命令。
