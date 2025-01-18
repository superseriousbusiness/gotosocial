# 缓存 API 响应

可以缓存某些 API 响应，以减少 GoToSocial 处理所有请求的负担。我们不建议缓存 `/api` 下请求的响应。

在使用[分域](../host-account-domain.md)部署方式时，你需要确保在主机域上配置缓存。账号域应仅发出重定向到主机域的指令，客户端会自动记住这些指令。

## 端点

### Webfinger 和 hostmeta

对 `/.well-known/webfinger` 和 `/.well-known/host-meta` 的请求可以安全地缓存。注意确保任何缓存策略都考虑到 webfinger 请求的查询参数，因为对该端点的请求形式为 `?resource=acct:@username@domain.tld`。

### 公钥

许多实现将定期请求用户的公钥，以验证收到消息的签名。这将在消息联合的过程中发生。这些密钥是长期存在的，因此可以用长时间缓存。

## 配置代码片段

=== "nginx"

请先在 nginx 中配置一个缓存区。该缓存区必须在 `http` 节内创建，而非 `server` 或 `location` 内。

```nginx
http {
    ...
    proxy_cache_path /var/cache/nginx keys_zone=gotosocial_ap_public_responses:10m inactive=1w;
}
```

这配置了一个 10MB 的缓存，其条目将在一周内未被访问时保留。

该区域命名为 `gotosocial_ap_public_responses`，你可以自行更改名称。10MB 可以容纳大量缓存键；在小实例上可以使用更小的值。

其次，我们需要更新 GoToSocial 的 nginx 配置，以便真正使用我们想要缓存的端点的缓存。

```nginx
server {
  server_name social.example.org;
  
  location ~ /.well-known/(webfinger|host-meta)$ {
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_cache gotosocial_ap_public_responses;
    proxy_cache_background_update on;
    proxy_cache_key $scheme://$host$uri$is_args$query_string;
    proxy_cache_valid 200 10m;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504 http_429;
    proxy_cache_lock on;
    add_header X-Cache-Status $upstream_cache_status;

    proxy_pass http://localhost:8080;
  }

  location ~ ^\/users\/(?:[a-z0-9_\.]+)\/main-key$ {
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_cache gotosocial_ap_public_responses;
    proxy_cache_background_update on;
    proxy_cache_key $scheme://$host$uri;
    proxy_cache_valid 200 604800s;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504 http_429;
    proxy_cache_lock on;
    add_header X-Cache-Status $upstream_cache_status;

    proxy_pass http://localhost:8080;
  }
}
```

`proxy_pass` 和 `proxy_set_header` 大致相同，但 `proxy_cache*` 条目需要一些说明：

- `proxy_cache gotosocial_ap_public_responses` 告诉 nginx 使用我们之前创建的 `gotosocial_ap_public_responses` 缓存区。如果你用的是其他名称，需要更改此值。
- `proxy_cache_background_update on` 表示 nginx 会尝试在后台刷新即将过期的缓存资源，以确保磁盘上有最新副本。
- `proxy_cache_key` 的配置确保缓存时考虑到查询字符串。所以请求 `.well-known/webfinger?acct=user1@example.org` 和 `.well-known/webfinger?acct=user2@example.org` 被视为不同请求。
- `proxy_cache_valid 200 10m;` 意味着我们只缓存来自 GTS 的 200 响应，时间为 10 分钟。你可以添加类似 `proxy_cache_valid 404 1m;` 的其他行，来缓存 404 响应 1 分钟。
- `proxy_cache_use_stale` 告诉 nginx 允许在某些情况下使用过期的缓存条目（超过 10 分钟）。
- `proxy_cache_lock on` 表示如果资源未缓存且有多个并发请求，则查询将排队，以便只有一个请求通过，其他请求则从缓存中获取答案。
- `add_header X-Cache-Status $upstream_cache_status` 将 `X-Cache-Status` 头添加到响应中，以便你可以检查是否正在缓存。你可以删除此项。

上述配置将在代理到 GoToSocial 时出错、连接到 GoToSocial 时超时、GoToSocial 返回 `5xx` 状态码或 GoToSocial 返回 429（请求过多）时提供过期响应。`updating` 值表示允许在 nginx 刷新缓存时提供过期的条目。因为我们在 `proxy_cache_path` 指令中配置了 `inactive=1w`，所以如果满足 `proxy_cache_use_stale` 中的条件，nginx 可以提供最长一周的缓存响应。
