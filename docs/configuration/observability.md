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
```
