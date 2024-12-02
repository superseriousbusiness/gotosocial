# 可观测性

这些设置允许你调整和配置某些与可观测性相关的行为。

## 指标

在启用指标之前，[请阅读指南](../advanced/metrics.md)，并确保你已为设置采取适当的安全措施。

## 设置

```yaml
##################################
#####     可观测性设置     #####
##################################

# 字符串。用于提取请求或跟踪ID的请求头名称。通常由负载均衡器或代理设置。
# 默认值: "X-Request-Id"
request-id-header: "X-Request-Id"

# 布尔值。启用基于OpenTelemetry的跟踪支持。
# 默认值: false
tracing-enabled: false

# 字符串。设置跟踪系统的传输协议。可以是 "grpc" 表示OTLP gRPC，或 "http" 表示OTLP HTTP。
# 选项: ["grpc", "http"]
# 默认值: "grpc"
tracing-transport: "grpc"

# 字符串。跟踪收集器的端点。使用gRPC或HTTP传输时，应提供不含协议方案的地址/端口组合。
# 示例: ["localhost:4317"]
# 默认值: ""
tracing-endpoint: ""

# 布尔值。禁用gRPC和HTTP传输协议的TLS。
# 默认值: false
tracing-insecure-transport: false

# 布尔值。启用基于OpenTelemetry的指标支持。
# 默认值: false
metrics-enabled: false

# 布尔值。为Prometheus指标端点启用HTTP基本认证。
# 默认值: false
metrics-auth-enabled: false

# 字符串。Prometheus指标端点的用户名。
# 默认值: ""
metrics-auth-username: ""

# 字符串。Prometheus指标端点的密码。
# 默认值: ""
metrics-auth-password: ""
```
