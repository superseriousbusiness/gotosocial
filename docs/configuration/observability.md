# Observability

These settings let you tune and configure certain observability related behaviours.

## Settings

```yaml
##################################
##### OBSERVABILITY SETTINGS #####
##################################

# String. Header name to use to extract a request or trace ID from. Typically set by a
# loadbalancer or proxy.
# Default: "X-Request-Id"
request-id-header: "X-Request-Id"

# Bool. Enable OpenTelemetry based tracing support.
# Default: false
tracing-enabled: false

# String. Set the transport protocol for the tracing system. Can either be "grpc" for
# OTLP gRPC or "jaeger" for jaeger based ingesters.
# Options: ["grpc", "jaeger"]
# Default: "grpc"
tracing-transport: "grpc"

# String. Endpoint of the trace ingester. When using the gRPC based transport, the
# endpoint is usually a single address/port combination. For the jaeger transport it
# should be a fully qualified URL.
# OTLP gRPC or "jaeger" for jaeger based ingesters
# Examples: ["localhost:4317", "http://localhost:14268/api/traces"]
# Default: ""
tracing-endpoint: ""

# Bool. Disable HTTPS for the gRPC transport protocol.
# Default: false
tracing-insecure-transport: false
```
