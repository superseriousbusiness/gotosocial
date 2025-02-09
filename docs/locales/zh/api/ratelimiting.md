# 请求速率限制

为减轻对你的实例的滥用和抓取，系统使用了基于 IP 的 HTTP 速率限制。

不同的端点组有单独的速率限制规则。换句话说，一个部分的 API 被速率限制了，并不意味着其他部分也会被限制。以下列表中的每个项目都有单独的速率限制规则：

- `/users/*` 和 `/emoji/*` - ActivityPub (s2s) 端点。
- `/auth/*` 和 `/oauth/*` - 登录和 OAUTH 令牌请求。
- `/fileserver/*` - 媒体附件、表情符号等。
- `/nodeinfo/*` - NodeInfo 端点。
- `/.well-known/*` - webfinger 和 nodeinfo 请求。

默认情况下，每个速率限制规则允许在 5 分钟内最多进行 300 次请求：每个客户端 IP 地址每秒 1 次请求。

每个响应将包含速率限制的当前状态，具体表现为以下头信息：

- `X-Ratelimit-Limit`: 每个时间段允许的最大请求数。
- `X-Ratelimit-Remaining`: 在剩余时间内仍然可以进行的请求数量。
- `X-Ratelimit-Reset`: 表示速率限制何时重置的 ISO8601 时间戳。

如果超过速率限制，将返回 [HTTP 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429) 错误给请求者。

## 速率限制常见问题

### 我总是超出速率限制！为什么？

如果你发现自己的速率限制在正常使用时经常被超出（对于你自己和其他请求者也是如此），这可能是因为 GoToSocial 无法通过 IP 地址区分客户端。你可以通过查看实例的日志来调查这个问题。如果（几乎）所有记录的客户端 IP 地址似乎都是相同的 IP 地址（类似于 `172.x.x.x`），那么速率限制将导致问题。

这种情况通常发生在你的服务器运行在 NAT（端口转发）中，或者在没有正确配置的 HTTP 代理之后，导致你的实例将所有传入 IP 地址视为相同的地址：即你的反向代理或网关的 IP 地址。这意味着所有传入请求*共享同一个速率限制*，而不是按 IP 正确分开。

如果你使用了 HTTP 代理，那么你的 `trusted-proxies` 配置可能不正确。详情请参阅 [可信代理](../configuration/trusted_proxies.md) 文档以了解如何解决此问题。

如果没有使用 HTTP 代理，那么很可能是由 NAT 引起的。在这种情况下，你应该完全禁用速率限制。

### 我可以配置速率限制吗？可以关闭吗？

可以！在配置中设置 `advanced-rate-limit-requests: 0`。

### 我可以将一个或多个 IP 地址排除在速率限制之外，而保持其他的限制吗？

可以！在配置中设置 `advanced-rate-limit-exceptions`。
