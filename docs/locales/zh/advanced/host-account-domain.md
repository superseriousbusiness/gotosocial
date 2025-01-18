# 分域部署

本指南解释了如何使用 `@me@example.org` 这样的用户名，但将 GoToSocial 实例本身运行在例如 `social.example.org` 这样的子域名的方法。这种部署布局的配置**必须**在第一次启动 GoToSocial 前完成。

!!! danger
    一旦与他人联合后就无法更改域名布局。服务器会因此产生混淆，而你需要说服每个与你联合的实例管理员修改其数据库来解决问题。同时，你还需要在本地重新生成数据库，创建一个新的实例账户和加密密钥对。

## 背景

ActivityPub 实现通过一个称为 [webfinger](https://www.rfc-editor.org/rfc/rfc7033) 的协议来发现如何将你的账户域映射到你的主机域。这种映射通常会被服务器缓存，因此在事后无法更改。

它的工作原理是请求 `https://<账户域>/.well-known/webfinger?resource=acct:@me@example.org`。此时，服务器可以返回重定向到实际的 webfinger 端点 `https://<主机域>/.well-known/webfinger?resource=acct:@me@example.org` 或直接响应。返回的 JSON 文档告知应查询的用户端点：

```json
{
  "subject": "acct:me@example.org",
  "aliases": [
    "https://social.example.org/users/me",
    "https://social.example.org/@me"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "https://social.example.org/@me"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "https://social.example.org/users/me"
    }
  ]
}
```

ActivityPub 客户端和服务器将使用 `links` 数组中 `rel` 为 `self` 和 `type` 为 `application/activity+json` 的条目来查询更多信息，比如在哪里找到 `inbox` 以进行联合消息的传递。

## 配置

你需要关注两个配置设置：

* `host`，API 运行的域名，以及客户端和服务器与实例通信时使用的域
* `account-domain`，用户账户所属的域名

为了实现引言中描述的设置，你需要相应地设置这两个配置选项：

```yaml
host: social.example.org
account-domain: example.org
```

!!! info
    `host` 必须始终是运行 GoToSocial 实例的 DNS 名称。它不影响 GoToSocial 实例绑定的 IP 地址。该地址由 `bind-address` 控制。

## 反向代理

使用[反向代理](../getting_started/reverse_proxy/index.md)时，需要确保能够处理这两个域的流量。你需要将一些端点从账户域重定向到主机域。

重定向通常用于客户端侧检测域变化。需要从账户域重定向到主机域的端点是：

* `/.well-known/webfinger`
* `/.well-known/host-meta`
* `/.well-known/nodeinfo`

!!! tip
    不要将 API 端点 `/api/...` 的请求从账户域代理或重定向到主机域。这会混淆某些客户端用来检测分域部署的启发式方法，导致登录流程中断及其他异常行为。

### nginx

为了配置重定向，你需要在账户域上进行配置。假设账户域为 `example.org`，主机域为 `social.example.org`，以下配置代码展示了如何做到这一点：

```nginx
server {
  server_name example.org;                                                      # account-domain

  location /.well-known/webfinger {
    rewrite ^.*$ https://social.example.org/.well-known/webfinger permanent;    # host
  }

  location /.well-known/host-meta {
    rewrite ^.*$ https://social.example.org/.well-known/host-meta permanent;  # host
  }

  location /.well-known/nodeinfo {
    rewrite ^.*$ https://social.example.org/.well-known/nodeinfo permanent;   # host
  }
}
```

### Traefik

如果 `example.org` 运行在 [Traefik](https://doc.traefik.io/traefik/) 上，可以使用类似以下的标签设置重定向。

```yaml
myservice:
  image: foo
  # 其他配置
  labels:
    - 'traefik.http.routers.myservice.rule=Host(`example.org`)'                                 # account-domain
    - 'traefik.http.middlewares.myservice-gts.redirectregex.permanent=true'
    - 'traefik.http.middlewares.myservice-gts.redirectregex.regex=^https://(.*)/.well-known/(webfinger|nodeinfo|host-meta)(\?.*)?'  # host
    - 'traefik.http.middlewares.myservice-gts.redirectregex.replacement=https://social.${1}/.well-known/${2}${3}'                # host
    - 'traefik.http.routers.myservice.middlewares=myservice-gts@docker'
```

### Caddy 2

确保在你的 `Caddyfile` 中在账户域上配置重定向。以下示例假设账户域为 `example.com`，主机域为 `social.example.com`。

```
example.com {                                                                    # account-domain
        redir /.well-known/host-meta* https://social.example.com{uri} permanent  # host
        redir /.well-known/webfinger* https://social.example.com{uri} permanent  # host
        redir /.well-known/nodeinfo* https://social.example.com{uri} permanent   # host
}
```
