# 反向代理

GoToSocial 可以直接暴露到互联网上。不过，许多人更愿意使用反向代理来处理外部连接。这也可以使你对 TLS 配置有更大的控制权，并启用一些更复杂的场景，比如资源缓存。

## 一般步骤

要使用反向代理，通常需要做以下几件事：

* 配置某种方式获取主机域名的 TLS 证书
* 将 GoToSocial 绑定到一个本地 IP 而不是公网 IP，并使用非特权端口。调整 `bind-address` 和 `port` 配置选项
* 如果你使用了 Lets Encrypt，在 GoToSocial 中禁用它。将 `letsencrypt-enabled` 设置为 `false`
* 配置反向代理以处理 TLS 并将请求代理到 GoToSocial

!!! warning
    不要更改 `host` 配置选项的值。这必须保持为其他实例在互联网上看到的实际域名。相反，改变 `bind-address` 并更新 `port` 和 `trusted-proxies`。

### 容器

当你使用我们的[Docker Compose 示例指南](../installation/container.md)部署 GoToSocial 时，它默认绑定到端口 `443`，假设你希望直接将其暴露到互联网上。要在反向代理后运行它，你需要更改这些设置。

在 Compose 文件中：

* 注释掉 `ports` 定义中的 `- "443:8080"` 行
* 如果你启用了 Lets Encrypt 支持：
  * 注释掉 `ports` 定义中的 `- "80:80"` 行
  * 将 `GTS_LETSENCRYPT_ENABLED` 设置回 `"false"` 或注释掉
* 改为取消注释 `- "127.0.0.1:8080:8080"` 行

这将导致 Docker 仅在 `127.0.0.1` 的端口 `8080` 上转发连接到容器，有效地将其与外界隔离。你现在可以指示反向代理将请求发送到那里。

## 指南

我们为以下服务器提供了指南：

* [nginx](nginx.md)
* [Apache httpd](apache-httpd.md)
* [Caddy 2](caddy.md)

## WebSockets

使用反向代理时，必须特别注意允许 WebSockets 正常工作。因为许多客户端应用程序使用 WebSockets 来流式传输你的时间线。WebSockets 不用于联合。

请确保阅读 [WebSocket](websocket.md) 文档，并相应地配置你的反向代理。
