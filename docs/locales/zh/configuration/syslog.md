# Syslog

GoToSocial 可以将日志镜像到 [syslog](https://en.wikipedia.org/wiki/Syslog)，支持通过 udp/tcp 协议传输日志，或者直接连接到本地 syslog（例如，`/var/log/syslog`）。

如果你希望通过守护进程管理 GtS 并不想自行处理日志轮转等工作，使用此功能是非常有用的，因为它依赖于经过验证的实现。

在 syslog 中的日志看起来会像这样：

```text
Dec 12 17:44:03 dilettante ./gotosocial[246860]: time=2021-12-12T17:44:03+01:00 level=info msg=connected to SQLITE database
Dec 12 17:44:03 dilettante ./gotosocial[246860]: time=2021-12-12T17:44:03+01:00 level=info msg=there are no new migrations to run func=doMigration
```

## 设置

```yaml
#########################
##### SYSLOG CONFIG #####
#########################

# 额外的 syslog 日志钩子的配置。请参阅 https://en.wikipedia.org/wiki/Syslog，
# 和 https://github.com/sirupsen/logrus/tree/master/hooks/syslog。
#
# 当需要通过守护进程管理 GtS 并将日志发送到特定位置时(无论是发送到本地位置还是 syslog 服务器)，这些设置都很有用。
# 大多数用户不需要修改这些设置。

# 布尔值。启用 syslog 日志钩子。日志将被镜像到配置的目标。
# 选项: [true, false]
# 默认值: false
syslog-enabled: false

# 字符串。指定将日志发送到 syslog 时使用的协议。留空以连接到本地 syslog。
# 选项: ["udp", "tcp", ""]
# 默认值: "udp"
syslog-protocol: "udp"

# 字符串。发送 syslog 日志的目标地址和端口。留空以连接到本地 syslog。
# 默认值: "localhost:514"
syslog-address: "localhost:514"
```
