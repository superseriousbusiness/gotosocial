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

# String. Set the transport protocol for the tracing system. Can either be "grpc" 
# for OTLP gRPC, or "http" for OTLP HTTP.
# Options: ["grpc", "http"]
# Default: "grpc"
tracing-transport: "grpc"

# String. Endpoint of the trace ingester. When using the gRPC or HTTP based 
# transports, provide the endpoint as a single address/port combination without a 
# protocol scheme.
# Examples: ["localhost:4317"]
# Default: ""
tracing-endpoint: ""

# Bool. Disable TLS for the gRPC and HTTP transport protocols.
# Default: false
tracing-insecure-transport: false

# Bool. Enable OpenTelemetry based metrics support.
# Default: false
metrics-enabled: false

# Bool. Enable HTTP Basic Authentication for Prometheus metrics endpoint.
# Default: false
metrics-auth-enabled: false

# String. Username for Prometheus metrics endpoint.
# Default: ""
metrics-auth-username: ""

# String. Password for Prometheus metrics endpoint.
# Default: ""
metrics-auth-password: ""
```