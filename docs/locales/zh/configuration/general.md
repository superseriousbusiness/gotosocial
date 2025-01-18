# 基础配置

GoToSocial 的基础配置，包括域名、端口、绑定地址和传输协议等基本内容。

这里*真正*需要设置的只有 `host`，也就是你实例可以访问的域名，可能还需要设置 `port`。

## 设置

```yaml
###########################
##### 通用配置 ######
###########################

# 字符串。应用程序使用的日志级别，必须是小写。
# 选项: ["trace","debug","info","warn","error","fatal"]
# 默认: "info"
log-level: "info"

# 布尔值。当日志级别设置为 debug 或 trace 时记录数据库查询。
# 这一设置会产生详细的日志，因此最好在你尝试定位问题时才启用。
# 选项: [true, false]
# 默认: false
log-db-queries: false

# 布尔值。在日志行中包含客户端 IP。
# 选项: [true, false]
# 默认: true
log-client-ip: true

# 字符串。日志行中时间戳的格式。
# 如果设置为空字符串，则时间戳将完全从日志中省略。
#
# 该格式必须符合 Go 的 time.Layout 规定，
# 详见 https://pkg.go.dev/time#pkg-constants。
#
# 示例: ["2006-01-02T15:04:05.000Z07:00", ""]
# 默认: "02/01/2006 15:04:05.000"
log-timestamp-format: "02/01/2006 15:04:05.000"

# 字符串。内部使用的应用程序名称。
# 示例: ["My Application","gotosocial"]
# 默认: "gotosocial"
application-name: "gotosocial"

# 字符串。在首页显示的用户。如果没有设置用户，将显示默认的首页。
# 示例: "admin"
# 默认: ""
landing-page-user: ""

# 字符串。可以访问到本实例的主机名。默认值为用于本地测试的 localhost，
# 但在实际运行时你*绝对*应该更改此设置，否则你的服务器将无法正常工作。
# 在你的实例已经运行过一次后，请不要更改此项，否则会导致问题！
# 示例: ["gts.example.org","some.server.com"]
# 默认: "localhost"
host: "localhost"

# 字符串。在交换账户信息时使用的域名。当你希望服务器位于
# "gts.example.org"，但希望账户域名为 "example.org" 时，这会更好看，
# 或更加简短易记。
#
# 为使此设置正常工作，你需要将 "example.org/.well-known/webfinger" 的请求
# 重定向到 "gts.example.org/.well-known/webfinger"，以便 GtS 正常处理它们。
#
# 你还应该以同样的方式重定向 "example.org/.well-known/nodeinfo" 的请求。
#
# 你还应该以同样的方式重定向 "example.org/.well-known/host-meta" 的请求。
# 这个端点被许多客户端用于在主机名和账户域名不同时发现 API 端点。
#
# 空字符串（即，未设置）表示将使用 'host' 的相同值。
#
# 在你的服务器已经运行过一次后请不要更改此项，否则会导致问题！
#
# 在更改此设置前，请阅读安装指南的相应部分：
# https://docs.gotosocial.org/zh-cn/latest/advanced/host-account-domain/
#
# 示例: ["example.org","server.com"]
# 默认: ""
account-domain: ""

# 字符串。服务器从外界可访问的协议。
#
# 仅在本地测试时，才需将其更改为 HTTP！在 99.99% 的情况下你不应该更改此项！
#
# 这应该是你的服务器实际可以访问的 URI 的协议部分。
# 因此，即使你在处理 SSL 证书的反向代理之后运行 GoToSocial，
# 而不是使用内置的 letsencrypt，它仍然应该是 https，而不是 http。
#
# 再次强调，仅在本地测试时才需将其更改为 HTTP！如果你将其设置为 `http`，启动实例，
# 然后再更改为 `https`，你的实例上已有的用户的 URI 生成过程将被破坏。在 100% 知道自己在做什么时才更改此设置。
#
# 选项: ["http","https"]
# 默认: "https"
protocol: "https"

# 字符串。GoToSocial 服务器绑定的地址。
# 可以是 IPv4 地址或 IPv6 地址（用方括号括起来），或者是主机名。
# 默认值为绑定到所有接口，使服务器可以被其他机器访问。在大多数场景中无需更改此项。
# 如果你在与代理同一台机器上使用反向代理设置 GoToSocial，
# 建议将其设置为 "localhost" 或等效值，以防止代理被绕过。
# 示例: ["0.0.0.0", "172.128.0.16", "localhost", "[::]", "[2001:db8::fed1]"]
# 默认: "0.0.0.0"
bind-address: "0.0.0.0"

# 整数。GoToSocial 网页服务器和 API 的监听端口。如果你在反向代理和/或 Docker 容器中运行，
# 请将其设置为任意值（或保留默认值），并确保正确转发。
# 如果你启用了内建 letsencrypt 并在本机直接运行 GoToSocial，
# 可能希望将其设置为 443（标准 https 端口），除非你有其他服务正在使用该端口。
# 此项*不得*与下面指定的 letsencrypt 端口相同，除非禁用 letsencrypt。
# 示例: [443, 6666, 8080]
# 默认: 8080
port: 8080

# 字符串数组。用于通过反向代理确定真实客户端 IP 的受信任代理的 CIDR 或 IP 地址。
# 如果你的实例在 Docker 容器中运行，且位于 Traefik 或 Nginx 后，请添加你的 Docker 网络的子网，
# 或 Docker 网络的网关，和/或反向代理的地址（如果不是运行在本机上）。
# 示例: ["127.0.0.1/32", "172.20.0.1"]
# 默认: ["127.0.0.1/32", "::1"] (本地主机 ipv4 + ipv6)
trusted-proxies:
  - "127.0.0.1/32"
  - "::1"
```