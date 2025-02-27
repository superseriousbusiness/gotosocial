# 可信代理

为了正确执行[速率限制](../api/ratelimiting.md)，GoToSocial 依赖于“可信代理”的概念，以准确确定访问你的实例的客户端的 IP 地址。

“可信代理”是一个中间网络跳转层，GoToSocial 可以配置为信任由该代理层提供的正确的客户端 IP 地址。

例如，如果你使用 Docker + Nginx 的反向代理配置中运行，那么 Nginx 的 Docker 网络地址应该被配置为可信代理，因为从广域互联网传入的所有流量将通过 Nginx 进入 GoToSocial。

如果没有正确设置 `trusted-proxies`， GoToSocial 将看到所有的入站客户端的 IP 地址都是同一个地址，这会导致速率限制的问题，因为 GoToSocial 使用客户端 IP 地址来执行速率限制。

## 总结：如何正确设置 `trusted-proxies`

如果你的 `trusted-proxies` 设置没有正确配置，你可能会在实例的网页视图中看到以下警告（v0.18.0及以上版本）：

> 警告！此实例的配置中 trusted-proxies 的设置似乎不正确。这可能导致速率限制问题，进而导致联合问题。
>
> 如果你是实例管理员，你应该通过将 `SUGGESTED_IP_RANGE` 添加到你的 trusted-proxies 来修复此问题。

要解决这个问题，可以复制消息中的IP范围，并编辑你的 `config.yaml` 文件，将IP范围添加到你的 `trusted-proxies` 中。

!!! tip "即使你没有看到上述警告，你也可能会遇到速率限制！"
    如果你使用的是低于 v0.18.0 版本的 GoToSocial，或者你在 Cloudflare（不推荐） 这样的 CDN 之后运行，你将不会看到警告消息。相反，你会在 GoToSocial 日志中看到所有客户端的 IP 都是同一个地址。在这种情况下,可以将重复出现的客户端IP值作为`SUGGESTED_IP_RANGE`。

在下面例子中，我们假定`SUGGESTED_IP_RANGE`为`172.17.0.1/16`（默认的Docker桥接网络子网）。

修改之前（默认配置）:

```yaml
trusted-proxies:
  - "127.0.0.1/32"
  - "::1"
```

修改之后（新配置）:

```yaml
trusted-proxies:
  - "172.17.0.1/16"
  - "127.0.0.1/32"
  - "::1"
```

如果你使用[环境变量](../configuration/index.md#环境变量)来配置你的实例，可以通过设置环境变量`GTS_TRUSTED_PROXIES`为以逗号分隔的IP范围列表来配置`trusted-proxies`，如下所示:

```env
GTS_TRUSTED_PROXIES="172.17.0.1/16,127.0.0.1/32,::1"
```

如果你使用 docker compose，你的 docker-compose.yaml 文件在更改后应如下所示（注意 yaml 使用 `:` 而不是 `=`）:

```yaml
################################
# 其他配置内容 #
################################
    environment:
      ############################
      # 其他环境变量 #
      ############################
      ## 对于反向代理设置:
      GTS_TRUSTED_PROXIES: "172.17.0.1/16,127.0.0.1/32,::1"
################################
# 其他配置内容 #
################################
```

一旦你完成了必要的配置更改，**重启你的实例**并刷新主页。

如果消息消失，则问题已解决！

如果你仍然看到警告消息，但显示了一个不同的建议添加到`trusted-proxies`的 IP 范围，那么重复上述步骤，在你的配置中添加新的建议 IP 范围。

!!! tip "Cloudflare 的 IP 地址列表"
    如果你在 GoToSocial 实例前面使用 CDN/代理，例如 Cloudflare （不推荐），那么你可能需要将一个或多个 Cloudflare IP 地址添加到你的 `trusted-proxies` 中，以便速率限制正常工作。你可以在这里找到Cloudflare 的 IP 地址列表： https://www.cloudflare.com/ips/

## 我可能无法正确配置 `trusted-proxies`，可以直接禁用警告吗?

在某些情况下,很难实际正确配置 `trusted-proxies` 来检测入站请求的真实客户端 IP，或者确保真实客户端 IP 是准确、但是仍显示为在私有网络内的。

例如，如果你在家用网络上运行 GoToSocial，且实例位于无法注入 `X-Forwarded-For` 标头的家庭互联网路由器之后，那么建议你添加到 `trusted-proxies` 的条目看起来会像 `192.168.x.x`，但将其添加到 `trusted-proxies` 后问题依然无法解决。

另一个例子是：你在家庭网络上运行 GoToSocial，GoToSocial 连接到家庭网络的路由器，并且你从同样在你家庭网络设备（比如笔记本或手机）上访问 Web 前端。在这种情况下，你的路由器可能会直接将你发送到你的 GoToSocial 实例，且你的请求不会离开家用网络，因此 GtS 将正确地认为*你的*客户端 IP 地址是一个私人网络地址，但*其他*从更广泛的互联网传入的请求将显示其真实的远程客户端 IP 地址。在这种情况下，`trusted-proxies` 的警告实际上不适用。

如果你已尝试编辑 `trusted-proxies` 设置，但仍看到警告，可能上面的一个例子适用于你。你可以通过以下两种方式之一继续:

### 为家庭网络添加速率限制例外（推荐）

如果 `trusted-proxies` 警告中的建议 IP 范围看起来像 `192.168.x.x`，但你在 GoToSocial 日志中仍看到其他客户端 IP 不以 `192.168` 开头，那么可以尝试只为家庭网络上的设备添加速率限制例外，同时对外部 IP 地址保持速率限制。

例如，如果你的建议是类似 `192.168.1.128/32`，那么将 `/32` 换为 `/24`，以便使范围覆盖 `192.168.1.0` -> `192.168.1.255`，并将其添加到 `config.yaml` 文件中的 `advanced-rate-limit-exceptions` 设置中。

默认设置（修改前）：

```yaml
advanced-rate-limit-exceptions: []
```

设置修改后：

```yaml
advanced-rate-limit-exceptions:
  - "192.168.1.128/24"
```

如果你使用[环境变量](../configuration/index.md#环境变量)来配置实例，可以将环境变量 `GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS` 设为以逗号分隔的 IP 范围列表，来配置 `advanced-rate-limit-exceptions`，如下所示:

```env
GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS="192.168.1.128/24"
```

如果使用 docker compose，修改后的 docker-compose.yaml 文件应如下所示(注意 yaml 使用 `: ` 而不是 `=`)：

```yaml
################################
# 其他配置内容 #
################################
    environment:
      ############################
      # 其他环境变量 #
      ############################
      GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS: "192.168.1.128/24"
################################
# 其他配置内容 #
################################
```

完成必要的配置更改后，**重启你的实例**并刷新主页。

### 完全关闭速率限制(最后手段)

如果其他方法无效，你可以完全禁用速率限制，这也会禁用 `trusted-proxies` 检查和警告。

!!! warning "警告"
    完全关闭速率限制应被视为最后的手段，因为速率限制有助于保护你的实例免受骚扰信息和爬虫攻击。

要关闭速率限制，请在 `config.yaml` 中将 `advanced-rate-limit-requests` 设置为 0。

默认配置前:

```yaml
advanced-rate-limit-requests: 300
```

设置后:

```yaml
advanced-rate-limit-requests: 0
```

如果你使用[环境变量](../configuration/index.md#环境变量)来配置实例，可以通过将环境变量 `GTS_ADVANCED_RATE_LIMIT_REQUESTS` 设置为 0，来配置 `advanced-rate-limit-requests`，如下所示：

```env
GTS_ADVANCED_RATE_LIMIT_REQUESTS="0"
```

如果使用 docker compose，改变后的 docker-compose.yaml 文件应如下所示(注意 yaml 使用 `: ` 而不是 `=`)：

```yaml
################################
# 其他配置内容 #
################################
    environment:
      ############################
      # 其他环境变量 #
      ############################
      GTS_ADVANCED_RATE_LIMIT_REQUESTS: "0"
################################
# 其他配置内容 #
################################
```

完成必要的配置更改后，**重启你的实例**并刷新主页。
