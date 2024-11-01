# WebSocket

GoToSocial 使用安全的 [WebSocket 协议](https://en.wikipedia.org/wiki/WebSocket)（即 `wss`）来通过客户端应用程序（如 Semaphore）实现贴文和通知的实时更新。

为了使用此功能，你需要确保配置 GoToSocial 所在的代理允许 WebSocket 连接通过。

WebSocket 端点位于 `wss://example.org/api/v1/streaming`，其中 `example.org` 是你的 GoToSocial 实例的域名。

WebSocket 端点使用在[通用配置](../../configuration/general.md)的 `port` 部分中配置的相同端口。

典型的 WebSocket **请求**头，如 Pinafore 所发送的如下所示：

```text
Host: example.org
User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:99.0) Gecko/20100101 Firefox/99.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Sec-WebSocket-Version: 13
Origin: https://pinafore.social
Sec-WebSocket-Protocol: null
Sec-WebSocket-Extensions: permessage-deflate
Sec-WebSocket-Key: YWFhYWFhYm9vYmllcwo=
DNT: 1
Connection: keep-alive, Upgrade
Sec-Fetch-Dest: websocket
Sec-Fetch-Mode: websocket
Sec-Fetch-Site: cross-site
Pragma: no-cache
Cache-Control: no-cache
Upgrade: websocket
```

典型的 WebSocket **响应**头，如 GoToSocial 返回的如下所示：

```text
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: WVdGaFlXRmhZbTl2WW1sbGN3bz0K
```

无论你的设置如何，你都需要确保这些头在你的反向代理中被允许，这可能需要根据所使用的具体反向代理进行额外配置。
