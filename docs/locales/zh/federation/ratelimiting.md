# 请求限流与速率限制

GoToSocial 对 ActivityPub API 端点（收件箱、用户端点、表情符号等）应用了 HTTP 请求限流和速率限制。

这可确保外站服务器不能用虚假请求淹没 GoToSocial 实例。外站服务器在对 ActivityPub API 端点进行 GET 或 POST 请求时，应尊重 429 和 503 HTTP 状态码，并考虑 `retry-after` HTTP 响应头。

有关请求限流和速率限制行为的更多详细信息，请参阅 [限流](../api/throttling.md) 和 [速率限制](../api/ratelimiting.md) 文档。
